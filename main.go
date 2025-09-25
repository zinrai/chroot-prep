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
	cleanupOverlay := cleanupCmd.Bool("overlay", false, "Cleanup overlay environment")

	removeCmd := flag.NewFlagSet("remove", flag.ExitOnError)
	removeDir := removeCmd.String("dir", "", "Path to chroot environment to remove (required)")
	removeForce := removeCmd.Bool("force", false, "Force removal even if unmount fails")
	removeOverlay := removeCmd.Bool("overlay", false, "Remove overlay directory")

	// Parse subcommands
	switch os.Args[1] {
	case "setup":
		if err := setupCmd.Parse(os.Args[2:]); err != nil {
			log.Fatalf("Failed to parse setup command: %v", err)
		}

		if *setupDir == "" {
			log.Fatal("Please specify chroot directory using -dir flag")
		}

		// Handle overlay with optional name
		overlayName := ""
		if *setupOverlay {
			overlayName = "overlay" // default
			args := setupCmd.Args()
			if len(args) > 0 {
				overlayName = args[0]
			}
		}

		if err := Setup(*setupDir, overlayName); err != nil {
			log.Fatalf("Failed to setup: %v", err)
		}

	case "cleanup":
		if err := cleanupCmd.Parse(os.Args[2:]); err != nil {
			log.Fatalf("Failed to parse cleanup command: %v", err)
		}

		if *cleanupDir == "" {
			log.Fatal("Please specify chroot directory using -dir flag")
		}

		// Handle overlay with optional name
		overlayName := ""
		if *cleanupOverlay {
			overlayName = "overlay" // default
			args := cleanupCmd.Args()
			if len(args) > 0 {
				overlayName = args[0]
			}
		}

		if err := Cleanup(*cleanupDir, overlayName); err != nil {
			log.Fatalf("Failed to cleanup: %v", err)
		}

	case "remove":
		if err := removeCmd.Parse(os.Args[2:]); err != nil {
			log.Fatalf("Failed to parse remove command: %v", err)
		}

		if *removeDir == "" {
			log.Fatal("Please specify chroot directory using -dir flag")
		}

		// Handle overlay with optional name
		overlayName := ""
		if *removeOverlay {
			overlayName = "overlay" // default
			args := removeCmd.Args()
			if len(args) > 0 {
				overlayName = args[0]
			}
		}

		if err := Remove(*removeDir, *removeForce, overlayName); err != nil {
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
  chroot-prep setup -dir /path/to/chroot [-overlay [name]]
  chroot-prep cleanup -dir /path/to/chroot [-overlay [name]]
  chroot-prep remove -dir /path/to/chroot [-force] [-overlay [name]]

Commands:
  setup    Setup chroot environment with essential filesystems
  cleanup  Cleanup mounted filesystems from chroot environment
  remove   Remove chroot environment (with automatic unmounting)

Setup Options:
  -dir string    Path to chroot directory (required)
  -overlay       Use OverlayFS (optionally specify name, default: 'overlay')

Cleanup Options:
  -dir string    Path to chroot directory (required)
  -overlay       Cleanup overlay (optionally specify name, default: 'overlay')

Remove Options:
  -dir string    Path to chroot directory (required)
  -force         Force removal even if unmount fails
  -overlay       Remove only overlay (optionally specify name, default: 'overlay')

Examples:
  # Normal chroot setup
  sudo chroot-prep setup -dir /mnt/my-chroot

  # OverlayFS with default name
  sudo chroot-prep setup -dir /mnt/base -overlay

  # OverlayFS with custom name
  sudo chroot-prep setup -dir /mnt/base -overlay projectA

  # Cleanup specific overlay
  sudo chroot-prep cleanup -dir /mnt/base -overlay projectA

  # Remove everything (base + all overlays)
  sudo chroot-prep remove -dir /mnt/base

  # Remove only specific overlay (preserve base)
  sudo chroot-prep remove -dir /mnt/base -overlay projectA

Note: This program requires root privileges (sudo)`

	fmt.Println(usage)
}

func init() {
	// Ensure we're running as root
	if os.Geteuid() != 0 {
		log.Fatal("This program must be run as root (sudo)")
	}
}
