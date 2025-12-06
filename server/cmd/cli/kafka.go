package cli

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"node/pkg/kafkaPkg"
)

func kafkaCommand() *cli.Command {
	return &cli.Command{
		Name:  "kafka_consumer",
		Usage: "kafka cli tool",
		Flags: commonFlags(),
		Action: func(c *cli.Context) error {
			kafkaPkg.InitKafka(nil)
			fmt.Println("kafka_consumer")
			return nil
		},
	}
}
