package kafkaPkg

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"log"
	"node/conf"
	"sync"
	"syscall"
	"time"
)

const (
	MaxRetryCount    = 3               // 最大重试次数
	CommitTimeout    = 5 * time.Second // 提交超时时间
	RetryBackoffBase = time.Second     // 重试基础间隔
)

type Consumer struct {
	stop context.CancelFunc
	wg   sync.WaitGroup
}

func NewDialer(cfg *conf.KafkaConfig) (*kafka.Dialer, error) {
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	if cfg.Username != "" {
		sm, err := scram.Mechanism(scram.SHA512, cfg.Username, cfg.Password)
		if err != nil {
			return nil, fmt.Errorf("init scram mechanism failed: %w", err)
		}
		// 生产环境建议禁用不安全的 TLS 配置
		dialer.TLS = &tls.Config{
			InsecureSkipVerify: false, // 测试环境可临时设为 true
			MinVersion:         tls.VersionTLS12,
		}
		dialer.SASLMechanism = sm
	}
	return dialer, nil
}

func Subscribe(ctx context.Context, topic, group string, handler func(msg *kafka.Message) error) {
	NewConsumerWithContext(ctx, conf.Default(), topic, group, 1, func(msg *kafka.Message) error {
		err := handler(msg)
		if err != nil {
			log.Printf("处理消息失败: %v", err)
			return err
		}
		log.Printf("收到消息: topic=%s, partition=%d, offset=%d, key=%s, value=%s",
			msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
		return nil
	})
}

// NewConsumer 可并发启动多个消费者，支持优雅退出
// 注意：同一个topic+group 下多实例 concurrency 总数<= partition 数
func NewConsumer(cfg *conf.KafkaConfig, topic, groupID string, concurrency int, handler func(message *kafka.Message) error) (*Consumer, error) {
	return NewConsumerWithContext(context.Background(), cfg, topic, groupID, concurrency, handler)
}

// NewConsumerWithContext 使用指定的 context 创建消费者，支持共享 context
// 多个消费者可以共享同一个 context，实现统一控制
// 参数 parentCtx: 父级 context，用于外部控制消费者的生命周期
func NewConsumerWithContext(parentCtx context.Context, cfg *conf.KafkaConfig, topic, groupID string, concurrency int, handler func(message *kafka.Message) error) (*Consumer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, errors.New("no kafka brokers")
	}
	if topic == "" {
		return nil, errors.New("no kafka topic")
	}
	if groupID == "" {
		return nil, errors.New("no kafka group id")
	}
	if concurrency <= 0 {
		concurrency = 1
	}

	//2、初始化Dialer （支持SASL认证）
	dialer, err := NewDialer(cfg)
	if err != nil {
		return nil, err
	}

	//3、初始化消费者上下文（支持外部 context 控制）
	ctx, cancel := context.WithCancel(parentCtx)
	c := &Consumer{
		stop: cancel,
	}

	//4、启动多个消费者
	for i := 0; i < concurrency; i++ {
		c.wg.Add(1)
		go func(consumerId int) {
			defer func() {
				c.wg.Done()
				// 捕获 panic,避免单个消费者崩溃导致整个程序退出
				if r := recover(); r != nil {
					log.Printf("[消费者-%d] panic 恢复: %v", consumerId, r)
				}
			}()

			//5、初始化 Reader
			reader := kafka.NewReader(kafka.ReaderConfig{
				Brokers:          cfg.Brokers,
				GroupID:          groupID,
				Topic:            topic,
				Dialer:           dialer,
				MaxAttempts:      3,                // 连接最大尝试次数
				StartOffset:      kafka.LastOffset, // 从最新位置开始消费
				MaxWait:          3 * time.Second,  // 避免无谓等待,提高响应
				ReadBatchTimeout: 3 * time.Second,  // 单次拉取超时
				MinBytes:         1e3,              // 1KB,降低触发批次返回的数据量下限
				MaxBytes:         10e6,             // 10MB,提高单次请求能接受的最大数据量
				CommitInterval:   0,                // 禁用自动提交,手动提交
			})

			defer reader.Close()

			log.Printf("[消费者-%d] 开始消费 topic: %s, group: %s", consumerId, topic, groupID)
			//6、循环拉取消息
			for {
				select {
				case <-ctx.Done(): // 监听退出信号
					log.Printf("[消费者-%d] 收到停止信号,退出消费循环", consumerId)
					return
				default:
					// FetchMessage 会阻塞等待消息,不会导致 CPU 空转
					msg, err := reader.FetchMessage(ctx)
					if err != nil {
						// 检查是否是上下文取消
						if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
							return
						}
						// 忽略临时性错误
						if errors.Is(err, syscall.EAGAIN) {
							continue
						}
						log.Printf("[消费者-%d] 拉取消息失败: %v", consumerId, err)
						continue
					}

					// 处理消息,支持重试
					var handlerErr error
					for retry := 0; retry < MaxRetryCount; retry++ {
						handlerErr = handler(&msg)
						if handlerErr == nil {
							break
						}

						log.Printf("[消费者-%d] 处理消息失败 (重试 %d/%d): %v", consumerId, retry+1, MaxRetryCount, handlerErr)

						// 最后一次重试不需要等待
						if retry < MaxRetryCount-1 {
							time.Sleep(RetryBackoffBase * time.Duration(retry+1))
						}
					}

					// 记录最终处理结果
					if handlerErr != nil {
						log.Printf("[消费者-%d] 消息处理最终失败,已达最大重试次数: offset=%d, partition=%d, error=%v",
							consumerId, msg.Offset, msg.Partition, handlerErr)
					}

					// 提交偏移量 (在消费组模式下有效)
					if err := c.commitMessage(ctx, reader, msg, consumerId); err != nil {
						log.Printf("[消费者-%d] 提交偏移量失败: %v", consumerId, err)
					}
				}
			}
		}(i)
	}

	return c, nil
}

// commitMessage 提交消息偏移量,支持重试
func (c *Consumer) commitMessage(ctx context.Context, reader *kafka.Reader, msg kafka.Message, consumerId int) error {
	commitCtx, commitCancel := context.WithTimeout(context.Background(), CommitTimeout)
	defer commitCancel()

	var err error
	for retry := 0; retry < MaxRetryCount; retry++ {
		err = reader.CommitMessages(commitCtx, msg)
		if err == nil {
			return nil
		}

		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 最后一次重试不需要等待
		if retry < MaxRetryCount-1 {
			log.Printf("[消费者-%d] 提交偏移量失败 (重试 %d/%d): %v", consumerId, retry+1, MaxRetryCount, err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(RetryBackoffBase * time.Duration(retry+1)):
			}
		}
	}

	return fmt.Errorf("提交偏移量失败,已达最大重试次数: %w", err)
}

func (c *Consumer) Stop() {
	log.Println("开始停止所有消费者...")

	// 1. 发送取消信号给所有消费者协程
	c.stop()
	// 3. 等待所有 goroutine 退出
	c.wg.Wait()
	log.Println("所有消费者已成功停止")
}
