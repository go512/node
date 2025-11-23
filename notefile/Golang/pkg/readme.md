# kafka

| 维度     | Broker                      | Topic                              | Partition                     |
|----------|-----------------------------|------------------------------------|-------------------------------|
| 本质     | 物理服务器节点（服务载体）        | 逻辑消息分类（业务标签）           | 物理数据分片（存储单元）      |
| 作用     | 提供 Kafka 服务，存储数据、处理请求| 隔离不同业务的消息流               | 实现并行处理、存储具体消息数据 |
| 数据存储 | 存储多个 Topic 的 Partition 数据 | 不存储数据，仅作为逻辑分类         | 存储 Topic 的部分消息数据      |
| 数量关系 | 集群包含 N 个 Broker             | 集群可创建 M 个 Topic              | 单个 Topic 包含 K 个 Partition |
| 示例类比 | 2 个 Broker（Broker-0、Broker-1） | 不同类型的快递业务（如生鲜、普通） | 每个业务下的快递分拣区域      |

### Broker
1、kafka的节点和服务器数 接受生产者消息写入，将消息持久化到磁盘
2、管理Partition得副本（Leader/Follower），参与集群的故障转移
### Topic
1、kafka的逻辑消息分类，每个Topic下可以有多个Partition，每个Partition下可以有多个副本（Leader/Follower）
2、不同业务的分类如：如订单消息、日志消息、赔率消息）
3、Topic下Partition的副本数，可以配置多个，但至少一个，默认为1
### Partition
1、单个 Partition 内的消息严格按发送顺序存储（FIFO），但跨 Partition 不保证全局顺序；
2、每个 Partition 对应磁盘上的一个日志文件，消息以追加方式写入
3、Partition 数量决定了 Topic 的最大并行处理能力（消费者数量 ≤ Partition 数量）。