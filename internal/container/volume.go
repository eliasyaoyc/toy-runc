package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

func newWorkSpace(volume, imageName, containerName string) {
	createReadOnlyLayer(imageName)
	createWriteLayer(containerName)
	createMountpoint(containerName, imageName)
	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			mountVolume(volumeURLs, containerName)
			logrus.Infof("newWorkSpace volume urls %q", volumeURLs)
			return
		}
		logrus.Errorf("volume param input incorrect")
	}
}

func createReadOnlyLayer(imageName string) {
	unTarFolderUrl := RootUrl + "/" + imageName + "/"
	imageUrl := RootUrl + "/" + imageName + ".tar"
	exist, err := pathExist(unTarFolderUrl)
	if err != nil {
		logrus.Errorf("fail to judge whether dir %s exists error; %v", unTarFolderUrl, err)
		return
	}
	if !exist {
		if err := os.Mkdir(unTarFolderUrl, 0622); err != nil {
			logrus.Errorf("mkdir dir %s error; %v", unTarFolderUrl, err)
			return
		}
		if _, err := exec.Command("tar", "-xvf", imageUrl, "-C", unTarFolderUrl).CombinedOutput(); err != nil {
			logrus.Errorf("unTar dir %s error; %v", unTarFolderUrl, err)
			return
		}
	}
}

func createWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayerUrl, containerName)
	if err := os.Mkdir(writeURL, 0777); err != nil {
		logrus.Errorf("mkdir dir %s error. %v", writeURL, err)
	}
}

// Union filesystem.
func createMountpoint(containerName, imageName string) {
	mntURL := fmt.Sprintf(MntUrl, containerName)
	if err := os.Mkdir(mntURL, 0777); err != nil {
		logrus.Errorf("mkdir dir %s error. %v", mntURL, err)
		return
	}

	writeLayer := fmt.Sprintf(WriteLayerUrl, containerName)
	imageLocation := RootUrl + "/" + imageName

	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
		imageLocation, writeLayer, RootUrl+"/temp")

	cmd := exec.Command("mount", "-t", "overlay", "-o", options, "overlay", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("newWorkSpace create mountpoint error; %v", err)
	}
}

func mountVolume(volumeURLs []string, containerName string) error {
	parentUrl := volumeURLs[0]
	if err := os.Mkdir(parentUrl, 0777); err != nil {
		logrus.Infof("Mkdir parent dir %s error. %v", parentUrl, err)
	}
	containerUrl := volumeURLs[1]
	mntURL := fmt.Sprintf(MntUrl, containerName)
	containerVolumeURL := mntURL + "/" + containerUrl
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		logrus.Infof("Mkdir container dir %s error. %v", containerVolumeURL, err)
	}
	dirs := "dirs=" + parentUrl
	_, err := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL).CombinedOutput()
	if err != nil {
		logrus.Errorf("Mount volume failed. %v", err)
		return err
	}
	return nil
}

func DeleteWorkSpace(volume, containerName string) {
	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			deleteMountPointWithVolume(volumeURLs, containerName)
		} else {
			deleteMountPoint(containerName)
		}
	} else {
		deleteMountPoint(containerName)
	}
	deleteWriteLayer(containerName)
}

func deleteMountPoint(containerName string) {
	mntURL := fmt.Sprintf(MntUrl, containerName)
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("error %v", err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Errorf("remove dir %s error; %v", mntURL, err)
	}
}

func deleteMountPointWithVolume(volumeURLs []string, containerName string) error {
	mntURL := fmt.Sprintf(MntUrl, containerName)
	containerUrl := mntURL + "/" + volumeURLs[1]
	if _, err := exec.Command("umount", containerUrl).CombinedOutput(); err != nil {
		logrus.Errorf("Umount volume %s failed. %v", containerUrl, err)
		return err
	}

	if _, err := exec.Command("umount", mntURL).CombinedOutput(); err != nil {
		logrus.Errorf("Umount mountpoint %s failed. %v", mntURL, err)
		return err
	}

	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Errorf("Remove mountpoint dir %s error %v", mntURL, err)
	}

	return nil
}

func deleteWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayerUrl, containerName)
	if err := os.RemoveAll(writeURL); err != nil {
		logrus.Errorf("remove dir %s error; %v", writeURL, err)
	}
}
