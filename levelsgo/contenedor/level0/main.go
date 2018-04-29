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
func Run(entrypoint []string, envars []string) error {

	// Look for the full command path
	cmd, err := exec.LookPath(entrypoint[0])
	if err != nil {
		return errors.Wrap(err, "Run: cannot look executable path")
	}

	// Obtain current directory
	cwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "Run: cannot get current directory")
	}

	// Setting process attributes
	procAttr := os.ProcAttr{
		Dir:   cwd,
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Env:   envars,
	}

	proc, err := os.StartProcess(cmd, entrypoint, &procAttr)
	if err != nil {
		return errors.Wrapf(err, "Run: cannot run %s", entrypoint)
	}

	// Wait until process ends
	state, err := proc.Wait()
	if err != nil {
		return errors.Wrapf(err, "Run: error while running process %d", proc.Pid)
	}

	fmt.Printf("%d %s\n", proc.Pid, state.String())

	return nil
}
