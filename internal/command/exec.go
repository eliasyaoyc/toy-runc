package command

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"toy-runc/internal/container"
	_ "toy-runc/internal/nsenter"
)

var execCommand = cli.Command{
	Name:  "exec",
	Usage: "exec a command into container",
	Action: func(context *cli.Context) error {
		if os.Getenv(container.ENV_EXEC_PID) != "" {
			logrus.Infof("pid callback pid %v", os.Getgid())
			return nil
		}
		if len(context.Args()) < 2 {
			return fmt.Errorf("missing contaienr name or command")
		}
		containerName := context.Args().Get(0)
		var commandArray []string
		for _, arg := range context.Args().Tail() {
			commandArray = append(commandArray, arg)
		}
		container.ExecContainer(containerName, commandArray)
		return nil
	},
}
