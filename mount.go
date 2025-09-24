package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"syscall"
)

// mountProc mounts procfs to the target directory
func mountProc(target string) error {
	if isMounted(target) {
		fmt.Printf("%s is already mounted, skipping...\n", target)
		return nil
	}

	if err := syscall.Mount("none", target, "proc", 0, ""); err != nil {
		return fmt.Errorf("failed to mount proc at %s: %w", target, err)
	}

	return nil
}

// mountDev bind mounts /dev to the target directory
func mountDev(target string) error {
	if isMounted(target) {
		fmt.Printf("%s is already mounted, skipping...\n", target)
		return nil
	}

	if err := syscall.Mount("/dev", target, "none", syscall.MS_BIND, ""); err != nil {
		return fmt.Errorf("failed to mount dev at %s: %w", target, err)
	}

	return nil
}

// mountSys bind mounts /sys to the target directory
func mountSys(target string) error {
	if isMounted(target) {
		fmt.Printf("%s is already mounted, skipping...\n", target)
		return nil
	}

	if err := syscall.Mount("/sys", target, "none", syscall.MS_BIND, ""); err != nil {
		return fmt.Errorf("failed to mount sys at %s: %w", target, err)
	}

	return nil
}

// mountEssentialFS mounts all essential filesystems (/proc, /dev, /sys)
func mountEssentialFS(chrootDir string) error {
	// Verify required directories exist
	requiredDirs := []string{"dev", "proc", "sys"}
	for _, dir := range requiredDirs {
		fullPath := filepath.Join(chrootDir, dir)
		if !dirExists(fullPath) {
			return fmt.Errorf("required directory %s does not exist in chroot environment", dir)
		}
	}

	// Mount proc
	if err := mountProc(filepath.Join(chrootDir, "proc")); err != nil {
		return err
	}

	// Mount dev
	if err := mountDev(filepath.Join(chrootDir, "dev")); err != nil {
		return err
	}

	// Mount sys
	if err := mountSys(filepath.Join(chrootDir, "sys")); err != nil {
		return err
	}

	return nil
}

// umountEssentialFS unmounts all essential filesystems
func umountEssentialFS(chrootDir string) error {
	// Unmount in reverse order to handle dependencies
	mounts := []string{
		filepath.Join(chrootDir, "sys"),
		filepath.Join(chrootDir, "proc"),
		filepath.Join(chrootDir, "dev"),
	}

	var firstErr error
	for _, mountpoint := range mounts {
		if !isMounted(mountpoint) {
			fmt.Printf("%s is not mounted, skipping...\n", mountpoint)
			continue
		}

		if err := umountPath(mountpoint); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			// Continue unmounting other filesystems
		}
	}

	return firstErr
}

// mountOverlayFS mounts an overlay filesystem
func mountOverlayFS(lower, upper, work, merged string) error {
	if isMounted(merged) {
		return fmt.Errorf("overlay is already mounted at %s", merged)
	}

	// Construct mount options
	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lower, upper, work)

	// Mount overlay
	if err := syscall.Mount("overlay", merged, "overlay", 0, opts); err != nil {
		return fmt.Errorf("failed to mount overlay: %w", err)
	}

	return nil
}

// umountPath unmounts a filesystem at the given path
func umountPath(path string) error {
	// Try normal unmount first
	err := syscall.Unmount(path, 0)
	if err == nil {
		return nil
	}

	// Normal unmount failed, try lazy unmount
	fmt.Printf("Normal unmount failed for %s, trying lazy unmount...\n", path)
	if err := syscall.Unmount(path, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("failed to unmount %s: %w", path, err)
	}

	return nil
}

// isMounted checks if a path is mounted
func isMounted(mountpoint string) bool {
	cmd := exec.Command("mountpoint", "-q", mountpoint)
	return cmd.Run() == nil
}
