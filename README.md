# chroot-prep

A command-line tool to manage filesystem mounts for chroot environments.

## Features

- Setup essential filesystems (`/dev`, `/proc`, `/sys`) into chroot environments
- Setup host's resolv.conf(5) to chroot environments
- Cleanup filesystems from chroot environments
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

### Setup environment

```bash
$ sudo chroot-prep setup -dir /path/to/chroot
```

This will mount:

- `/dev` (bind mount)
- `/proc` (procfs)
- `/sys` (bind mount)

### Cleanup environment

```bash
$ sudo chroot-prep cleanup -dir /path/to/chroot
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
