package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Setup sets up the chroot environment
func Setup(chrootDir string, overlayName string) error {
	// Resolve absolute path
	absPath, err := filepath.Abs(chrootDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if overlayName != "" {
		return setupOverlayEnvironment(absPath, overlayName)
	}
	return setupNormalEnvironment(absPath)
}

// Cleanup cleans up the chroot environment
func Cleanup(chrootDir string, overlayName string) error {
	// Resolve absolute path
	absPath, err := filepath.Abs(chrootDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if overlayName != "" {
		// Cleanup specific overlay
		return cleanupOverlayEnvironment(absPath, overlayName)
	}

	// Cleanup normal chroot
	return cleanupNormalEnvironment(absPath)
}

// Remove removes the chroot environment
func Remove(chrootDir string, force bool, overlayName string) error {
	// Resolve absolute path
	absPath, err := filepath.Abs(chrootDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// If overlay name is specified, remove only that overlay
	if overlayName != "" {
		return removeSpecificOverlay(absPath, overlayName, force)
	}

	// Remove everything (base + all overlays)
	return removeAll(absPath, force)
}

// setupNormalEnvironment sets up a normal chroot environment
func setupNormalEnvironment(chrootDir string) error {
	// Validate directory exists
	if err := validateChrootStructure(chrootDir); err != nil {
		return err
	}

	// Mount essential filesystems
	if err := mountEssentialFS(chrootDir); err != nil {
		return fmt.Errorf("failed to mount filesystems: %w", err)
	}

	// Setup resolv.conf
	if err := setupResolvConf(chrootDir); err != nil {
		// Cleanup on failure
		umountEssentialFS(chrootDir)
		return fmt.Errorf("failed to setup resolv.conf: %w", err)
	}

	fmt.Printf("Successfully set up chroot environment at %s\n", chrootDir)
	return nil
}

// setupOverlayEnvironment sets up an overlay chroot environment with named overlay
func setupOverlayEnvironment(chrootDir string, overlayName string) error {
	// Validate base directory exists
	if err := validateChrootStructure(chrootDir); err != nil {
		return err
	}

	// Check if this specific overlay is already set up
	if isOverlaySetup(chrootDir, overlayName) {
		return fmt.Errorf("overlay '%s' is already set up at %s", overlayName, chrootDir)
	}

	// Setup overlay directories
	upper, work, merged, err := setupOverlayDirs(chrootDir, overlayName)
	if err != nil {
		return err
	}

	// Validate overlay requirements
	if err := validateOverlayRequirements(chrootDir, overlayName); err != nil {
		return err
	}

	// Mount overlay filesystem
	if err := mountOverlayFS(chrootDir, upper, work, merged); err != nil {
		return fmt.Errorf("failed to mount overlay: %w", err)
	}

	// Mount essential filesystems on merged directory
	if err := mountEssentialFS(merged); err != nil {
		// Cleanup on failure
		umountPath(merged)
		return fmt.Errorf("failed to mount essential filesystems: %w", err)
	}

	// Setup resolv.conf in merged directory
	if err := setupResolvConf(merged); err != nil {
		// Cleanup on failure
		umountEssentialFS(merged)
		umountPath(merged)
		return fmt.Errorf("failed to setup resolv.conf: %w", err)
	}

	overlayDir := getOverlayDir(chrootDir, overlayName)
	fmt.Printf("Successfully set up overlay chroot environment\n")
	fmt.Printf("Base: %s\n", chrootDir)
	fmt.Printf("Overlay: %s\n", overlayDir)
	fmt.Printf("Use: sudo chroot %s\n", merged)
	return nil
}

// cleanupNormalEnvironment cleans up a normal chroot environment
func cleanupNormalEnvironment(chrootDir string) error {
	// Check if it's actually a normal environment
	if !dirExists(chrootDir) {
		return fmt.Errorf("chroot directory %s does not exist", chrootDir)
	}

	// Unmount essential filesystems
	if err := umountEssentialFS(chrootDir); err != nil {
		return fmt.Errorf("failed to unmount filesystems: %w", err)
	}

	// Cleanup resolv.conf
	if err := cleanupResolvConf(chrootDir); err != nil {
		// Non-critical error, just warn
		fmt.Printf("Warning: failed to cleanup resolv.conf: %v\n", err)
	}

	fmt.Printf("Successfully cleaned up chroot environment at %s\n", chrootDir)
	return nil
}

// cleanupOverlayEnvironment cleans up a specific overlay chroot environment
func cleanupOverlayEnvironment(chrootDir string, overlayName string) error {
	_, _, merged := getOverlayPaths(chrootDir, overlayName)

	// Check if overlay exists
	overlayDir := getOverlayDir(chrootDir, overlayName)
	if !dirExists(overlayDir) {
		return fmt.Errorf("overlay '%s' does not exist at %s", overlayName, chrootDir)
	}

	// Unmount essential filesystems from merged directory
	if err := umountEssentialFS(merged); err != nil {
		fmt.Printf("Warning: failed to unmount essential filesystems: %v\n", err)
	}

	// Cleanup resolv.conf from merged directory
	if err := cleanupResolvConf(merged); err != nil {
		fmt.Printf("Warning: failed to cleanup resolv.conf: %v\n", err)
	}

	// Unmount overlay
	if err := umountOverlay(chrootDir, overlayName); err != nil {
		return fmt.Errorf("failed to unmount overlay: %w", err)
	}

	fmt.Printf("Successfully cleaned up overlay '%s' at %s\n", overlayName, chrootDir)
	return nil
}

// removeSpecificOverlay removes only a specific overlay directory
func removeSpecificOverlay(chrootDir string, overlayName string, force bool) error {
	overlayDir := getOverlayDir(chrootDir, overlayName)

	// Check if overlay exists
	if !dirExists(overlayDir) {
		return fmt.Errorf("overlay '%s' does not exist at %s", overlayName, chrootDir)
	}

	// Try to cleanup first
	if err := cleanupOverlayEnvironment(chrootDir, overlayName); err != nil && !force {
		return fmt.Errorf("failed to cleanup before removal: %w", err)
	}

	// Remove overlay directory
	if err := os.RemoveAll(overlayDir); err != nil {
		return fmt.Errorf("failed to remove overlay directory: %w", err)
	}

	fmt.Printf("Successfully removed overlay '%s'\n", overlayName)
	fmt.Printf("Base directory %s is preserved\n", chrootDir)
	return nil
}

// removeAll removes base and all overlays
func removeAll(chrootDir string, force bool) error {
	// Find and remove all overlays
	if err := removeAllOverlays(chrootDir, force); err != nil && !force {
		return err
	}

	// Remove base directory if it exists
	if err := removeBaseDirectory(chrootDir, force); err != nil {
		return err
	}

	fmt.Println("Successfully removed all environments")
	return nil
}

// removeAllOverlays finds and removes all overlay directories for a base
func removeAllOverlays(chrootDir string, force bool) error {
	parentDir := filepath.Dir(chrootDir)
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		// If we can't read the parent directory, skip overlay cleanup
		return nil
	}

	baseName := filepath.Base(chrootDir)
	prefix := baseName + "."

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if !isOverlayDirectory(entry.Name(), baseName, prefix) {
			continue
		}

		overlayName := entry.Name()[len(prefix):]
		if err := removeOverlayDirectory(chrootDir, overlayName, parentDir, entry.Name(), force); err != nil && !force {
			return err
		}
	}

	return nil
}

// isOverlayDirectory checks if a directory name matches overlay pattern
func isOverlayDirectory(name, baseName, prefix string) bool {
	return len(name) > len(baseName) && name[:len(prefix)] == prefix
}

// removeOverlayDirectory removes a single overlay directory
func removeOverlayDirectory(chrootDir, overlayName, parentDir, dirName string, force bool) error {
	// Try to cleanup first
	if err := cleanupOverlayEnvironment(chrootDir, overlayName); err != nil && !force {
		fmt.Printf("Warning: failed to cleanup overlay '%s': %v\n", overlayName, err)
	}

	// Remove the overlay directory
	overlayPath := filepath.Join(parentDir, dirName)
	if err := os.RemoveAll(overlayPath); err != nil && !force {
		fmt.Printf("Warning: failed to remove overlay '%s': %v\n", overlayName, err)
		return err
	}

	fmt.Printf("Removed overlay: %s\n", overlayName)
	return nil
}

// removeBaseDirectory removes the base chroot directory
func removeBaseDirectory(chrootDir string, force bool) error {
	if !dirExists(chrootDir) {
		return nil
	}

	// Try to cleanup as normal environment
	if err := cleanupNormalEnvironment(chrootDir); err != nil && !force {
		fmt.Printf("Warning: failed to cleanup base: %v\n", err)
	}

	// Remove base directory
	if err := os.RemoveAll(chrootDir); err != nil {
		return fmt.Errorf("failed to remove base directory: %w", err)
	}

	fmt.Printf("Removed base directory: %s\n", chrootDir)
	return nil
}
