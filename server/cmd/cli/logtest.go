package cli

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"node/conf"
	"node/pkg/mysqlPkg"
	"node/pkg/qlog"
)

func logCommand() *cli.Command {
	return &cli.Command{
		Name:  "log_cli",
		Usage: "log cli tool",
		Flags: commonFlags(),
		Action: func(ctx *cli.Context) error {
			log.Println("启动 log 消费者...")

			qlog.Infof("test log")
			config, err := conf.Load(ctx.String("config"))
			if err != nil {
				return fmt.Errorf("加载配置文件失败: %w", err)
			}

			qlog.Infof("config: %+v", config)
			for name, v := range config.Mysql {
				qlog.Infof("mysql: name:=%s  val: %+v", name, v)
			}

			mysqlManager := mysqlPkg.NewManager(config.Mysql)

			db, err := mysqlManager.GetClient("default")
			if err != nil {
				return fmt.Errorf("获取 mysql 客户端失败: %w", err)
			}
			qlog.Infof("db: %+v", db.DB.DryRun)
			return nil
		},
	}
}
