# chroot-prep

A command-line tool to manage filesystem mounts for chroot environments.

## Features

- Mount essential filesystems (`/dev`, `/proc`, `/sys`) into chroot environments
- Unmount filesystems from chroot environments
- Clean up chroot environments with automatic unmounting

## Requirements

- Linux operating system
- Root privileges ( sudo )

## Installation

build from source:

```bash
$ go build
```

## Usage

### Mount filesystems

```bash
$ sudo chroot-prep mount -dir /path/to/chroot
```

This will mount:

- `/dev` (bind mount)
- `/proc` (procfs)
- `/sys` (bind mount)

### Unmount filesystems

```bash
$ sudo chroot-prep umount -dir /path/to/chroot
```

### Remove chroot environment

Normal removal (with automatic unmounting):

```bash
$ sudo chroot-prep remove -dir /path/to/chroot
```

Force removal even if unmount fails:

```bash
$ sudo chroot-prep remove -f -dir /path/to/chroot
```

## License

This project is licensed under the MIT License - see the [LICENSE](https://opensource.org/license/mit) for details.
