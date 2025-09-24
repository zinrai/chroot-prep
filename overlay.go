package main

import (
	"fmt"
	"path/filepath"
	"syscall"
)

// getOverlayPaths returns the paths for upper, work, and merged directories
func getOverlayPaths(chrootDir string) (upper, work, merged string) {
	overlayDir := getOverlayDir(chrootDir)
	upper = filepath.Join(overlayDir, UpperDir)
	work = filepath.Join(overlayDir, WorkDir)
	merged = filepath.Join(overlayDir, MergedDir)
	return upper, work, merged
}

// setupOverlayDirs creates the necessary directories for overlay
func setupOverlayDirs(chrootDir string) (upper, work, merged string, err error) {
	overlayDir := getOverlayDir(chrootDir)

	// Create overlay base directory
	if err = ensureDir(overlayDir, 0755); err != nil {
		return "", "", "", fmt.Errorf("failed to create overlay directory: %w", err)
	}

	// Get paths
	upper, work, merged = getOverlayPaths(chrootDir)

	// Create subdirectories
	if err = ensureDir(upper, 0755); err != nil {
		return "", "", "", fmt.Errorf("failed to create upper directory: %w", err)
	}

	if err = ensureDir(work, 0755); err != nil {
		return "", "", "", fmt.Errorf("failed to create work directory: %w", err)
	}

	if err = ensureDir(merged, 0755); err != nil {
		return "", "", "", fmt.Errorf("failed to create merged directory: %w", err)
	}

	return upper, work, merged, nil
}

// mountOverlay mounts the overlay filesystem for the chroot
func mountOverlay(chrootDir string) error {
	// Check if already mounted
	if isOverlaySetup(chrootDir) {
		return fmt.Errorf("overlay is already set up for %s", chrootDir)
	}

	// Setup directories if they don't exist
	upper, work, merged, err := setupOverlayDirs(chrootDir)
	if err != nil {
		return err
	}

	// Validate before mounting
	err = validateOverlayRequirements(chrootDir)
	if err != nil {
		return err
	}

	// Mount overlay filesystem
	err = mountOverlayFS(chrootDir, upper, work, merged)
	if err != nil {
		return fmt.Errorf("failed to mount overlay: %w", err)
	}

	return nil
}

// umountOverlay unmounts the overlay filesystem
func umountOverlay(chrootDir string) error {
	_, _, merged := getOverlayPaths(chrootDir)

	if !isMounted(merged) {
		fmt.Printf("Overlay at %s is not mounted\n", merged)
		return nil
	}

	// Unmount merged directory
	err := umountPath(merged)
	if err != nil {
		return fmt.Errorf("failed to unmount overlay: %w", err)
	}

	return nil
}

// cleanupOverlayDirs removes overlay directories (optional, for complete cleanup)
func cleanupOverlayDirs(chrootDir string) error {
	overlayDir := getOverlayDir(chrootDir)

	// First ensure nothing is mounted
	if !isOverlaySetup(chrootDir) {
		// Nothing mounted, can safely remove
		return removeIfExists(overlayDir)
	}

	// Need to unmount first
	err := umountOverlay(chrootDir)
	if err != nil {
		return fmt.Errorf("failed to unmount before cleanup: %w", err)
	}

	// Remove the entire overlay directory
	err = removeIfExists(overlayDir)
	if err != nil {
		return fmt.Errorf("failed to remove overlay directory: %w", err)
	}

	return nil
}

// validateOverlayRequirements validates that overlay can be set up
func validateOverlayRequirements(chrootDir string) error {
	// Check if base chroot directory exists
	if !dirExists(chrootDir) {
		return fmt.Errorf("base directory %s does not exist", chrootDir)
	}

	// Get overlay paths
	upper, work, _ := getOverlayPaths(chrootDir)

	// Ensure upper and work directories exist
	if !dirExists(upper) {
		return fmt.Errorf("upper directory %s does not exist", upper)
	}

	if !dirExists(work) {
		return fmt.Errorf("work directory %s does not exist", work)
	}

	// Check that upper and work are on the same filesystem
	var statUpper, statWork syscall.Stat_t

	if err := syscall.Stat(upper, &statUpper); err != nil {
		return fmt.Errorf("failed to stat upper directory: %w", err)
	}

	if err := syscall.Stat(work, &statWork); err != nil {
		return fmt.Errorf("failed to stat work directory: %w", err)
	}

	if statUpper.Dev != statWork.Dev {
		return fmt.Errorf("upper and work directories must be on the same filesystem")
	}

	return nil
}
