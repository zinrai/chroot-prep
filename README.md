# chroot-prep

A command-line tool to manage filesystem mounts for chroot environments with OverlayFS support.

## Features

- Setup essential filesystems (`/dev`, `/proc`, `/sys`) into chroot environments
- Setup host's resolv.conf(5) to chroot environments
- **OverlayFS support** for layered chroot environments
- **Named overlays** for multiple independent environments from the same base
- Preserve base environments while making experimental changes
- Automatic environment type detection
- Clean removal with automatic unmounting

## Requirements

- Linux operating system (kernel 3.18+ for OverlayFS support)
- Root privileges (sudo)
- `mountpoint` command (usually part of util-linux package)

## Installation

Build from source:

```bash
$ go build
$ sudo cp chroot-prep /usr/local/bin/
```

## Prerequisites

For the examples in this documentation, we'll use a Debian trixie environment created with debootstrap:

```bash
# Install debootstrap (on Debian/Ubuntu)
$ sudo apt-get install debootstrap

# Create a Debian trixie chroot environment
$ sudo debootstrap --arch=amd64 trixie trixie-amd64 http://deb.debian.org/debian
```

## Quick Start

### Traditional chroot setup

```bash
# Setup the chroot environment
$ sudo chroot-prep setup -dir trixie-amd64

# Enter the chroot
$ sudo chroot trixie-amd64 /bin/bash

# When done, cleanup
$ sudo chroot-prep cleanup -dir trixie-amd64
```

### OverlayFS mode

```bash
# Setup with OverlayFS (default name "overlay")
$ sudo chroot-prep setup -dir trixie-amd64 -overlay

# Setup with custom overlay name
$ sudo chroot-prep setup -dir trixie-amd64 -overlay projectA

# Enter the overlay chroot
$ sudo chroot trixie-amd64.overlay/merged /bin/bash
# or for named overlay:
$ sudo chroot trixie-amd64.projectA/merged /bin/bash

# When done, cleanup specific overlay
$ sudo chroot-prep cleanup -dir trixie-amd64 -overlay
$ sudo chroot-prep cleanup -dir trixie-amd64 -overlay projectA
```

## OverlayFS Mode

OverlayFS mode creates a layered filesystem where your base chroot environment remains read-only, and all changes are written to a separate overlay layer.

### Directory Structure

When you use `-overlay`, the following structure is created:

```
./
├── trixie-amd64/                 # Base environment (read-only)
│   ├── bin/
│   ├── boot/
│   ├── dev/
│   ├── etc/
│   └── ...
├── trixie-amd64.overlay/          # Default overlay
│   ├── upper/                     # Changes are stored here
│   ├── work/                      # OverlayFS working directory
│   └── merged/                    # Combined view (use this for chroot)
└── trixie-amd64.projectA/         # Named overlay
    ├── upper/
    ├── work/
    └── merged/
```

### Benefits

1. **Base Protection**: Your original chroot environment is never modified
2. **Easy Reset**: Remove the overlay to return to a clean state
3. **Space Efficient**: Only changes consume additional disk space
4. **Multiple Experiments**: Create different named overlays from the same base

## Command Reference

### setup

Setup a chroot environment with essential filesystems.

```bash
# Normal mode
$ sudo chroot-prep setup -dir trixie-amd64

# OverlayFS mode (default name "overlay")
$ sudo chroot-prep setup -dir trixie-amd64 -overlay

# OverlayFS mode with custom name
$ sudo chroot-prep setup -dir trixie-amd64 -overlay projectA
```

**Options:**

- `-dir string`: Path to chroot directory (required)
- `-overlay [name]`: Use OverlayFS with optional name (default: "overlay")

### cleanup

Unmount filesystems from a chroot environment.

```bash
# Cleanup normal chroot
$ sudo chroot-prep cleanup -dir trixie-amd64

# Cleanup default overlay
$ sudo chroot-prep cleanup -dir trixie-amd64 -overlay

# Cleanup specific named overlay
$ sudo chroot-prep cleanup -dir trixie-amd64 -overlay projectA
```

**Options:**

- `-dir string`: Path to chroot directory (required)
- `-overlay [name]`: Cleanup specific overlay (default: "overlay")

### remove

Remove a chroot environment with automatic unmounting.

```bash
# Remove everything (base + all overlays)
$ sudo chroot-prep remove -dir trixie-amd64

# Remove only the default overlay (preserve base)
$ sudo chroot-prep remove -dir trixie-amd64 -overlay

# Remove specific named overlay (preserve base)
$ sudo chroot-prep remove -dir trixie-amd64 -overlay projectA

# Force removal even if unmounting fails
$ sudo chroot-prep remove -dir trixie-amd64 -force
```

**Options:**

- `-dir string`: Path to chroot directory (required)
- `-force`: Force removal even if unmount fails
- `-overlay [name]`: Remove only specific overlay (default: "overlay")

## Example: Multiple Overlays

```bash
# Create multiple independent environments from the same base
$ sudo chroot-prep setup -dir trixie-amd64 -overlay dev
$ sudo chroot-prep setup -dir trixie-amd64 -overlay test

# Work in different environments
$ sudo chroot trixie-amd64.dev/merged /bin/bash   # Development
$ sudo chroot trixie-amd64.test/merged /bin/bash  # Testing

# Remove specific overlay when done
$ sudo chroot-prep remove -dir trixie-amd64 -overlay test

# List all overlays
$ ls -d trixie-amd64*/
trixie-amd64/  trixie-amd64.dev/  trixie-amd64.overlay/
```

## Notes

- Always run with `sudo` or as root
- The tool automatically detects environment types
- When removing without `-overlay`, all overlays and the base are removed
- OverlayFS requires that upper and work directories are on the same filesystem
- The base directory remains read-only when using OverlayFS mode

## License

This project is licensed under the [MIT License](./LICENSE).
