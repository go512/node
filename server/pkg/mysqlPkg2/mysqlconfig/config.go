package mysqlconfig

import "time"

type Config struct {
	Host        string `json:"host"  toml:"host"`
	Port        int    `json:"port"  toml:"port"`
	User        string `json:"user"  toml:"user"`
	Password    string `json:"password"  toml:"password"`
	DBName      string `json:"db"  toml:"dbname"`
	SSLMode     string `json:"ssl_mode" toml:"ssl_mode"`
	SSLRootCert string `json:"ssl_root_cert" toml:"ssl_root_cert"`

	MaxOpenConns    int           `json:"max_open_conns" toml:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns" toml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_life_time" toml:"conn_max_life_time"`
}
