package conf

import "github.com/BurntSushi/toml"

var gConfig *Config

type MysqlDBNode struct {
	Host     string `json:"host" toml:"host"`
	Port     int    `json:"port" toml:"prot"`
	User     string `json:"user" toml:"user"`
	Password string `json:"password" toml:"password"`
	DBName   string `json:"db_name" toml:"db_name"`
	MaxOpen  int    `json:"max_open" toml:"max_open"`
	MaxIdle  int    `json:"max_idle" toml:"max_idle"`
}
type DatabaseConfig struct {
	Event MysqlDBNode `json:"event" toml:"event" yaml:"event"`
}
type Config struct {
	Kafka KafkaConfig    `json:"kafka" toml:"kafka" yaml:"kafka"`
	Mysql DatabaseConfig `json:"mysql" toml:"mysql" yaml:"mysql"`
}

func Load(configPath string) (cfg *Config, err error) {
	gConfig = &Config{}
	_, err = toml.DecodeFile(configPath, &gConfig)

	return gConfig, err
}
