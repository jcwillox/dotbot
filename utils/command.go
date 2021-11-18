package utils

import (
	"errors"
	"github.com/creasty/defaults"
	"github.com/google/shlex"
	"golang.org/x/sys/execabs"
	"gopkg.in/yaml.v3"
	"os"
	"runtime"
	"strings"
)

type Command struct {
	Command string
	Shell   bool `default:"true"`
	// Enable writing to stdin
	Stdin bool `default:"true"`
	// Enable piping to stdout
	Stdout bool `default:"true"`
	// Enable piping to stderr
	Stderr bool `default:"true"`
	// Run command as root
	Sudo bool
	// Attempt to run command as root
	TrySudo bool `yaml:"try_sudo"`
}

func (c *Command) UnmarshalYAML(n *yaml.Node) error {
	defaults.MustSet(c)
	type CommandT Command
	return n.Decode((*CommandT)(c))
}

func (c Command) Cmd() (*execabs.Cmd, error) {
	var cmd *execabs.Cmd

	sudo, err := c.needsSudo()
	if err != nil {
		return nil, err
	}

	if c.Shell {
		shell, args := GetShellCommand(c.Command)
		if sudo {
			cmd = execabs.Command("sudo", append([]string{"-E", shell}, args...)...)
		} else {
			cmd = execabs.Command(shell, args...)
		}
	} else {
		args, err := shlex.Split(c.Command)
		if err != nil {
			return nil, err
		}
		if sudo {
			cmd = execabs.Command("sudo", append([]string{"-E"}, args...)...)
		} else {
			cmd = execabs.Command(args[0], args[1:]...)
		}
	}

	if c.Stdin {
		cmd.Stdin = os.Stdin
	}
	if c.Stdout {
		cmd.Stdout = os.Stdout
	}
	if c.Stderr {
		cmd.Stderr = os.Stderr
	}

	return cmd, nil
}

func (c Command) needsSudo() (bool, error) {
	if !IsRoot() && (c.Sudo || c.TrySudo) {
		if CanSudo() {
			c.Stdin = true
			return true, nil
		} else if !c.TrySudo {
			return false, errors.New("unable to sudo")
		}
	}
	return false, nil
}

func (c Command) String() string {
	return c.Command
}

func (c Command) ShortString() string {
	s := strings.SplitN(c.Command, "\n", 2)[0]
	if s != c.Command {
		return s + "..."
	}
	return s
}

func GetShellCommand(command string) (string, []string) {
	if runtime.GOOS == "windows" {
		// we use powershell on Windows as cmd is not particularly useful for running scripts
		// or commandline snippets and directly running executables should not use shell mode
		return "powershell.exe", []string{"-NoProfile", "-NoLogo", "-Command", command}
		//return "cmd.exe", []string{"/c", command}
	}
	return GetDefaultShell(), []string{"-c", command}
}

func GetDefaultShell() string {
	for _, shell := range []string{"bash", "ash", "sh"} {
		path, _ := execabs.LookPath(shell)
		if path != "" {
			return path
		}
	}
	// return bash anyway as command will fail when executed
	return "bash"
}
