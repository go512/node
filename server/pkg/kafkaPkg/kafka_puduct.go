package kafkaPkg

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"log"
	"node/conf"
	"sync"
	"time"
)

type kfProducer struct {
	cfg     conf.KafkaConfig
	writers map[string]*kafka.Writer
	mu      sync.Mutex
}

func NewKfProducer(cfg *conf.KafkaConfig) *kfProducer {
	return &kfProducer{
		cfg:     *cfg,
		writers: make(map[string]*kafka.Writer),
	}
}

func NewKafkaWriter(topic string, cfg *conf.KafkaConfig) (w *kafka.Writer, err error) {
	fmt.Println("NewKafkaWriter", cfg)
	//可选设置回调函数
	completionFunc := func(msgs []kafka.Message, err error) {
		if err != nil {
			return
		}
		latestOffset := msgs[len(msgs)-1].Offset
		log.Printf("投递成功，最新 Offset: %d, 消息数: %d", latestOffset, len(msgs))
	}
	w = &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Balancer: &kafka.Hash{},
		Topic:    topic,
		//&kafka.LeastBytes{},:每次发送消息时，选择当前字节数最少的分区投递
		// &kafka.Hash{}  hash(key) % len(partitions)
		//kafka.RoundRobin{} - 轮询分区策略 partition0 -> 1 -> 2 -> 0 -> 1 -> 2
		//kafka.Manual{} 手动指定
		RequiredAcks: kafka.RequireOne,
		Async:        true,            //默认false，（false同步发送，writeMessage会堵塞，直到broker返回Ack确认或返回超时，true异步，直接返回nil）
		Completion:   completionFunc,  //发送完成回调
		MaxAttempts:  5,               // 自定义最大重试次数为 5 次（默认 10 次）
		WriteTimeout: 3 * time.Second, // 单次发送超时时间
	}

	if cfg.Username != "" {
		sm, err := scram.Mechanism(scram.SHA512, cfg.Username, cfg.Password)
		if err != nil {
			return nil, err
		}
		w.Transport = &kafka.Transport{
			TLS: &tls.Config{
				InsecureSkipVerify: false,
				ServerName:         cfg.ServerName, //证书域名
			},
			SASL: sm,
		}
	}

	return w, nil
}

/**
 * 获取kafka生产者， 当Broker端 auto.create.topics.enable=true （默认值）
 *  生产者向不存在的topic发送消息的，broker会自动创建topic，并分配分区，然后写入数据
 *  如果是false：生产者向不存在的topic发送消息，broker会返回错误：kafka: unknown topic or partition
 */
func (p *kfProducer) getWriter(topic string) (w *kafka.Writer, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	w, ok := p.writers[topic]
	if ok {
		return w, nil
	}

	//只有在Broker端 auto.create.topics.enable=false，尽量通过配置创建，而非代码
	if !p.cfg.AutoCreateTopic {
		if err := p.createTopic(context.Background(), topic); err != nil {
			return nil, err
		}
	}

	w, err = NewKafkaWriter(topic, &p.cfg)
	if err != nil {
		err = fmt.Errorf("NewKafkaWriter err:%v", err)
		return nil, err
	}
	p.writers[topic] = w
	return w, nil
}

/**
 * 创建topic
 * partition: 分区数 只能增加，不能减少，(若需减少，需要重建Topic) （建议每个broker承载100-200个分区）
 * replication: 副本数（不能超过集群的Broker节点数，例如：2个broker最多设置2个）
 * 每个消费者最多消费一个分区，默认轮询，（分区数决定了消费组的最大并发数）
 */
func (p *kfProducer) createTopic(ctx context.Context, topic string) (err error) {
	dialer, err := NewKafkaDialer(&p.cfg)
	//用kafka.conn 连接broker
	if err != nil {
		err = fmt.Errorf("NewKafkaDialer err:%w", err)
		return err
	}

	conn, err := dialer.DialContext(context.Background(), "tcp", p.cfg.Brokers[0])
	if err != nil {
		err = fmt.Errorf("kafka.DialContext err:%w", err)
		return err
	}
	defer conn.Close()

	//读取集群中所有分区信息，检查topic 是否存在
	partitions, err := conn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("ReadPartitions err:%w", err)
	}

	var exist bool
	for _, partition := range partitions {
		if partition.Topic == topic {
			exist = true
			break
		}
	}
	//存在直接退出
	if exist {
		return
	}

	//topic 不存在，获取controller broker
	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("get controller broker err:%w", err)
	}
	controllerAddr := fmt.Sprintf("%s:%d", controller.Host, controller.Port)
	controllerConn, err := dialer.DialContext(context.Background(), "tcp", controllerAddr)
	if err != nil {
		return fmt.Errorf("dial controller broker err:%w", err)
	}
	defer controllerConn.Close()

	fmt.Println(fmt.Sprintf("creat topic: %v,partition: %v,replication: %v", topic, p.cfg.Partition, p.cfg.Replication))

	if err = controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     p.cfg.Partition,
		ReplicationFactor: p.cfg.Replication,
	}); err != nil {
		return fmt.Errorf("CreateTopics err:%w", err)
	}

	return
}

func (p *kfProducer) publish(topic string, key, value []byte, headers []kafka.Header) error {
	w, err := p.getWriter(topic)
	if err != nil {
		err = fmt.Errorf("get kafka writer err:%w", err)
		return err
	}

	msg := kafka.Message{
		Key:     key,
		Value:   value,
		Headers: headers,
	}
	err = w.WriteMessages(context.Background(), msg)
	if err != nil {
		err = fmt.Errorf("write message err:%w", err)
	}
	return err
}

// 其实这里不需要进行手动重试，因为kafka会自动重试
func PublishRetry(topic string, key, value []byte, headers []kafka.Header, retry int) (err error) {
	for i := 0; i < retry; i++ {
		if i > 0 {
			time.Sleep(500 * time.Millisecond * time.Duration(i))
		}
		err = gProducer.publish(topic, key, value, headers)
		if err == nil {
			return
		}
		fmt.Sprintln(fmt.Sprintf("PublishRetry topic:%s retry:%d error:%v", topic, i, err))
	}
	err = fmt.Errorf("try max but failed")
	return
}

func Publish(topic string, key, value []byte, headers []kafka.Header) error {
	return gProducer.publish(topic, key, value, headers)
}

func NewKafkaDialer(cfg *conf.KafkaConfig) (*kafka.Dialer, error) {
	var err error
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}
	if cfg.Username != "" {
		dialer.SASLMechanism, err = scram.Mechanism(scram.SHA512, cfg.Username, cfg.Password)
		if err != nil {
			return nil, err
		}
		dialer.TLS = &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         cfg.ServerName,
		}
	}
	return dialer, nil
}
