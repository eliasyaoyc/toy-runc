package command

import (
	"errors"
	"github.com/urfave/cli"
	"toy-runc/internal/container"
)

var logCommand = cli.Command{
	Name:  "logs",
	Usage: "print logs of a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return errors.New("please input your container name")
		}

		containerName := context.Args().Get(0)
		container.LogContainer(containerName)
		return nil
	},
}
