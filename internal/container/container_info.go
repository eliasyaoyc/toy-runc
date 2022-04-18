package container

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

type ContainerInfo struct {
	Pid         string `json:"pid"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	CreatedTime string `json:"createdTime"`
	Status      string `json:"status"`
}

var (
	RUNNING             string = "running"
	STOP                string = "stopped"
	Exit                string = "exited"
	DefaultInfoLocation string = "/var/run/myRunc/%s/"
	ConfigName          string = "config.json"
)

func RecordContainerInfo(containerPID int, commandArray []string, containerName string) (string, error) {
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

func randStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
