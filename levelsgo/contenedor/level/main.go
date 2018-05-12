package level

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

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/docker/distribution/uuid"
	"github.com/overdrive3000/justforfunc32/contenedor/utils"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
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

// mountDevices mount special devices and fs in container rootfs
func mountDevices(c Container) error {
	type device struct {
		dest string
		opts uintptr
		data string
	}
	fs := map[string]device{
		"proc": device{
			dest: "proc",
			opts: 0,
			data: "",
		},
		"sysfs": device{
			dest: "sys",
			opts: 0,
			data: "",
		},
		"tmpfs": device{
			dest: "dev",
			opts: syscall.MS_NOSUID | syscall.MS_STRICTATIME,
			data: "mode=755",
		},
	}
	for f, d := range fs {
		if err := syscall.Mount(f, filepath.Join(c.getContainerPath(), d.dest), f, d.opts, d.data); err != nil {
			return errors.Wrap(err, "-> mountDevices Failed to mount device")
		}
	}
	return nil
}

func createDevices(c Container) error {
	// Create devpts folder is doesn't exist
	dpts := filepath.Join(c.getContainerPath(), "dev", "pts")
	if _, err := os.Stat(dpts); os.IsNotExist(err) {
		if err = os.MkdirAll(dpts, 0755); err != nil {
			return errors.Wrap(err, "create devpts")
		}
	}
	if err := syscall.Mount("devpts", dpts, "devpts", 0, ""); err != nil {
		return errors.Wrap(err, "failed to mount devpts")
	}
	for i, fd := range []string{"stdin", "stdout", "stderr"} {
		if err := os.Symlink(filepath.Join("/proc/self/fd", strconv.Itoa(i)), filepath.Join(c.getContainerPath(), "dev", fd)); err != nil {
			return errors.Wrap(err, "failed to create fd")
		}
	}

	// create devices
	devices := []struct {
		name  string
		attr  uint32
		major uint32
		minor uint32
	}{
		{name: "null", attr: 0666 | unix.S_IFCHR, major: 1, minor: 3},
		{name: "zero", attr: 0666 | unix.S_IFCHR, major: 1, minor: 3},
		{name: "random", attr: 0666 | unix.S_IFCHR, major: 1, minor: 8},
		{name: "urandom", attr: 0666 | unix.S_IFCHR, major: 1, minor: 9},
		{name: "console", attr: 0666 | unix.S_IFCHR, major: 136, minor: 1},
		{name: "tty", attr: 0666 | unix.S_IFCHR, major: 5, minor: 0},
		{name: "full", attr: 0666 | unix.S_IFCHR, major: 1, minor: 7},
	}
	for _, dev := range devices {
		dt := int(unix.Mkdev(dev.major, dev.minor))
		if err := unix.Mknod(filepath.Join(c.getContainerPath(), "dev", dev.name), dev.attr, dt); err != nil {
			return errors.Wrap(err, "failed to create device")
		}
	}
	return nil
}

// Run execute a process using fork-exec
func Run(entrypoint []string, c Container) error {

	// Generate container id
	c.containerID = uuid.Generate().String()
	// Look for the full command path
	cmd, err := exec.LookPath(entrypoint[0])
	if err != nil {
		return errors.Wrap(err, "Run:")
	}

	// Create container filesystem
	cp, err := createContainerRoot(c)
	if err != nil {
		return err
	}

	// Avoid mount point propagation
	if err := syscall.Mount("", "/", "none", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return errors.Wrap(err, "Run: failed to set / private")
	}

	// Mount special fs and devices
	if err = mountDevices(c); err != nil {
		return errors.Wrap(err, "Run: failed to mount devices")
	}

	// Create special devices
	if err = createDevices(c); err != nil {
		return errors.Wrap(err, "Run: failed to create devices")
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
		Sys: &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWNS,
		},
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
