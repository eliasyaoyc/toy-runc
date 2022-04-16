package main

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"strings"
	"toy-runc/internal/cgroups"
	"toy-runc/internal/cgroups/subsystems"
	"toy-runc/internal/container"
)

const Usage = "This is a simple container runtime implementation."

func main() {
	app := cli.NewApp()
	app.Name = "ToyRunC"
	app.Usage = Usage

	app.Commands = []cli.Command{
		initCommand,
		runCommand,
	}

	app.Before = func(context *cli.Context) error {
		logrus.SetFormatter(&logrus.JSONFormatter{})

		logrus.SetOutput(os.Stdout)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Error(err)
	}
}

var runCommand = cli.Command{
	Name:  "run",
	Usage: "Create a container with namespace and cgroups limit docker run -it [command]",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "it",
			Usage: "enable tty",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
	},

	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return errors.New("missing container command")
		}
		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		tty := context.Bool("it")
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuShare:    context.String("cpushare"),
			CpuSet:      context.String("cpuset"),
		}
		Run(tty, cmdArray, resConf)
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(context *cli.Context) error {
		logrus.Info("runC init.")
		err := container.RunContainerInitProcess()
		return err
	},
}

func Run(tty bool, cmdArray []string, res *subsystems.ResourceConfig) {
	parent, writePipe := container.NewParentProcess(tty)
	if err := parent.Start(); err != nil {
		logrus.Error(err)
	}
	manager := cgroups.NewCgroupManager("toyRunC-cgroup")
	defer manager.Destroy()
	manager.Set(res)
	manager.Apply(parent.Process.Pid)

	sendInitCommand(cmdArray, writePipe)
	parent.Wait()
}

func sendInitCommand(cmdArray []string, writePipe *os.File) {
	command := strings.Join(cmdArray, " ")
	logrus.Infof("runC send run command; %s", command)
	writePipe.WriteString(command)
}
