package kafkaPkg

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"node/conf"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// MultiTopicConsumerManager 多topic消费者管理器
// 通过共享 context 实现统一控制所有消费者
type MultiTopicConsumerManager struct {
	consumers []*Consumer
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex   // 保护 consumers 数组的并发访问
	wg        sync.WaitGroup // 等待所有消费者完全启动
	started   bool           // 标记是否已启动
}

// NewMultiTopicConsumerManager 创建多topic消费者管理器
func NewMultiTopicConsumerManager() *MultiTopicConsumerManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &MultiTopicConsumerManager{
		consumers: make([]*Consumer, 0),
		ctx:       ctx,
		cancel:    cancel,
		started:   false,
	}
}

// AddConsumer 添加一个消费者（订阅一个topic）
// 所有消费者共享同一个 context，可以统一控制
// 注意：必须在 Start() 之前调用
func (m *MultiTopicConsumerManager) AddConsumer(cfg *conf.KafkaConfig, topic, groupID string, concurrency int, handler func(message *kafka.Message) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已启动
	if m.started {
		return fmt.Errorf("无法添加消费者：管理器已启动，请在 Start() 之前添加所有消费者")
	}

	// 参数验证
	if cfg == nil {
		return fmt.Errorf("配置不能为空")
	}
	if topic == "" {
		return fmt.Errorf("topic 不能为空")
	}
	if groupID == "" {
		return fmt.Errorf("groupID 不能为空")
	}

	// 检查是否重复订阅相同的 topic
	for _, c := range m.consumers {
		// 注意：这里简化处理，实际可能需要更复杂的检查
		_ = c // 预留，后续可添加 topic 字段到 Consumer
	}

	consumer, err := NewConsumerWithContext(m.ctx, cfg, topic, groupID, concurrency, handler)
	if err != nil {
		return fmt.Errorf("创建消费者失败: %w", err)
	}

	m.consumers = append(m.consumers, consumer)
	log.Printf("已添加消费者: topic=%s, group=%s, concurrency=%d", topic, groupID, concurrency)
	return nil
}

// Start 启动所有消费者并等待中断信号
func (m *MultiTopicConsumerManager) Start() error {
	m.mu.Lock()
	if m.started {
		m.mu.Unlock()
		return fmt.Errorf("管理器已经启动")
	}
	if len(m.consumers) == 0 {
		m.mu.Unlock()
		return fmt.Errorf("没有添加任何消费者，请先使用 AddConsumer 添加")
	}
	m.started = true
	consumerCount := len(m.consumers)
	m.mu.Unlock()

	log.Printf("启动 %d 个消费者...", consumerCount)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit

	// 统一停止所有消费者
	m.StopAll()
	return nil
}

// StartAsync 异步启动所有消费者（不阻塞）
// 返回一个 channel，当所有消费者都退出时会关闭
func (m *MultiTopicConsumerManager) StartAsync() (<-chan struct{}, error) {
	m.mu.Lock()
	if m.started {
		m.mu.Unlock()
		return nil, fmt.Errorf("管理器已经启动")
	}
	if len(m.consumers) == 0 {
		m.mu.Unlock()
		return nil, fmt.Errorf("没有添加任何消费者，请先使用 AddConsumer 添加")
	}
	m.started = true
	consumerCount := len(m.consumers)
	m.mu.Unlock()

	log.Printf("异步启动 %d 个消费者...", consumerCount)

	// 创建完成通道
	done := make(chan struct{})
	go func() {
		defer close(done)
		<-m.ctx.Done()
		m.waitAllConsumers()
	}()

	return done, nil
}

// StopAll 停止所有消费者
func (m *MultiTopicConsumerManager) StopAll() {
	m.mu.Lock()
	if !m.started {
		m.mu.Unlock()
		log.Println("管理器未启动，无需停止")
		return
	}
	m.mu.Unlock()

	log.Println("开始停止所有消费者...")

	// 取消 context，所有消费者会收到停止信号
	m.cancel()

	// 等待所有消费者停止
	m.waitAllConsumers()

	log.Println("所有消费者已停止")
}

// waitAllConsumers 等待所有消费者的协程退出
func (m *MultiTopicConsumerManager) waitAllConsumers() {
	m.mu.RLock()
	consumers := m.consumers
	m.mu.RUnlock()

	// 等待所有消费者的协程退出
	for i, consumer := range consumers {
		log.Printf("等待消费者 %d/%d 停止...", i+1, len(consumers))
		consumer.wg.Wait() // 等待该消费者的所有协程退出
	}
}

// IsStarted 检查管理器是否已启动
func (m *MultiTopicConsumerManager) IsStarted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.started
}

// ExampleMultiTopicConsumer 示例：如何使用多topic消费者管理器
func ExampleMultiTopicConsumer() {
	cfg := conf.Default()

	// 创建管理器
	manager := NewMultiTopicConsumerManager()

	// 订阅多个 topic，它们共享同一个 context
	err := manager.AddConsumer(cfg, "topic1", "group1", 2, func(msg *kafka.Message) error {
		log.Printf("[Topic1] 收到消息: %s", string(msg.Value))
		return nil
	})
	if err != nil {
		log.Fatal("添加 topic1 消费者失败:", err)
	}

	err = manager.AddConsumer(cfg, "topic2", "group1", 1, func(msg *kafka.Message) error {
		log.Printf("[Topic2] 收到消息: %s", string(msg.Value))
		return nil
	})
	if err != nil {
		log.Fatal("添加 topic2 消费者失败:", err)
	}

	err = manager.AddConsumer(cfg, "topic3", "group1", 3, func(msg *kafka.Message) error {
		log.Printf("[Topic3] 收到消息: %s", string(msg.Value))
		return nil
	})
	if err != nil {
		log.Fatal("添加 topic3 消费者失败:", err)
	}

	// 启动并等待，按 Ctrl+C 时会统一停止所有消费者
	if err := manager.Start(); err != nil {
		log.Fatal("启动管理器失败:", err)
	}
}
