package conf

type Config struct {
	Brokers []string `json:"brokers"`

	// stl
	Username   string `json:"username"`
	Password   string `json:"password"`
	ServerName string `json:"server_name" toml:"server_name"`

	Partition   int `json:"partition"`
	Replication int `json:"replication"`

	// apollo
	Group string `json:"group" toml:"group"`

	AutoCreateTopic bool `json:"auto_create_topic"` //默认在broker 开启自动创建，否则需要手动创建
}

func Default() *Config {
	return &Config{
		Brokers: []string{"localhost:9092"},

		Partition:   1,
		Replication: 1,

		Group:           "group_01",
		AutoCreateTopic: false,
	}
}
