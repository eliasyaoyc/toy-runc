package main

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"toy-runc/internal/command"
)

const Usage = "This is a simple container runtime implementation."

func main() {
	app := cli.NewApp()
	app.Name = "ToyRunC"
	app.Usage = Usage

	app.Commands = command.Commands

	app.Before = func(context *cli.Context) error {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Error(err)
	}
}
