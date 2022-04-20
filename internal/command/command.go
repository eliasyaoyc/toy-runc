package command

import "github.com/urfave/cli"

var Commands []cli.Command

func init() {
	Commands = append(
		Commands,
		initCommand,
		runCommand,
		commitCommand,
		listCommand,
		logCommand,
		execCommand,
		stopCommand,
		removeCommand,
	)
}
