package registry

import (
	"time"
)

type ConsulConfig struct {
	Address                        string        `config:"address"`
	Service                        string        `config:"service"`
	Tag                            []string      `config:"tags"`
	DeregisterCriticalServiceAfter time.Duration `config:"deregister.after"`
	Interval                       time.Duration `config:"interval"`
}

var DefaultConsulConfig = ConsulConfig{
	Address: "127.0.0.1:8500",
	Service: "default",
	Tag:     []string{"default"},
	DeregisterCriticalServiceAfter: 1 * time.Minute,
	Interval:                       10 * time.Second,
}
