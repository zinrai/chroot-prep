package main

import (
	"fmt"
	"path/filepath"
)

// detectEnvironmentType detects whether the environment is normal or overlay
func detectEnvironmentType(chrootDir string, overlayName string) EnvironmentType {
	overlayDir := getOverlayDir(chrootDir, overlayName)
	if dirExists(overlayDir) {
		return OverlayEnvironment
	}
	// Check if base directory exists as normal chroot
	if dirExists(chrootDir) {
		return NormalEnvironment
	}
	return NormalEnvironment // default
}

// getOverlayDir returns the overlay directory path for a given chroot directory and name
func getOverlayDir(chrootDir string, overlayName string) string {
	if overlayName == "" {
		return ""
	}
	return fmt.Sprintf("%s.%s", chrootDir, overlayName)
}

// isOverlaySetup checks if a specific overlay is already set up
func isOverlaySetup(chrootDir string, overlayName string) bool {
	overlayDir := getOverlayDir(chrootDir, overlayName)
	if overlayDir == "" {
		return false
	}
	mergedPath := filepath.Join(overlayDir, MergedDir)

	// Check if overlay directory exists and merged is mounted
	return dirExists(overlayDir) && isMounted(mergedPath)
}

// validateChrootStructure validates that the chroot directory has the required structure
func validateChrootStructure(chrootDir string) error {
	// Check if the directory exists
	if !dirExists(chrootDir) {
		return fmt.Errorf("chroot directory %s does not exist", chrootDir)
	}

	// For a minimal check, just ensure it's a directory
	// Additional validation can be added here if needed
	return nil
}

// validateOverlayStructure validates the overlay directory structure
func validateOverlayStructure(chrootDir string, overlayName string) error {
	overlayDir := getOverlayDir(chrootDir, overlayName)

	// Check if overlay directory exists
	if !dirExists(overlayDir) {
		return fmt.Errorf("overlay directory %s does not exist", overlayDir)
	}

	// Check required subdirectories
	requiredDirs := []string{UpperDir, WorkDir, MergedDir}
	for _, dir := range requiredDirs {
		fullPath := filepath.Join(overlayDir, dir)
		if !dirExists(fullPath) {
			return fmt.Errorf("required overlay directory %s does not exist", fullPath)
		}
	}

	return nil
}

// ensureChrootDirs ensures that essential directories exist in the chroot
func ensureChrootDirs(chrootDir string) error {
	// Essential directories that should exist in any chroot
	essentialDirs := []string{"dev", "proc", "sys", "etc"}

	for _, dir := range essentialDirs {
		fullPath := filepath.Join(chrootDir, dir)
		if err := ensureDir(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create essential directory %s: %w", dir, err)
		}
	}

	return nil
}
