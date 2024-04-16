package virtualbox

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

func (vbcmd command) setOptions(opts ...option) Command {
	var cmd Command = &vbcmd
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd
}

func (vbcmd command) isGuest() bool {
	return vbcmd.guest
}

func (vbcmd command) path() string {
	return vbcmd.program
}

func (vbcmd command) run(args ...string) (string, string, error) {
	defer vbcmd.setOptions(sudo(false))
	cmd := vbcmd.prepare(args)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		if ex, ok := err.(*exec.Error); ok && ex == exec.ErrNotFound {
			err = errors.New("command not found")
		}
	}
	return stdout.String(), stderr.String(), err
}

func sudo(sudo bool) option {
	return func(cmd Command) {
		vbcmd := cmd.(*command)
		vbcmd.sudo = sudo
	}
}

func isSudoer() (bool, error) {
	me, err := user.Current()
	if err != nil {
		return false, err
	}
	if groupIDs, err := me.GroupIds(); runtime.GOOS == "linux" {
		if err != nil {
			return false, err
		}
		for _, groupID := range groupIDs {
			group, err := user.LookupGroupId(groupID)
			if err != nil {
				return false, err
			}
			if group.Name == "sudo" {
				return true, nil
			}
		}
	}
	return false, nil
}

func lookProg(prog string) (string, error) {
	if runtime.GOOS == "windows" {
		if p := os.Getenv("VBOX_INSTALL_PATH"); p != "" {
			prog = filepath.Join(p, prog+".exe")
		} else {
			prog = filepath.Join("C:\\", "Program Files", "Oracle", "VirtualBox", prog+".exe")
		}
	}
	return exec.LookPath(prog)
}

func (vbcmd command) prepare(args []string) *exec.Cmd {
	program := vbcmd.program
	argv := []string{}
	if vbcmd.sudoer && vbcmd.sudo && runtime.GOOS != "windows" {
		program = "sudo"
		argv = append(argv, vbcmd.program)
	}
	for _, arg := range args {
		argv = append(argv, arg)
	}
	return exec.Command(program, argv...)
}

func Manage() Command {
	if manage != nil {
		return manage
	}
	sudoer, err := isSudoer()
	if err != nil {
		err = errors.New("error getting sudoer status")
	}
	if prog, err := lookProg("VBoxManage"); err == nil {
		manage = command{program: prog, sudoer: sudoer, guest: false}
	} else if prog, err := lookProg("VBoxControl"); err == nil {
		manage = command{program: prog, sudoer: sudoer, guest: true}
	} else {
		manage = command{program: "false", sudoer: false, guest: false}
	}
	return manage
}

func (m *VBox) SetCloudData(key, val string) error {
	_, _, err := Manage().run("setextradata", m.Name, key, val)
	return err
}

func (m *VBox) GetCloudData(key string) (*string, error) {
	value, _, err := Manage().run("getextradata", m.Name, key)
	if err != nil {
		return nil, err
	}
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "No value set") {
		return nil, nil
	}
	trimmed := strings.TrimPrefix(value, "Value: ")
	return &trimmed, nil
}
