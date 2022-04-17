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
	readPipe, writePipe, err := newPipe()
	if err != nil {
		logrus.Errorf("new pipe error %v", err)
		return nil, nil
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWIPC,
	}

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	cmd.ExtraFiles = []*os.File{readPipe}
	logrus.Infof("runC recv run command; %s", cmd.String())
	mntUrl := "/root/mnt/"
	rootUrl := "/root/"
	newWorkSpace(rootUrl, mntUrl)
	cmd.Dir = mntUrl
	return cmd, writePipe
}

func newPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	return read, write, nil
}

func newWorkSpace(rootUrl string, mntUrl string) {
	createReadOnlyLayer(rootUrl)
	createWriteLayer(rootUrl)
	createMountpoint(rootUrl, mntUrl)
}

func createReadOnlyLayer(rootUrl string) {
	busyboxUrl := rootUrl + "busybox/"
	busyboxTarUrl := rootUrl + "busybox.tar"
	exist, err := pathExist(busyboxUrl)
	if err != nil {
		logrus.Errorf("fail to judge whether dir %s exists error; %v", busyboxUrl, err)
	}
	if exist == false {
		if err := os.Mkdir(busyboxUrl, 0777); err != nil {
			logrus.Errorf("mkdir dir %s error; %v", busyboxUrl, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarUrl, "-C", busyboxUrl).CombinedOutput(); err != nil {
			logrus.Errorf("unTar dir %s error; %v", busyboxTarUrl, err)
		}
	}
}

func createWriteLayer(rootUrl string) {
	writeUrl := rootUrl + "writeLayer/"
	if err := os.Mkdir(writeUrl, 0777); err != nil {
		logrus.Errorf("mkdir dir %s error; %v", writeUrl, err)
	}
}

func createMountpoint(rootUrl, mntUrl string) {
	if err := os.Mkdir(mntUrl, 0777); err != nil {
		logrus.Errorf("mkdir dir %s error; %v", mntUrl, err)
	}
	dirs := "dirs=" + rootUrl + "writeLayer:" + rootUrl + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		logrus.Errorf("%v", err)
	}
}

func pathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func DeleteWorkSpace(rootUrl, mntUrl string) {
	deleteMountPoint(rootUrl, mntUrl)
	deleteWriteLayer(rootUrl)
}

func deleteMountPoint(rootUrl, mntUrl string) {
	cmd := exec.Command("umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("error %v", err)
	}
	if err := os.RemoveAll(mntUrl); err != nil {
		logrus.Errorf("remove dir %s error; %v", mntUrl, err)
	}
}

func deleteWriteLayer(rootUrl string) {
	writeUrl := rootUrl + "writeLayer/"
	if err := os.RemoveAll(writeUrl); err != nil {
		logrus.Errorf("remove dir %s error; %v", writeUrl, err)
	}
}
