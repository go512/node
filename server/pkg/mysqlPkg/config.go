package mysqlPkg

import (
	"github.com/go-sql-driver/mysql"
	"strings"
	"time"
)

type Config struct {
	Driver               string        `yaml:"driver" toml:"driver"`
	DSN                  string        `yaml:"dsn" toml:"dsn"`
	DialTimeout          time.Duration `yaml:"dial_timeout" toml:"dial_timeout"`                       // 连接超时时间，默认 1000ms
	ReadTimeout          time.Duration `yaml:"read_timeout" toml:"read_timeout"`                       // socket 读超时时间，默认 3000ms
	WriteTimeout         time.Duration `yaml:"write_timeout" toml:"write_timeout"`                     // socket 写超时时间，默认 3000ms
	MaxOpenConns         int           `yaml:"max_open_conns" toml:"max_open_conns"`                   // 最大连接数，默认 200
	MaxIdleConns         int           `yaml:"max_idle_conns" toml:"max_idle_conns"`                   // 最大空闲连接数，默认 80
	MaxLifetime          int           `yaml:"max_life_time" toml:"max_life_time"`                     // 空闲连接最大存活时间，默认 600s
	TraceIncludeNotFound bool          `yaml:"trace_include_not_found" toml:"trace_include_not_found"` // 是否将NotFound error作为错误记录在trace中，默认为否

	//internal
	mycfg *mysql.Config `yaml:"mycfg" toml:"mycfg"`
}

func (c *Config) FillWithDefault() {
	if c == nil {
		return
	}

	if c.Driver == "" {
		c.Driver = "mysql"
	}

	if c.DialTimeout <= 0 {
		c.DialTimeout = 1000 * time.Millisecond
	}

	if c.ReadTimeout <= 0 {
		c.ReadTimeout = 3000 * time.Millisecond
	}

	if c.WriteTimeout <= 0 {
		c.WriteTimeout = 3000 * time.Millisecond
	}

	if c.MaxOpenConns <= 0 {
		c.MaxOpenConns = 200
	}

	if c.MaxIdleConns <= 0 {
		c.MaxIdleConns = 80
	}

	if c.MaxLifetime <= 0 {
		c.MaxLifetime = 600
	}
}

func (c *Config) NewMycfg() (dsn *mysql.Config, err error) {
	if c.DSN != "" {
		dsn, err = mysql.ParseDSN(c.DSN)
		if err != nil {
			return
		}
		if dsn.Timeout <= 0 {
			dsn.Timeout = c.DialTimeout * time.Millisecond
		}
		if dsn.ReadTimeout <= 0 {
			dsn.ReadTimeout = c.ReadTimeout * time.Millisecond
		}
		if dsn.WriteTimeout <= 0 {
			dsn.WriteTimeout = c.WriteTimeout * time.Millisecond
		}

		c.DSN = dsn.FormatDSN()
		return dsn, nil
	}

	return
}

func (c *Config) NewWithDB(dbName string) (*Config, error) {
	mycfg, err := mysql.ParseDSN(c.DSN)
	if err != nil {
		return nil, err
	}
	mycfg.DBName = dbName

	copied := *c
	copied.DSN = mycfg.FormatDSN()
	copied.mycfg = mycfg
	return &copied, nil

}

func (c *Config) IsEqualDB(dbname string) bool {
	dsn, err := mysql.ParseDSN(c.DSN)
	if err != nil {
		return false
	}

	return strings.Compare(dsn.DBName, dbname) == 0
}

type ManagerConfig map[string]*Config
