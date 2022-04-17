package container

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

	// init mount point.
	setUpMount()

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

func setUpMount() {
	pwd, err := os.Getwd()
	if err != nil {
		logrus.Errorf("get current location err; %v", err)
		return
	}
	logrus.Infof("current location is %s", pwd)
	pivotRoot(pwd)

	// mount proc
	defaultMointFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMointFlags), "")
	if err != nil {
		logrus.Errorf("mount proc error; %v", err)
		return
	}

	err = syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	if err != nil {
		logrus.Errorf("mount tmpfs error; %v", err)
		return
	}
}

func pivotRoot(root string) error {
	// systemd 加入linux 之后， mount namespace 就编程 shared by default, 必须显示声明新的 mount namespac 独立
	err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		return errors.New(fmt.Sprintf("mount systemd error; %v", err))
	}

	// 重新mount root
	// bind mount 将相同内容换挂载点
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return errors.New(fmt.Sprintf("mount rootfs to itself error: %v", err))
	}

	// 创建 rootfs/.pivot_root 存储旧的 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}

	// pivot_root 到新的rootfs, 老的 old_root挂载在rootfs/.pivot_root
	// 挂载点现在依然可以在mount命令中看到
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return errors.New(fmt.Sprintf("pivot_root error; %v", err))
	}

	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return errors.New(fmt.Sprintf("chdir / error;%v", err))
	}

	pivotDir = filepath.Join("/", ".pivot_root")

	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return errors.New(fmt.Sprintf("unmount pivot_root dir error; %v", err))
	}

	return os.Remove(pivotDir)
}
