package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Subcommands
	setupCmd := flag.NewFlagSet("setup", flag.ExitOnError)
	setupDir := setupCmd.String("dir", "", "Path to chroot environment (required)")
	setupOverlay := setupCmd.Bool("overlay", false, "Use OverlayFS for chroot environment")

	cleanupCmd := flag.NewFlagSet("cleanup", flag.ExitOnError)
	cleanupDir := cleanupCmd.String("dir", "", "Path to chroot environment (required)")

	removeCmd := flag.NewFlagSet("remove", flag.ExitOnError)
	removeDir := removeCmd.String("dir", "", "Path to chroot environment to remove (required)")
	removeForce := removeCmd.Bool("force", false, "Force removal even if unmount fails")
	removeOverlayOnly := removeCmd.Bool("overlay", false, "Remove only overlay directory (preserve base)")

	// Parse subcommands
	switch os.Args[1] {
	case "setup":
		if err := setupCmd.Parse(os.Args[2:]); err != nil {
			log.Fatalf("Failed to parse setup command: %v", err)
		}

		if *setupDir == "" {
			log.Fatal("Please specify chroot directory using -dir flag")
		}

		if err := Setup(*setupDir, *setupOverlay); err != nil {
			log.Fatalf("Failed to setup: %v", err)
		}

	case "cleanup":
		if err := cleanupCmd.Parse(os.Args[2:]); err != nil {
			log.Fatalf("Failed to parse cleanup command: %v", err)
		}

		if *cleanupDir == "" {
			log.Fatal("Please specify chroot directory using -dir flag")
		}

		if err := Cleanup(*cleanupDir); err != nil {
			log.Fatalf("Failed to cleanup: %v", err)
		}

	case "remove":
		if err := removeCmd.Parse(os.Args[2:]); err != nil {
			log.Fatalf("Failed to parse remove command: %v", err)
		}

		if *removeDir == "" {
			log.Fatal("Please specify chroot directory using -dir flag")
		}

		if err := Remove(*removeDir, *removeForce, *removeOverlayOnly); err != nil {
			log.Fatalf("Failed to remove: %v", err)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	const usage = `chroot-prep - Manage filesystem mounts for chroot environments

Usage:
  chroot-prep setup -dir /path/to/chroot [-overlay]
  chroot-prep cleanup -dir /path/to/chroot
  chroot-prep remove -dir /path/to/chroot [-force] [-overlay]

Commands:
  setup    Setup chroot environment with essential filesystems
  cleanup  Cleanup mounted filesystems from chroot environment
  remove   Remove chroot environment (with automatic unmounting)

Setup Options:
  -dir string    Path to chroot directory (required)
  -overlay       Use OverlayFS (base directory as read-only lower layer)

Cleanup Options:
  -dir string    Path to chroot directory (required)

Remove Options:
  -dir string    Path to chroot directory (required)
  -force         Force removal even if unmount fails
  -overlay       Remove only overlay directory (preserve base)

Examples:
  # Normal chroot setup
  sudo chroot-prep setup -dir /mnt/my-chroot

  # OverlayFS chroot setup (using /mnt/base as read-only base)
  sudo chroot-prep setup -dir /mnt/base -overlay

  # Cleanup (automatically detects environment type)
  sudo chroot-prep cleanup -dir /mnt/base

  # Remove everything (base + overlay)
  sudo chroot-prep remove -dir /mnt/base

  # Remove only overlay (preserve base)
  sudo chroot-prep remove -dir /mnt/base -overlay

Note: This program requires root privileges (sudo)`

	fmt.Println(usage)
}

func init() {
	// Ensure we're running as root
	if os.Geteuid() != 0 {
		log.Fatal("This program must be run as root (sudo)")
	}
}
