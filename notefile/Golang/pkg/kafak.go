package pkg

import (
	"context"
	"net"
	"time"
)

//"github.com/segmentio/kafka-go"用于连接和操作kafka
/**
"github.com/segmentio/kafka-go/sasl/scram"提供了sasl认证方法
 Writer 可并发安全调用，配置在首次使用后不应该在修改
*/

type Writer struct {
	Addr  net.Addr //使用kafka.TCP("127.0.0.1:9092")
	Topic string
	//Balance                       //负载均衡 默认round-robin
	MaxAttempts     int           // The default is to try at most 10 times.
	WriteBackoffMin time.Duration // Default: 100ms
	WriteBackoffMax time.Duration // Default: 1s
	BatchSize       int           // default 100 messages.
	BatchBytes      int64         // default 1048576
	//RequiredAcks RequiredAcks // Defaults to RequireNone. 可设置为RequireOne、RequireAll
	//Transport RoundTripper // If nil, DefaultTransport is used.
}

type Message struct{}

/*
*

	WriteMessages writes a batch of messages to the topic.

除非写入器被配置为异步写入消息，否则该方法将阻塞，直到所有消息都写入成功或失败。
*/
func (w *Writer) WriteMessages(ctx context.Context, msgs ...Message) error {
	return nil
}

/**
*
	Reader
*/

type ReaderConfig struct {
	Broker      []string
	GropID      string // 与Partition互斥，设置后需搭配Topic（单主题）或 GroupTopics（多主题）
	GroupTopics []string
	Topic       string
	Partition   int // 必须与Topic搭配使用
	//Dialer *Dialer
	QueueCapacity    int           // 缓存消费消息，defaults 100
	MinBytes         int           // 单次拉取的最小字节数，Default: 1
	MaxBytes         int           //单次拉取的最大字节数，需覆盖最大单条消息大小，Default: 1MB
	MaxWait          time.Duration // 拉取消息的最大等待时间，Default: 10s
	ReadBatchTimeout time.Duration // 从内部批量消息中读取的超时时间，Default: 10s
	ReadLagInterval  time.Duration // 消费延迟（lag）的更新频率，设为负数则禁用延迟上报
	//GroupBalancers []GroupBalancer // Default: [Range, RoundRobin], Only used when GroupID is set
	CommitInterval         time.Duration // Default: 0, Only used when GroupID is set
	PartitionWatchInterval time.Duration // Default: 5s, Only used when GroupID is set and WatchPartitionChanges is set.
}

type Reader struct {
	Config ReaderConfig
}

// 如果使用了消费者组，ReadMessage 会在被调用时自动提交消息的偏移量。
// 请注意，这可能会导致偏移量在消息被完全处理之前就被提交。
//
// 如果需要更精细地控制偏移量提交的时机，建议改用 FetchMessage 配合 CommitMessages 方法
// 自动提交：优点不需要手动管理偏移量，缺点可能存在偏移量提交时机问题 可能丢消息，比如处理失败，但已经提交了
func (r *Reader) ReadMessage(ctx context.Context) (Message, error) {
	return Message{}, nil
}

// 不会自动提交偏移量，需要手动提交偏移量，
func (r *Reader) FetchMesage(ctx context.Context) (Message, error) {
	return Message{}, nil
}

func (r *Reader) CommitMessages(ctx context.Context, msgs ...Message) error {
	return nil
}

type Client struct{}

/*client.OffsetFetch 获取指定主题的偏移量
Dialer结构体封装了kafka连接的创建逻辑，包级别定义了默认的
连接到指定地址
func (d *Dialer) Dial(network string, address string) (*Conn, error) //

*/
