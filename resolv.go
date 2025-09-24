package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	hostResolvConf = "/etc/resolv.conf"
	resolvConfName = "etc/resolv.conf"
)

// setupResolvConf copies the host's resolv.conf to the chroot environment
func setupResolvConf(chrootDir string) error {
	// Check if source exists
	if !fileExists(hostResolvConf) {
		return fmt.Errorf("host %s does not exist", hostResolvConf)
	}

	// Destination path
	chrootResolvConf := filepath.Join(chrootDir, resolvConfName)

	// Ensure etc directory exists
	etcDir := filepath.Join(chrootDir, "etc")
	if err := ensureDir(etcDir, 0755); err != nil {
		return fmt.Errorf("failed to create etc directory: %w", err)
	}

	// Read source file
	content, err := os.ReadFile(hostResolvConf)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", hostResolvConf, err)
	}

	// Write to destination
	if err := os.WriteFile(chrootResolvConf, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", chrootResolvConf, err)
	}

	return nil
}

// cleanupResolvConf removes the resolv.conf from the chroot environment
func cleanupResolvConf(chrootDir string) error {
	chrootResolvConf := filepath.Join(chrootDir, resolvConfName)

	// Remove resolv.conf if it exists
	if fileExists(chrootResolvConf) {
		if err := os.Remove(chrootResolvConf); err != nil {
			return fmt.Errorf("failed to remove %s: %w", chrootResolvConf, err)
		}
	}

	return nil
}

// validateResolvConf checks if resolv.conf is properly set up
func validateResolvConf(chrootDir string) bool {
	chrootResolvConf := filepath.Join(chrootDir, resolvConfName)
	return fileExists(chrootResolvConf)
}
