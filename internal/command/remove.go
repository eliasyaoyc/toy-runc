package command

import (
	"fmt"
	"github.com/urfave/cli"
	"toy-runc/internal/container"
)

var removeCommand = cli.Command{
	Name:  "rm",
	Usage: "remove unused containers",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing contaienr name")
		}
		containerName := context.Args().Get(0)
		container.RemoveContainer(containerName)
		return nil
	},
}
