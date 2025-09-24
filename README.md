# chroot-prep

A command-line tool to manage filesystem mounts for chroot environments with OverlayFS support.

## Features

- Setup essential filesystems (`/dev`, `/proc`, `/sys`) into chroot environments
- Setup host's resolv.conf(5) to chroot environments
- **OverlayFS support** for layered chroot environments
- Preserve base environments while making experimental changes
- Automatic environment type detection (normal vs overlay)
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
# Setup with OverlayFS (preserves the base)
$ sudo chroot-prep setup -dir trixie-amd64 -overlay

# Enter the overlay chroot
$ sudo chroot trixie-amd64.overlay/merged /bin/bash

# When done, cleanup
$ sudo chroot-prep cleanup -dir trixie-amd64
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
└── trixie-amd64.overlay/          # Overlay management directory
    ├── upper/                     # Changes are stored here
    ├── work/                      # OverlayFS working directory
    └── merged/                    # Combined view (use this for chroot)
```

### Benefits

1. **Base Protection**: Your original chroot environment is never modified
2. **Easy Reset**: Remove the overlay to return to a clean state
3. **Space Efficient**: Only changes consume additional disk space
4. **Multiple Experiments**: Create different overlays from the same base

## Command Reference

### setup

Setup a chroot environment with essential filesystems.

```bash
# Normal mode
$ sudo chroot-prep setup -dir trixie-amd64

# OverlayFS mode
$ sudo chroot-prep setup -dir trixie-amd64 -overlay
```

**Options:**
- `-dir string`: Path to chroot directory (required)
- `-overlay`: Use OverlayFS with the directory as read-only base

### cleanup

Unmount filesystems from a chroot environment. Automatically detects the environment type.

```bash
$ sudo chroot-prep cleanup -dir trixie-amd64
```

**Options:**
- `-dir string`: Path to chroot directory (required)

### remove

Remove a chroot environment with automatic unmounting.

```bash
# Remove everything (base + overlay if present)
$ sudo chroot-prep remove -dir trixie-amd64

# Remove only the overlay (preserve base)
$ sudo chroot-prep remove -dir trixie-amd64 -overlay

# Force removal even if unmounting fails
$ sudo chroot-prep remove -dir trixie-amd64 -force
```

**Options:**
- `-dir string`: Path to chroot directory (required)
- `-force`: Force removal even if unmount fails
- `-overlay`: Remove only overlay directory (preserve base)

## Notes

- Always run with `sudo` or as root
- The tool automatically detects whether an environment is using OverlayFS
- When removing a directory, the tool will attempt to cleanup mounts first
- OverlayFS requires that upper and work directories are on the same filesystem
- The base directory remains read-only when using OverlayFS mode

## License

This project is licensed under the [MIT License](./LICENSE).
