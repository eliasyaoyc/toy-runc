package command

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os/exec"
	"toy-runc/internal/container"
)

var commitError = errors.New("missing container name")

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 2 {
			return commitError
		}
		containerName := context.Args().Get(0)
		imageName := context.Args().Get(1)
		commitContainer(containerName, imageName)
		return nil
	},
}

func commitContainer(containerName, imageName string) {
	mntURL := fmt.Sprintf(container.MntUrl, containerName)
	mntURL += "/"

	imageTar := container.RootUrl + "/" + imageName + ".tar"
	logrus.Infof("%s", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		logrus.Errorf("tar folder %s error; %v", mntURL, err)
	}
}
