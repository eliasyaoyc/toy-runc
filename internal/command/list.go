package command

import (
	"github.com/urfave/cli"
	"toy-runc/internal/container"
)

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all the container",
	Action: func(context *cli.Context) error {
		container.ListContainer()
		return nil
	},
}
