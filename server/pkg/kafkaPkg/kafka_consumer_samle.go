package kafkaPkg

import (
	"context"
	"crypto/tls"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"log"
	"sync"
	"time"
)

func ExampleClient(brokers []string, username, password string) {
	client := &kafka.Client{
		Addr:    kafka.TCP(brokers...),
		Timeout: 10 * time.Second,
	}
	if username != "" {
		sm, err := scram.Mechanism(scram.SHA512, username, password)
		if err != nil {
			panic(err)
		}
		client.Transport = &kafka.Transport{
			TLS: &tls.Config{
				InsecureSkipVerify: false,
			},
			SASL: sm,
		}
	}

	createTopicResp, err := client.CreateTopics(context.Background(), &kafka.CreateTopicsRequest{
		Addr: nil, //若未指定优先使用，否则使用client.Addr
		Topics: []kafka.TopicConfig{
			{
				Topic:             "test",
				NumPartitions:     1,
				ReplicationFactor: 1,
			},
			{
				Topic:             "test2",
				NumPartitions:     1,
				ReplicationFactor: 1,
			},
		},
		ValidateOnly: false, //是否只验证不创建
	})

	if err != nil {
		panic(err)
	}
	log.Printf("createTopicResp: %+v", createTopicResp)

	deleteTopicResp, err := client.DeleteTopics(context.Background(), &kafka.DeleteTopicsRequest{
		Topics: []string{"test", "test2"},
	})

	if err != nil {
		panic(err)
	}
	log.Printf("deleteTopicResp: %+v", deleteTopicResp)
}

// --------------------------例子
type Config struct {
	Brokers            []string
	Username, Password string
}

// NewProducer 创建通用生产者，多数场景优先使用单个Writer处理多Topic
func NewProducer(cfg *Config) *kafka.Writer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireOne,
	}
	if cfg.Username != "" {
		sm, err := scram.Mechanism(scram.SHA512, cfg.Username, cfg.Password)
		if err != nil {
			log.Fatal("error: ", err)
		}
		w.Transport = &kafka.Transport{
			TLS:  &tls.Config{},
			SASL: sm,
		}
	}
	return w
}

type ConsumerVV struct {
	stop context.CancelFunc
	wg   sync.WaitGroup
}

func NewConsumerVV(cfg *Config, topic string, groupID string, concurrency int, handler func(msg *kafka.Message)) *ConsumerVV {
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	if cfg.Username != "" {
		sm, err := scram.Mechanism(scram.SHA512, cfg.Username, cfg.Password)
		if err != nil {
			log.Fatal("error: ", err)
		}
		dialer.TLS = &tls.Config{}
		dialer.SASLMechanism = sm
	}

	ctx, cannel := context.WithCancel(context.Background())
	c := &ConsumerVV{stop: cannel}
	for range concurrency {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			r := kafka.NewReader(kafka.ReaderConfig{
				Brokers: cfg.Brokers,
				GroupID: groupID,
				Topic:   topic,
				Dialer:  dialer,
			})

			for {
				select {
				case <-ctx.Done():
					return
				default:
					msg, err := r.ReadMessage(ctx)
					if err != nil {
						log.Printf("error: %v", err)
						continue
					}
					handler(&msg)
					if groupID != "" {
						r.CommitMessages(ctx, msg)
					}
				}
			}
		}()
	}
	return c
}

func (c *ConsumerVV) Stop() {
	c.stop()
	c.wg.Wait()
}

func MapToHanders(m map[string]string) []kafka.Header {
	headers := make([]kafka.Header, 0, len(m))
	for k, v := range m {
		headers = append(headers, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}
	return headers
}

func testx() {
	proucer := NewProducer(&Config{
		Brokers:  []string{"127.0.0.1:9092"},
		Username: "admin",
		Password: "<PASSWORD>",
	})
	defer proucer.Close()

	err := proucer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte("key"),
		Value: []byte("value"),
		Headers: MapToHanders(map[string]string{
			"key": "value",
		}),
	}, kafka.Message{
		Key:   []byte("key"),
		Value: []byte("value"),
		Headers: MapToHanders(map[string]string{
			"key": "value",
		}),
	})
	if err != nil {
		log.Printf("error: %v", err)
	}

	consumer := NewConsumerVV(&Config{
		Brokers:  []string{"127.0.0.1:9092"},
		Username: "admin",
		Password: "<PASSWORD>",
	}, "test", "test", 1, func(msg *kafka.Message) {
		log.Printf("msg: %+v", msg)
	})
	defer consumer.Stop()
}
