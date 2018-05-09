/*
Docker From Scratch Workshop - Level 0: Starting a new process.
Goal: We want to start a new linux process using the fork & exec model.
Note: At this level we don't care about containment yet.
Usage:
    running:
        contenedor run /bin/sh
    will:
        - fork a new process which will exec '/bin/sh'
        - while the parent waits for it to finish
*/

package level

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/docker/distribution/uuid"
	"github.com/overdrive3000/justforfunc32/contenedor/utils"
	"github.com/pkg/errors"
)

// Container data structure
type Container struct {
	Env          []string
	ImageName    string
	ImageDir     string
	ContainerDir string
	containerID  string
}

// exported constant
const (
	SUFFIX = "tar"

	USAGE = "run"
	SHORT = "contenerdor run /bin/sh"
	LONG  = `contenerdor run will:

- fork a new process which will exec '/bin/sh'
- while the parent waits for it to finish`
)

// container id short version
func (c *Container) shortID() string {
	return strings.Split(c.containerID, "-")[4]
}

// return complete image path
func (c *Container) getImagePath(suffix string) string {
	return filepath.Join(c.ImageDir, strings.Join([]string{c.ImageName, suffix}, "."))
}

// return complete root path
func (c *Container) getContainerPath() string {
	return filepath.Join(c.ContainerDir, c.containerID, "rootfs")
}

// Creates the container root folder
func createContainerRoot(c Container) (string, error) {
	// Check if image exists
	ip := c.getImagePath(SUFFIX)
	if _, err := os.Stat(ip); os.IsNotExist(err) {
		return "", errors.Wrap(err, "create container")
	}

	// Create container folder is doesn't exist
	cp := c.getContainerPath()
	if _, err := os.Stat(cp); os.IsNotExist(err) {
		if err = os.MkdirAll(cp, 0755); err != nil {
			return "", errors.Wrap(err, "create container")
		}
	}

	// Extract image in container root path
	reader, err := os.Open(ip)
	if err != nil {
		return "", errors.Wrap(err, "create container")
	}
	if err = utils.Untar(cp, reader); err != nil {
		return "", errors.Wrap(err, "create container")
	}

	return cp, nil
}

// Run execute a process using fork-exec
func Run(entrypoint []string, c Container) error {

	// Generate container id
	c.containerID = uuid.Generate().String()
	// Look for the full command path
	cmd, err := exec.LookPath(entrypoint[0])
	if err != nil {
		return errors.Wrap(err, "Run: cannot look executable path")
	}

	// Create container filesystem
	cp, err := createContainerRoot(c)
	if err != nil {
		return err
	}
	// Chroot within the new filesystem
	if err = syscall.Chroot(cp); err != nil {
		return errors.Wrapf(err, "Run: cannot chroot in %s", cp)
	}
	// Change to new root path
	if err = os.Chdir("/"); err != nil {
		return errors.Wrap(err, "Run: cannot chdir into / in chroot")
	}

	// Setting process attributes
	procAttr := os.ProcAttr{
		Dir:   "/",
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Env:   c.Env,
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

	/*just for testing
	fmt.Println("Container ID ", c.containerID)
	fmt.Println("Short ID ", c.shortID())
	fmt.Println("Image Path", c.getImagePath(SUFFIX))

	fmt.Println("Container Path", cp)*/

	return nil
}
