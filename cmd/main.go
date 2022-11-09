package main

import (
	"os"

	"github.com/squarefactory/cloud-burster/cmd/create"
	"github.com/squarefactory/cloud-burster/cmd/delete"
	"github.com/squarefactory/cloud-burster/logger"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var flags = []cli.Flag{}

var app = &cli.App{
	Name:  "cloud-burster",
	Usage: "Burst into the cloud.",
	Flags: flags,
	Commands: []*cli.Command{
		create.Command,
		delete.Command,
	},
}

func main() {
	if err := app.Run(os.Args); err != nil {
		logger.I.Fatal("app crashed", zap.Error(err))
	}
}
