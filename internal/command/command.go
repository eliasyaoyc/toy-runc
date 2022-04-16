package command

import "github.com/urfave/cli"

var AllSupportCommands []cli.Command

func init() {
	AllSupportCommands = append(AllSupportCommands, initCommand, runCommand)
}
