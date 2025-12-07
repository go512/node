package cli

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/urfave/cli/v2"
	"log"
	"node/conf"
	"node/pkg/kafkaPkg"
	"os"
	"os/signal"
	"syscall"
)

func kafkaCommand() *cli.Command {
	return &cli.Command{
		Name:  "kafka_consumer",
		Usage: "kafka cli tool",
		Flags: commonFlags(),
		Action: func(ctx *cli.Context) error {
			log.Println("启动 Kafka 消费者...")
			config, err := conf.Load(ctx.String("config"))
			if err != nil {
				return fmt.Errorf("加载配置文件失败: %w", err)
			}

			log.Printf("配置信息: Brokers=%v, Group=%s", config.Kafka.Brokers, config.Kafka.Group)

			// 初始化 Kafka，传入配置
			kafkaPkg.InitKafka(&config.Kafka)
			log.Println("开始订阅 topic: kafka_topic")
			// 传入空字符串，使用配置文件中的 Group
			kafkaPkg.Subscribe(ctx.Context, "kafka_topic", "group_01", Fun)

			//manager := kafkaPkg.NewMultiTopicConsumerManager()
			//err = manager.Subscribe(&config.Kafka, "kafka_topic", "group_01", Fun)
			//log.Println("xx err %+v", err)
			//manager.Start()
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
			<-quit
			return nil
		},
	}
}

func Fun(msg *kafka.Message) error {
	log.Printf("收到消息: topic=%s, partition=%d, offset=%d, key=%s, value=%s",
		msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
	return nil
}
