//go:build !linux && !darwin && !windows

package checker

import (
	"os/exec"
	"time"
)

func configureProcess(command *exec.Cmd) { command.WaitDelay = 2 * time.Second }
