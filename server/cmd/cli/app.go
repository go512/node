package cli

import (
	"github.com/urfave/cli/v2"
)

func commonFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "log",
			Usage:   "log level: debug | info | warn | error",
			Value:   "info",
			EnvVars: []string{"LOG"},
		},
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Value:   "config.toml",
		},
	}
}

func NewApp() *cli.App {
	return &cli.App{
		Name:  "kafka-cli",
		Usage: "kafka cli tool",
		Flags: commonFlags(),
		Before: func(c *cli.Context) error {
			return nil
		},
		Commands: []*cli.Command{
			kafkaCommand(),
		},
	}
}
