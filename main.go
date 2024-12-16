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
	mountCmd := flag.NewFlagSet("mount", flag.ExitOnError)
	mountDir := mountCmd.String("dir", "", "Path to existing chroot environment (required)")

	umountCmd := flag.NewFlagSet("umount", flag.ExitOnError)
	umountDir := umountCmd.String("dir", "", "Path to existing chroot environment (required)")

	switch os.Args[1] {
	case "mount":
		mountCmd.Parse(os.Args[2:])
		if *mountDir == "" {
			log.Fatal("Please specify chroot directory using -dir flag")
		}
		handleMount(*mountDir)

	case "umount":
		umountCmd.Parse(os.Args[2:])
		if *umountDir == "" {
			log.Fatal("Please specify chroot directory using -dir flag")
		}
		handleUmount(*umountDir)

	default:
		printUsage()
		os.Exit(1)
	}
}

func handleMount(chrootDir string) {
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
	if err := mountFilesystems(absPath); err != nil {
		log.Fatalf("Failed to mount filesystems: %v", err)
	}

	fmt.Printf("Successfully mounted filesystems in chroot environment at %s\n", absPath)
}

func handleUmount(chrootDir string) {
	absPath, err := filepath.Abs(chrootDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	if err := umountFilesystems(absPath); err != nil {
		log.Fatalf("Failed to unmount filesystems: %v", err)
	}

	fmt.Printf("Successfully unmounted filesystems in chroot environment at %s\n", absPath)
}

func mountFilesystems(chrootDir string) error {
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

func umountFilesystems(chrootDir string) error {
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
	fmt.Println("  mountchroot mount -dir /path/to/chroot   # Mount filesystems")
	fmt.Println("  mountchroot umount -dir /path/to/chroot  # Unmount filesystems")
}

func init() {
	// Ensure we're running as root
	if os.Geteuid() != 0 {
		log.Fatal("This program must be run as root (sudo)")
	}
}
