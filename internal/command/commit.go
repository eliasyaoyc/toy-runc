package command

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os/exec"
)

var commitError = errors.New("missing container name")

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return commitError
		}
		imageName := context.Args().Get(0)
		commitContainer(imageName)
		return nil
	},
}

func commitContainer(imageName string) {
	mntUrl := "/root/mnt"
	imageTar := "/root/" + imageName + ".tar"
	logrus.Infof("%s", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntUrl, ".").CombinedOutput(); err != nil {
		logrus.Errorf("tar folder %s error; %v", mntUrl, err)
	}
}
