package cli

import (
	"github.com/urfave/cli/v2"
	"log"
	"node/pkg/qlog"
)

func logCommand() *cli.Command {
	return &cli.Command{
		Name:  "log_cli",
		Usage: "log cli tool",
		Flags: commonFlags(),
		Action: func(ctx *cli.Context) error {
			log.Println("启动 log 消费者...")

			logg := qlog.New()
			logg.Debug("test log")
			return nil
		},
	}
}
