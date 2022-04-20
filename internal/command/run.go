package command

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"strings"
	"toy-runc/internal/cgroups/subsystems"
	"toy-runc/internal/container"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: "Create a container with namespace and cgroups limit docker run -it [command]",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "it",
			Usage: "enable tty",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
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
		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
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
		detach := context.Bool("d")

		if tty && detach {
			return fmt.Errorf("it and d paramter can not both provided")
		}

		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuShare:    context.String("cpushare"),
			CpuSet:      context.String("cpuset"),
		}
		containerName := context.String("name")
		run(tty, cmdArray, resConf, containerName)
		return nil
	},
}

func run(tty bool, cmdArray []string, res *subsystems.ResourceConfig, containerName string) {
	parent, writePipe := container.NewParentProcess(tty)
	if err := parent.Start(); err != nil {
		logrus.Error(err)
	}

	containerName, err := container.RecordContainerInfo(parent.Process.Pid, cmdArray, containerName)
	if err != nil {
		logrus.Errorf("record container info error; %v", err)
		return
	}

	//cgroupManager := cgroups.NewCgroupManager("toyRunC-cgroup")
	//defer cgroupManager.Destroy()
	//cgroupManager.Set(res)
	//cgroupManager.Apply(parent.Process.Pid)

	sendInitCommand(cmdArray, writePipe)
	if tty {
		parent.Wait()
		container.DeleteContainerInfo(containerName)
		mntURL := "/root/mnt/"
		rootURL := "/root/"
		container.DeleteWorkSpace(rootURL, mntURL)
	}
}

func sendInitCommand(cmdArray []string, writePipe *os.File) {
	defer writePipe.Close()
	command := strings.Join(cmdArray, " ")
	logrus.Infof("runC send run command; %s", command)
	writePipe.WriteString(command)
}
