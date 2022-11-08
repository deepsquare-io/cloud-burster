package main

import (
	"os"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var flags = []cli.Flag{
}

var app = &cli.App{
	Name:  "cloud-burster",
	Usage: "Burst into the cloud.",
	Flags: flags,
	Action: func(ctx *cli.Context) error {

		return nil
	},
}

func main() {
	if err := app.Run(os.Args); err != nil {
		logger.I.Fatal("app crashed", zap.Error(err))
	}
}
