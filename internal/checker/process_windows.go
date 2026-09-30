//go:build windows

package checker

import (
	"os/exec"
	"strconv"
	"time"
)

func configureProcess(command *exec.Cmd) {
	command.Cancel = func() error {
		if err := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(command.Process.Pid)).Run(); err == nil {
			return nil
		}
		return command.Process.Kill()
	}
	command.WaitDelay = 2 * time.Second
}
