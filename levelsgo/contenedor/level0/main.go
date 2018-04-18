/*
Docker From Scratch Workshop - Level 0: Starting a new process.
Goal: We want to start a new linux process using the fork & exec model.
Note: At this level we don't care about containment yet.
Usage:
    running:
        rd.py run /bin/sh
    will:
        - fork a new process which will exec '/bin/sh'
        - while the parent waits for it to finish
*/

package level0

import (
	"fmt"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"
)

// exported constant
const (
	USAGE = "run"
	SHORT = "contenerdor run /bin/sh"
	LONG  = `will:

- fork a new process which will exec '/bin/sh
- while the parent waits for it to finish`
)

// Run execute a process using fork-exec
func Run(entrypoint []string) error {
	args := make([]string, len(entrypoint))
	if len(entrypoint) > 1 {
		args = entrypoint[1:]
	}
	cmd, err := exec.LookPath(entrypoint[0])
	if err != nil {
		return errors.Wrap(err, "Run: cannot look executable path")
	}
	procAttr := syscall.ProcAttr{
		Dir:   "/tmp",
		Files: []uintptr{uintptr(syscall.Stdin), uintptr(syscall.Stdout), uintptr(syscall.Stderr)},
		Env:   []string{},
		Sys: &syscall.SysProcAttr{
			Foreground: false,
		},
	}
	pid, err := syscall.ForkExec(cmd, args, &procAttr)
	fmt.Println("Spawned proc", pid)
	return nil
}
