# node

```mermaid
flowchart TD
    A(开始) --> B(结束)
    
```

```mermaid
flowchart TD
start --> connect[连接到AMQP并创建信道] -.-> nofityConn[监听msgs通道状态]  -.关闭.-> connect --> declare[声明队列,绑定交换机] --> recoverall[恢复全部product] 
declare --> subscribe[从msgs通道获取消息]
subgraph loop
    subscribe -.alive事件.-> analyze -.in订阅的product且subscribed=0.-> recoveryOne[恢复单个product]
    analyze --> publishalive[投递到kafka.alive]
    subscribe -.complete事件.-> remark[记录恢复完成时间,\n目前仅记日志]
    subscribe -.其余事件.-> publish[投递到kafka对应事件topic]
end
```


```shell
cd server && go run .
```

###deamon
```shell
(cd ./server/cmd && go run . kafka_consumer -c ./../config.toml)
```