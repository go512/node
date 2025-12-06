package kafkaPkg

import "node/conf"

var (
	gProducer *kfProducer
)

func InitKafka(cfg *conf.Config) {
	if cfg == nil || len(cfg.Brokers) == 0 {
		cfg = conf.Default()
	}
	gProducer = NewKfProducer(cfg)
}
