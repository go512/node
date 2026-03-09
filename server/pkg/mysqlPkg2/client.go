package mysqlPkg2

import (
	"gorm.io/gorm"
	"node/pkg/mysqlPkg2/mysqlconfig"
)

type Database struct {
	cfg    *mysqlconfig.Config
	gormDb *gorm.DB
}

func (p *Database) Open() (err error) {
	mys
}
