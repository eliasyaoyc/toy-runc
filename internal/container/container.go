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
	"text/tabwriter"
	"time"
)

const (
	ENV_EXEC_PID = "myrunc_pid"
	ENV_EXEC_CMD = "myrunc_cmd"
)

var (
	RUNNING             = "running"
	STOP                = "stopped"
	Exit                = "exited"
	DefaultInfoLocation = "/var/run/myRunc/%s/"
	ContainerLogFile    = "container.log"
	ConfigName          = "config.json"
	RootUrl             = "/root"
	MntUrl              = "/root/mnt/%s"
	WriteLayerUrl       = "/root/writeLayer/%s"
)

type ContainerInfo struct {
	Pid         string `json:"pid"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	CreatedTime string `json:"createdTime"`
	Status      string `json:"status"`
	Volume      string `json:"volume"`
}

func RecordContainerInfo(containerPID int, commandArray []string, containerName, volume string) (string, error) {
	id := randStringBytes(10)
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")
	if containerName == "" {
		containerName = id
	}
	containerInfo := &ContainerInfo{
		Pid:         strconv.Itoa(containerPID),
		Id:          id,
		Name:        containerName,
		Command:     command,
		CreatedTime: createTime,
		Status:      RUNNING,
		Volume:      volume,
	}
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		logrus.Errorf("record container info error; %v", err)
		return "", err
	}
	jsonStr := string(jsonBytes)

	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		logrus.Errorf("mkdir error %s, error %v", dirUrl, err)
		return "", err
	}
	fileName := dirUrl + "/" + ConfigName
	file, err := os.Create(fileName)
	if err != nil {
		logrus.Errorf("create file %s error; %v", fileName, err)
		return "", err
	}
	if _, err := file.WriteString(jsonStr); err != nil {
		logrus.Errorf("file write string error; %v", err)
		return "", err
	}
	return containerName, nil
}

func DeleteContainerInfo(containerId string) {
	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirUrl); err != nil {
		logrus.Errorf("remove dir %s error; %v", dirUrl, err)
	}
}

func getContainerInfo(file os.FileInfo) (*ContainerInfo, error) {
	containerName := file.Name()
	configFileDir := fmt.Sprintf(DefaultInfoLocation, containerName)
	configFileDir = configFileDir + ConfigName
	content, err := ioutil.ReadFile(configFileDir)

	if err != nil {
		logrus.Errorf("read file %s error; %v", configFileDir, err)
		return nil, err
	}
	var containerInfo ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		logrus.Errorf("json unmarshal error %v", err)
		return nil, err
	}
	return &containerInfo, nil
}

// NewParentProcess create the execution env for the current process.
// /proc/self/exe represent current program
// create namespace-isolated container processes.
func NewParentProcess(tty bool, containerName, volume, imageName string, envSlice []string) (*exec.Cmd, *os.File) {
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
			logrus.Errorf("NewParentProcess create file %s error; %v", stdLogFilePath, err)
			return nil, nil
		}
		cmd.Stdout = stdLogFile
	}

	cmd.ExtraFiles = []*os.File{readPipe}
	cmd.Env = append(os.Environ(), envSlice...)
	logrus.Infof("runC recv run command; %s", cmd.String())
	newWorkSpace(volume, imageName, containerName)
	cmd.Dir = fmt.Sprintf(MntUrl, containerName)
	return cmd, writePipe
}

func ListContainer() {
	dirUrl := fmt.Sprintf(DefaultInfoLocation, "")
	dirUrl = dirUrl[:len(dirUrl)-1]
	files, err := ioutil.ReadDir(dirUrl)
	if err != nil {
		logrus.Errorf("read dir %s error; %v", dirUrl, err)
		return
	}
	var containers []*ContainerInfo
	for _, file := range files {
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			logrus.Errorf("get container info error; %v", err)
			continue
		}
		containers = append(containers, tmpContainer)
	}
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime,
		)
	}
	if err := w.Flush(); err != nil {
		logrus.Errorf("flush error %v", err)
		return
	}
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

	os.Setenv(ENV_EXEC_PID, pid)
	os.Setenv(ENV_EXEC_CMD, cmdStr)
	containerEnvs := getEnvsByPid(pid)

	cmd.Env = append(os.Environ(), containerEnvs...)

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
