//go:build linux
// +build linux

package container

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"toy-runc/internal/command"
)

// NewParentProcess create the execution env for the current process.
// /proc/self/exe represent current program
// create namespace-isolated container processes.
func NewParentProcess(tty bool, containerName string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := newPipe()
	if err != nil {
		logrus.Errorf("new pipe error %v", err)
		return nil, nil
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
		if err := os.MkdirAll(dirURL, 0622); err != nil {
			logrus.Errorf("NewParentProcess mkdir %s error; %v", dirURL, err)
			return nil, nil
		}
		stdLogFilePath := dirURL + ContainerLogFile
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			logrus.Errorf("NewParentProcess create file %s error; %v", stdLogFile, err)
			return nil, nil
		}
		cmd.Stdout = stdLogFile
	}

	mntURL := "/root/mnt/"
	rootURL := "/root/"

	cmd.ExtraFiles = []*os.File{readPipe}
	cmd.Dir = mntURL
	logrus.Infof("runC recv run command; %s", cmd.String())
	newWorkSpace(rootURL, mntURL)
	return cmd, writePipe
}

func LogContainer(containerName string) {
	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
	logFileLocation := dirURL + ContainerLogFile
	file, err := os.Open(logFileLocation)
	defer file.Close()
	if err != nil {
		logrus.Errorf("log container open file %s error; %v", logFileLocation, err)
		return
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		logrus.Errorf("log container read file %s error; %v", logFileLocation, err)
		return
	}
	fmt.Fprint(os.Stdout, string(content))
}

func ExecContainer(containerName string, cmdArray []string) {
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		logrus.Errorf("exec container getContainerPidByName %s error; %v", containerName, err)
		return
	}
	cmdStr := strings.Join(cmdArray, " ")
	logrus.Infof("container pid %s", pid)
	logrus.Infof("command %s", cmdStr)

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	os.Setenv(command.ENV_EXEC_PID, pid)
	os.Setenv(command.ENV_EXEC_CMD, cmdStr)

	if err := cmd.Run(); err != nil {
		logrus.Errorf("exec container %s error; %v", containerName, err)
	}
}

func StopContainer(containerName string) {
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		logrus.Errorf("get container pid by name %s error; %v", containerName, err)
		return
	}

	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		logrus.Errorf("conver pid from string to int error; %v", err)
		return
	}

	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		logrus.Errorf("stop container %s error; %v", containerName, err)
	}

	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		logrus.Errorf("get container %s info error; %v", containerName, err)
		return
	}

	containerInfo.Status = STOP
	containerInfo.Pid = " "
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		logrus.Errorf("json marshal %s,error; %v", containerName, err)
		return
	}
	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
	configFilePath := dirURL + ConfigName
	if err := ioutil.WriteFile(configFilePath, newContentBytes, 0622); err != nil {
		logrus.Errorf("write file %s error; %v", configFilePath, err)
	}
}

func RemoveContainer(containerName string) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		logrus.Errorf("get container %s info error; %v", containerName, err)
		return
	}
	if containerInfo.Status != STOP {
		logrus.Errorf("could't remove running container")
		return
	}
	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
	if err := os.RemoveAll(dirURL); err != nil {
		logrus.Errorf("remove file %s error; %v", dirURL, err)
		return
	}
}

func getContainerInfoByName(containerName string) (*ContainerInfo, error) {
	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
	configFilePath := dirURL + ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		logrus.Errorf("read file %s error; %v", configFilePath, err)
		return nil, err
	}
	var containerInfo ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		logrus.Errorf("getContainerInfoByName unmarshal error; %v", err)
		return nil, err
	}
	return &containerInfo, nil
}

func getContainerPidByName(containerName string) (string, error) {
	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
	configFilePath := dirURL + ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		logrus.Errorf("read file %s error; %v", configFilePath, err)
		return "", err
	}
	var containerInfo ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
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

func createWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.Mkdir(writeURL, 0777); err != nil {
		logrus.Errorf("mkdir dir %s error. %v", writeURL, err)
	}
}

// Union filesystem.
func createMountpoint(rootURL, mntURL string) {
	if err := os.Mkdir(mntURL, 0777); err != nil {
		logrus.Errorf("mkdir dir %s error. %v", mntURL, err)
	}

	//dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"

	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
		rootURL+"busybox", rootURL+"writeLayer", rootURL+"temp")

	cmd := exec.Command("mount", "-t", "overlay", "-o", options, "overlay", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("newWorkSpace create mountpoint error; %v", err)
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
	deleteMountPoint(mntUrl)
	deleteWriteLayer(rootUrl)
}

func deleteMountPoint(mntUrl string) {
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
