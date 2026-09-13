//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package driver

import "golang.org/x/sys/unix"

const (
	tcgets = unix.TIOCGETA
	tcsets = unix.TIOCSETA
)
