//go:build linux

package driver

import "golang.org/x/sys/unix"

const (
	tcgets = unix.TCGETS
	tcsets = unix.TCSETS
)
