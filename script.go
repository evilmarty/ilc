package main

import (
	"os"
	"os/exec"
)

var (
	ShellArgs = []string{"sh", "-c"}
)

func SetShell(args []string) {
	if len(args) > 0 {
		ShellArgs = args
	}
}

func ScriptCommand(script string) *exec.Cmd {
	args := append(ShellArgs, script)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
