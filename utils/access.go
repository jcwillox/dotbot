//go:build !windows
// +build !windows

package utils

import (
	"golang.org/x/sys/unix"
)

func IsWritable(path string) bool {
	return unix.Access(path, unix.W_OK) == nil
}
