//go:build linux
// +build linux

package container

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess create the execution env for the current process.
// /proc/self/exe represent current program
// create namespace-isolated container processes.
func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWIPC,
	}

	readPipe, writePipe, err := newPipe()
	if err != nil {
		logrus.Errorf("new pipe error %v", err)
		return nil, nil
	}

	cmd.ExtraFiles = []*os.File{readPipe}

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	logrus.Infof("runC recv run command; %s", cmd.String())

	return cmd, writePipe
}

func newPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	return read, write, nil
}
