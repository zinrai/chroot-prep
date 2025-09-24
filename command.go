package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Setup sets up the chroot environment
func Setup(chrootDir string, useOverlay bool) error {
	// Resolve absolute path
	absPath, err := filepath.Abs(chrootDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if useOverlay {
		return setupOverlayEnvironment(absPath)
	}
	return setupNormalEnvironment(absPath)
}

// Cleanup cleans up the chroot environment
func Cleanup(chrootDir string) error {
	// Resolve absolute path
	absPath, err := filepath.Abs(chrootDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Detect environment type
	envType := detectEnvironmentType(absPath)

	switch envType {
	case OverlayEnvironment:
		return cleanupOverlayEnvironment(absPath)
	case NormalEnvironment:
		return cleanupNormalEnvironment(absPath)
	default:
		return fmt.Errorf("unknown environment type")
	}
}

// Remove removes the chroot environment
func Remove(chrootDir string, force bool, overlayOnly bool) error {
	// Resolve absolute path
	absPath, err := filepath.Abs(chrootDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Detect environment type
	envType := detectEnvironmentType(absPath)

	// Handle overlay-only removal
	if overlayOnly {
		return handleOverlayOnlyRemoval(absPath, envType, force)
	}

	// Handle complete removal
	return handleCompleteRemoval(absPath, envType, force)
}

// handleOverlayOnlyRemoval handles the -overlay flag for remove command
func handleOverlayOnlyRemoval(absPath string, envType EnvironmentType, force bool) error {
	// This only makes sense for overlay environments
	if envType != OverlayEnvironment {
		if dirExists(absPath) {
			return fmt.Errorf("no overlay environment found at %s (this is a normal chroot or base directory)", absPath)
		}
		return fmt.Errorf("chroot directory %s does not exist", absPath)
	}

	// Cleanup first
	if err := Cleanup(absPath); err != nil && !force {
		return fmt.Errorf("failed to cleanup before removal: %w", err)
	}

	return removeOverlayOnly(absPath)
}

// handleCompleteRemoval handles the default removal behavior
func handleCompleteRemoval(absPath string, envType EnvironmentType, force bool) error {
	// Check existence based on environment type
	if envType == NormalEnvironment && !dirExists(absPath) {
		return fmt.Errorf("chroot directory %s does not exist", absPath)
	}

	if envType == OverlayEnvironment && !dirExists(absPath) && !dirExists(getOverlayDir(absPath)) {
		return fmt.Errorf("chroot directory %s does not exist", absPath)
	}

	// Try to cleanup first
	if err := Cleanup(absPath); err != nil && !force {
		return fmt.Errorf("failed to cleanup before removal: %w", err)
	}

	// Remove based on environment type
	switch envType {
	case OverlayEnvironment:
		return removeOverlayCompletely(absPath)
	case NormalEnvironment:
		return removeNormalEnvironment(absPath)
	default:
		return fmt.Errorf("unknown environment type")
	}
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

// setupOverlayEnvironment sets up an overlay chroot environment
func setupOverlayEnvironment(chrootDir string) error {
	// Validate base directory exists
	if err := validateChrootStructure(chrootDir); err != nil {
		return err
	}

	// Check if already set up
	if isOverlaySetup(chrootDir) {
		return fmt.Errorf("overlay environment is already set up at %s", chrootDir)
	}

	// Setup overlay directories
	upper, work, merged, err := setupOverlayDirs(chrootDir)
	if err != nil {
		return err
	}

	// Validate overlay requirements
	if err := validateOverlayRequirements(chrootDir); err != nil {
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

	fmt.Printf("Successfully set up overlay chroot environment\n")
	fmt.Printf("Base: %s\n", chrootDir)
	fmt.Printf("Overlay: %s\n", getOverlayDir(chrootDir))
	fmt.Printf("Use: sudo chroot %s\n", merged)
	return nil
}

// cleanupNormalEnvironment cleans up a normal chroot environment
func cleanupNormalEnvironment(chrootDir string) error {
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

// cleanupOverlayEnvironment cleans up an overlay chroot environment
func cleanupOverlayEnvironment(chrootDir string) error {
	_, _, merged := getOverlayPaths(chrootDir)

	// Unmount essential filesystems from merged directory
	if err := umountEssentialFS(merged); err != nil {
		fmt.Printf("Warning: failed to unmount essential filesystems: %v\n", err)
	}

	// Cleanup resolv.conf from merged directory
	if err := cleanupResolvConf(merged); err != nil {
		fmt.Printf("Warning: failed to cleanup resolv.conf: %v\n", err)
	}

	// Unmount overlay
	if err := umountOverlay(chrootDir); err != nil {
		return fmt.Errorf("failed to unmount overlay: %w", err)
	}

	fmt.Printf("Successfully cleaned up overlay chroot environment at %s\n", chrootDir)
	return nil
}

// removeNormalEnvironment removes a normal chroot environment completely
func removeNormalEnvironment(chrootDir string) error {
	// Ensure the directory exists
	if !dirExists(chrootDir) {
		return fmt.Errorf("chroot directory %s does not exist", chrootDir)
	}

	// Remove the entire directory
	if err := os.RemoveAll(chrootDir); err != nil {
		return fmt.Errorf("failed to remove chroot directory: %w", err)
	}

	fmt.Printf("Successfully removed chroot environment at %s\n", chrootDir)
	return nil
}

// removeOverlayOnly removes only the overlay directory (preserves base)
func removeOverlayOnly(chrootDir string) error {
	overlayDir := getOverlayDir(chrootDir)

	// Ensure overlay directory exists
	if !dirExists(overlayDir) {
		return fmt.Errorf("overlay directory %s does not exist", overlayDir)
	}

	// Remove overlay directory only
	if err := os.RemoveAll(overlayDir); err != nil {
		return fmt.Errorf("failed to remove overlay directory: %w", err)
	}

	fmt.Printf("Successfully removed overlay environment at %s\n", overlayDir)
	fmt.Printf("Base directory %s is preserved\n", chrootDir)
	return nil
}

// removeOverlayCompletely removes both base and overlay directories
func removeOverlayCompletely(chrootDir string) error {
	overlayDir := getOverlayDir(chrootDir)

	// First remove overlay directory if it exists
	if dirExists(overlayDir) {
		if err := os.RemoveAll(overlayDir); err != nil {
			return fmt.Errorf("failed to remove overlay directory: %w", err)
		}
		fmt.Printf("Removed overlay directory: %s\n", overlayDir)
	}

	// Then remove base directory if it exists
	if dirExists(chrootDir) {
		if err := os.RemoveAll(chrootDir); err != nil {
			return fmt.Errorf("failed to remove base directory: %w", err)
		}
		fmt.Printf("Removed base directory: %s\n", chrootDir)
	}

	fmt.Println("Successfully removed complete overlay environment")
	return nil
}
