package sudo

import (
	"bytes"
	"github.com/jcwillox/dotbot/store"
	"golang.org/x/sys/execabs"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"syscall"
)

func IsRoot() bool {
	return os.Getegid() == 0
}

var (
	HasUsedSudo = false
	canSudo     = -1
)

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

func WouldSudo() bool {
	return !IsRoot() && CanSudo()
}

func Configs(configs interface{}) error {
	if !WouldSudo() {
		// we shouldn't be able to reach this, but we also want to
		// ensure we don't recursively sudo
		return os.ErrPermission
	}

	vars := store.GetVars()
	if len(vars) > 0 {
		configs = map[string]interface{}{"config": configs, "vars": vars}
	}

	path, err := os.Executable()
	if err != nil {
		log.Panicln("Failed to get dotbot executable path", err)
	}

	cmd := execabs.Command("sudo", "-E", path, "run", "--stdin")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = append(os.Environ(), "DOTBOT_SUDO=true")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		stdin.Close()
		return err
	}

	data, err := yaml.Marshal(configs)
	if err != nil {
		stdin.Close()
		cmd.Wait()
		return err
	}

	_, err = stdin.Write(data)
	if err != nil {
		return err
	}

	stdin.Close()
	if err := cmd.Wait(); err != nil {
		return err
	}
	HasUsedSudo = true
	return nil
}

func Config(directive string, config interface{}) error {
	return Configs([]map[string]interface{}{{directive: config}})
}

func IsPermission(err error) bool {
	if os.IsPermission(err) {
		return true
	}
	if err, ok := err.(*os.PathError); ok && err.Err == syscall.EACCES {
		return true
	}
	return false
}
