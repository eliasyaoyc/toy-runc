package command

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os/exec"
)

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return errors.New(fmt.Sprintf("missing container name."))
		}
		imageName := context.Args().Get(0)
		commitContainer(imageName)
		return nil
	},
}

func commitContainer(imageName string) {
	mntUrl := "/root/mnt"
	imageTar := "/root/" + imageName + ".tar"
	logrus.Info("%s", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntUrl, ".").CombinedOutput(); err != nil {
		logrus.Errorf("tar folder %s error; %v", mntUrl, err)
	}
}
