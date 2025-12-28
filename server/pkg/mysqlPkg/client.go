package mysqlPkg

import (
	"context"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

type Client struct {
	mux    sync.RWMutex
	ctx    context.Context
	DB     *gorm.DB
	value  atomic.Value
	config *Config
}

func NewClient(config *Config) (*Client, error) {
	client := &Client{}

	err := client.Reload(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) load() *Client {
	db, ok := c.value.Load().(*gorm.DB)
	if ok {
		return &Client{
			DB:     db.WithContext(c.ctx),
			value:  atomic.Value{},
			config: c.config,
		}
	}

	return c
}

func (c *Client) Reload(config *Config) (err error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	config.FillWithDefault()

	mycfg, err := config.NewMycfg()
	if err != nil {
		return err
	}

	// 检查配置是否有变化（需要比较新旧配置）
	if c.config != nil {
		oldMycfg, _ := c.config.NewMycfg()
		if reflect.DeepEqual(oldMycfg, mycfg) {
			return nil
		}
	}

	var dialer gorm.Dialector

	switch config.Driver {
	case "mysql":
		dialer = mysql.Open(mycfg.FormatDSN())
	case "postgres":
		dialer = postgres.Open(mycfg.FormatDSN())
	default:
		return fmt.Errorf("unsupported driver: %s", config.Driver)
	}

	logcfg := logger.Config{
		SlowThreshold: 100 * time.Millisecond,
		LogLevel:      logger.Warn,
		Colorful:      false,
	}

	if config.DebugSQL {
		logcfg.LogLevel = logger.Info
	}
	var logface logger.Interface
	logface = logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logcfg)

	db, err := gorm.Open(dialer, &gorm.Config{
		Logger:               logface,
		PrepareStmt:          true,
		DisableAutomaticPing: false,
		AllowGlobalUpdate:    false,
	})
	if err != nil {
		return err
	}

	mydb, err := db.DB()
	if err != nil {
		return err
	}

	err = mydb.Ping()
	if err != nil {
		return err
	}

	if config.MaxIdleConns > 0 {
		mydb.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.MaxOpenConns > 0 {
		mydb.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxLifetime > 0 {
		mydb.SetConnMaxLifetime(time.Duration(config.MaxLifetime) * time.Second)
	}

	if oldDB, ok := c.value.Load().(*gorm.DB); ok {
		oldMycfg, _ := c.config.NewMycfg()
		defer func(old *gorm.DB, dsn string) {
			oldMydb, err := old.DB()
			if err == nil {
				logger.Default.Error(context.Background(), fmt.Sprintf("%T.Close(%s): %+v", old, dsn, oldMydb.Close()))
			}
		}(oldDB, oldMycfg.FormatDSN())
	}

	c.value.Store(db)
	c.config = config
	c.config.mycfg = mycfg

	//registerTraceCallbacks(c)
	//registerMetricsCallbacks(c)

	return nil
}
