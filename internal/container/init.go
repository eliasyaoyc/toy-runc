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

var (
	containerInitCmdError = errors.New("run container get user command error; cmdArray is empty")
)

// RunContainerInitProcess execute inside the container and using mount
// to mount the proc file system so that you can later use `ps` to view
// the current process resources etc.
func RunContainerInitProcess() error {
	cmdArray := readUserCommand()

	if cmdArray == nil || len(cmdArray) == 0 {
		return containerInitCmdError
	}

	// init mount point.
	if err := setUpMount(); err != nil {
		return err
	}

	path, err := exec.LookPath(cmdArray[0])

	if err != nil {
		return fmt.Errorf("exec loop path error; %w", err)
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

func setUpMount() error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("pwd error:%v", err)
	}

	if err = pivotRoot(pwd); err != nil {
		return err
	}

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		return fmt.Errorf("mount proc error: %v", err)
	}

	err = syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	if err != nil {
		return fmt.Errorf("mount tmpfs error: %v", err)
	}

	return nil

}

func pivotRoot(root string) error {
	// systemd 加入linux之后, mount namespace 就变成 shared by default, 必须显式声明新的mount namespace独立。
	err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		return err
	}

	// 重新mount root
	// bind mount：将相同内容换挂载点
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs to itself error: %v", err)
	}

	// 创建 rootfs/.pivot_root 存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}

	// pivot_root 到新的rootfs, 老的 old_root挂载在rootfs/.pivot_root
	// 挂载点现在依然可以在mount命令中看到
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root error: %v", err)
	}

	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir error: %v", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")

	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir error: %v", err)
	}

	// 删除临时文件夹
	return os.Remove(pivotDir)
}
