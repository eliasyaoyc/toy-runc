package container

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
)

func getEnvsByPid(pid string) []string {
	path := fmt.Sprintf("/proc/%s/environ", pid)
	contentBytes, err := ioutil.ReadFile(path)

	if err != nil {
		logrus.Errorf("read file %s error; %v", path, err)
		return nil
	}
	envs := strings.Split(string(contentBytes), "\u0000")
	return envs
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

func randStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func newPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	return read, write, nil
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
