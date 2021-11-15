package utils

import (
	"bytes"
	"golang.org/x/sys/execabs"
	"os"
)

func IsRoot() bool {
	return os.Getegid() == 0
}

var canSudo = -1

func CanSudo() bool {
	if canSudo > 0 {
		return canSudo == 0
	}
	if IsRoot() {
		canSudo = 0
		return true
	}
	cmd := execabs.Command("sudo", "-n", "-v")
	data, err := cmd.CombinedOutput()
	if err != nil {
		if !bytes.HasPrefix(data, []byte("sudo:")) {
			canSudo = 1
			return false
		}
	}
	canSudo = 0
	return true
}
