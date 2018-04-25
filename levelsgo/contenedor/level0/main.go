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
	"os"
	"os/exec"

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
	cmd, err := exec.LookPath(entrypoint[0])
	if err != nil {
		return errors.Wrap(err, "Run: cannot get command path")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "Run: cannot get current directory")
	}
	pa := os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Dir:   cwd,
	}
	// Start given process via fork-exec method
	proc, err := os.StartProcess(cmd, entrypoint, &pa)
	if err != nil {
		return errors.Wrapf(err, "Run: cannot run the given command with pid %d", proc.Pid)
	}
	// Wait until process finishes
	state, err := proc.Wait()
	if err != nil {
		return errors.Wrapf(err, "ERROR")
	}
	fmt.Printf("%d exited with status %d\n", proc.Pid, state)

	return nil
}
