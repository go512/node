package mysql

import "sync"

var (
	nameM sync.Map
	once  sync.Once
)
