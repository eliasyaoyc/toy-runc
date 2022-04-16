package command

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"toy-runc/internal/container"
)

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(context *cli.Context) error {
		logrus.Info("runC init begin.")
		if err := container.RunContainerInitProcess(); err != nil {
			logrus.Errorf("runC init command error; %v", err)
			return err
		}
		return nil
	},
}
