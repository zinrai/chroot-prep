package main

// EnvironmentType represents the type of chroot environment
type EnvironmentType int

const (
	// NormalEnvironment is a standard chroot environment
	NormalEnvironment EnvironmentType = iota
	// OverlayEnvironment is a chroot environment using OverlayFS
	OverlayEnvironment
)

// Directory name constants for OverlayFS
const (
	OverlaySuffix = ".overlay"
	UpperDir      = "upper"
	WorkDir       = "work"
	MergedDir     = "merged"
)

// Mount type constants
const (
	MountTypeProc    = "proc"
	MountTypeSysfs   = "sysfs"
	MountTypeOverlay = "overlay"
)
