package cli

import (
	"syslogmonitor/conf"
	"sync"
)

var once sync.Once

func Init() {
	once.Do(func() {
		conf.Init()
	})
}

