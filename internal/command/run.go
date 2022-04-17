package command

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
		run(tty, cmdArray, resConf)
		return nil
	},
}

func run(tty bool, cmdArray []string, res *subsystems.ResourceConfig) {
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

	mntUrl := "/root/mnt/"
	rootUrl := "/root/"
	container.DeleteWorkSpace(rootUrl, mntUrl)
	os.Exit(0)
}

func sendInitCommand(cmdArray []string, writePipe *os.File) {
	defer writePipe.Close()
	command := strings.Join(cmdArray, " ")
	logrus.Infof("runC send run command; %s", command)
	writePipe.WriteString(command)
}
