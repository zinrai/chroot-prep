package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Subcommands
	setupCmd := flag.NewFlagSet("setup", flag.ExitOnError)
	setupDir := setupCmd.String("dir", "", "Path to existing chroot environment (required)")

	cleanupCmd := flag.NewFlagSet("cleanup", flag.ExitOnError)
	cleanupDir := cleanupCmd.String("dir", "", "Path to existing chroot environment (required)")

	removeCmd := flag.NewFlagSet("remove", flag.ExitOnError)
	removeDir := removeCmd.String("dir", "", "Path to chroot environment to remove (required)")
	force := removeCmd.Bool("f", false, "Force removal even if unmount fails")

	switch os.Args[1] {
	case "setup":
		setupCmd.Parse(os.Args[2:])
		if *setupDir == "" {
			log.Fatal("Please specify chroot directory using -dir flag")
		}
		handleSetup(*setupDir)

	case "cleanup":
		cleanupCmd.Parse(os.Args[2:])
		if *cleanupDir == "" {
			log.Fatal("Please specify chroot directory using -dir flag")
		}
		handleCleanup(*cleanupDir)

	case "remove":
		removeCmd.Parse(os.Args[2:])
		if *removeDir == "" {
			log.Fatal("Please specify chroot directory using -dir flag")
		}
		handleRemove(*removeDir, *force)

	default:
		printUsage()
		os.Exit(1)
	}
}

func handleSetup(chrootDir string) {
	absPath, err := filepath.Abs(chrootDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	// Verify the directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Fatalf("Chroot directory %s does not exist", absPath)
	}

	// Verify required directories exist
	requiredDirs := []string{"dev", "proc", "sys"}
	for _, dir := range requiredDirs {
		fullPath := filepath.Join(absPath, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			log.Fatalf("Required directory %s does not exist in chroot environment", dir)
		}
	}

	// Mount the filesystems
	if err := mountEssentialFS(absPath); err != nil {
		log.Fatalf("Failed to mount filesystems: %v", err)
	}

	if err := setupResolvConf(absPath); err != nil {
		// Cleanup mounted filesystems if resolv.conf setup fails
		if umountErr := umountEssentialFS(absPath); umountErr != nil {
			log.Printf("Warning: Failed to unmount filesystems: %v", umountErr)
		}
		log.Fatalf("Failed to setup resolv.conf: %v", err)
	}

	fmt.Printf("Successfully set up chroot environment at %s\n", absPath)
}

func handleCleanup(chrootDir string) {
	absPath, err := filepath.Abs(chrootDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	if err := umountEssentialFS(absPath); err != nil {
		log.Fatalf("Failed to unmount filesystems: %v", err)
	}

	if err := cleanupResolvConf(absPath); err != nil {
		log.Printf("Warning: Failed to cleanup resolv.conf: %v", err)
	}

	fmt.Printf("Successfully cleaned up chroot environment at %s\n", absPath)
}

func handleRemove(chrootDir string, force bool) {
	absPath, err := filepath.Abs(chrootDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Fatalf("Chroot directory %s does not exist", absPath)
	}

	// Check and unmount if necessary
	mountPoints := []string{"sys", "proc", "dev"}
	mounted := false
	for _, mp := range mountPoints {
		if isMounted(filepath.Join(absPath, mp)) {
			mounted = true
			break
		}
	}

	if mounted {
		fmt.Println("Found mounted filesystems, attempting to unmount...")
		if err := umountEssentialFS(absPath); err != nil {
			if !force {
				log.Fatalf("Failed to unmount filesystems: %v\nUse -f flag to force removal", err)
			}
			fmt.Printf("Warning: Failed to unmount some filesystems: %v\n", err)
		}
	}

	// Remove the directory
	if err := os.RemoveAll(absPath); err != nil {
		log.Fatalf("Failed to remove chroot environment: %v", err)
	}

	fmt.Printf("Successfully removed chroot environment at %s\n", absPath)
}

func mountEssentialFS(chrootDir string) error {
	mounts := []struct {
		source string
		target string
		fstype string
		flags  uintptr
		data   string
	}{
		// mount -t proc none ${CHROOT_DIR}/proc
		{"none", filepath.Join(chrootDir, "proc"), "proc", 0, ""},
		// mount --bind /dev ${CHROOT_DIR}/dev
		{"/dev", filepath.Join(chrootDir, "dev"), "none", syscall.MS_BIND, ""},
		// mount --bind /sys ${CHROOT_DIR}/sys
		{"/sys", filepath.Join(chrootDir, "sys"), "none", syscall.MS_BIND, ""},
	}

	for _, m := range mounts {
		// Check if already mounted
		if isMounted(m.target) {
			fmt.Printf("%s is already mounted, skipping...\n", m.target)
			continue
		}

		// Mount the filesystem
		if err := syscall.Mount(m.source, m.target, m.fstype, m.flags, m.data); err != nil {
			return fmt.Errorf("failed to mount %s: %v", m.target, err)
		}
	}

	return nil
}

func umountEssentialFS(chrootDir string) error {
	// Unmount in reverse order to handle dependencies
	mounts := []string{
		filepath.Join(chrootDir, "sys"),
		filepath.Join(chrootDir, "proc"),
		filepath.Join(chrootDir, "dev"),
	}

	for _, mountpoint := range mounts {
		if !isMounted(mountpoint) {
			fmt.Printf("%s is not mounted, skipping...\n", mountpoint)
			continue
		}

		// Try lazy unmount if normal unmount fails
		if err := syscall.Unmount(mountpoint, 0); err != nil {
			fmt.Printf("Normal unmount failed for %s, trying lazy unmount...\n", mountpoint)
			if err := syscall.Unmount(mountpoint, syscall.MNT_DETACH); err != nil {
				return fmt.Errorf("failed to unmount %s: %v", mountpoint, err)
			}
		}
	}

	return nil
}

func isMounted(mountpoint string) bool {
	cmd := exec.Command("mountpoint", "-q", mountpoint)
	return cmd.Run() == nil
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  chroot-prep setup -dir /path/to/chroot     # Setup chroot environment")
	fmt.Println("  chroot-prep cleanup -dir /path/to/chroot   # Cleanup chroot environment")
	fmt.Println("  chroot-prep remove -dir /path/to/chroot    # Remove chroot environment")
	fmt.Println("Options for remove:")
	fmt.Println("  -f    Force removal even if unmount fails")
}

func setupResolvConf(chrootDir string) error {
	// Source and destination paths
	hostResolvConf := "/etc/resolv.conf"
	chrootResolvConf := filepath.Join(chrootDir, "etc/resolv.conf")

	// Check if source exists
	if _, err := os.Stat(hostResolvConf); os.IsNotExist(err) {
		return fmt.Errorf("host %s does not exist", hostResolvConf)
	}

	// Read source file
	content, err := os.ReadFile(hostResolvConf)
	if err != nil {
		return fmt.Errorf("failed to read %s: %v", hostResolvConf, err)
	}

	// Write to destination
	if err := os.WriteFile(chrootResolvConf, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %v", chrootResolvConf, err)
	}

	return nil
}

func cleanupResolvConf(chrootDir string) error {
	chrootResolvConf := filepath.Join(chrootDir, "etc/resolv.conf")

	// Remove resolv.conf if it exists
	if _, err := os.Stat(chrootResolvConf); err == nil {
		if err := os.Remove(chrootResolvConf); err != nil {
			return fmt.Errorf("failed to remove %s: %v", chrootResolvConf, err)
		}
	}

	return nil
}

func init() {
	// Ensure we're running as root
	if os.Geteuid() != 0 {
		log.Fatal("This program must be run as root (sudo)")
	}
}
