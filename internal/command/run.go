package command

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"strconv"
	"strings"
	"toy-runc/internal/cgroups"
	"toy-runc/internal/cgroups/subsystems"
	"toy-runc/internal/container"
	"toy-runc/internal/network"
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
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
		cli.StringSliceFlag{
			Name:  "e",
			Usage: "set environment",
		},
		cli.StringFlag{
			Name:  "net",
			Usage: "container network",
		},
		cli.StringSliceFlag{
			Name:  "p",
			Usage: "port mapping",
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
		imageName := cmdArray[0]
		cmdArray = cmdArray[1:]

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
		volume := context.String("v")
		network := context.String("net")
		portMapping := context.StringSlice("p")
		envSlice := context.StringSlice("e")

		run(tty, cmdArray, resConf, containerName, volume, imageName, envSlice, network, portMapping)
		return nil
	},
}

func run(tty bool, cmdArray []string, res *subsystems.ResourceConfig, containerName, volume, imageName string, envSlice []string,
	nw string, portMapping []string) {
	containerID := container.RandStringBytes(10)

	if containerName == "" {
		containerName = containerID
	}

	parent, writePipe := container.NewParentProcess(tty, containerName, volume, imageName, envSlice)
	if err := parent.Start(); err != nil {
		logrus.Error(err)
	}

	containerName, err := container.RecordContainerInfo(parent.Process.Pid, cmdArray, containerName, containerID, volume)
	if err != nil {
		logrus.Errorf("record container info error; %v", err)
		return
	}

	cgroupManager := cgroups.NewCgroupManager("toyRunC-cgroup")
	defer cgroupManager.Destroy()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)

	if nw != "" {
		network.Init()
		containerInfo := &container.ContainerInfo{
			Pid:         strconv.Itoa(parent.Process.Pid),
			Id:          containerID,
			Name:        containerName,
			PortMapping: portMapping,
		}
		if err := network.Connect(nw, containerInfo); err != nil {
			logrus.Errorf("error connect network; %v", err)
			return
		}
	}
	sendInitCommand(cmdArray, writePipe)

	if tty {
		parent.Wait()
		container.DeleteContainerInfo(containerName)
		container.DeleteWorkSpace(volume, containerName)
	}
}

func sendInitCommand(cmdArray []string, writePipe *os.File) {
	defer writePipe.Close()
	command := strings.Join(cmdArray, " ")
	logrus.Infof("runC send run command; %s", command)
	writePipe.WriteString(command)
}
