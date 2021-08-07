package local

import "sync"

type Local struct {
	Ip       string
	HostName string
}

var local *Local
var once sync.Once

func GetLocal() *Local {
	once.Do(func() {
		local = &Local{}
	})
	return local
}
