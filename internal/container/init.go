package container

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// RunContainerInitProcess execute inside the container and using mount
// to mount the proc file system so that you can later use `ps` to view
// the current process resources etc.
func RunContainerInitProcess() error {
	cmdArray := readUserCommand()

	if cmdArray == nil || len(cmdArray) == 0 {
		return errors.New(fmt.Sprintf("run container get user command error; cmdArray is empty."))
	}

	path, err := exec.LookPath(cmdArray[0])

	if err != nil {
		logrus.Errorf("runC exec loop path error; %v", err)
		return err
	}
	// syscall.MS_NOEXEC: no other programs are allowed to run on this file system.
	// syscall.MS_NOSUID: set-user-id or set-group-id is not allowed when running programs in this system，
	// syscall.MS_NODEV: default param.
	//defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	//syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

	logrus.Infof("runC find path %s", path)

	// call int execve(cosnt char*filename, char*const argv[], char*const envp[]);
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		logrus.Errorf(err.Error())
	}
	return nil
}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		logrus.Errorf("runC read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}
