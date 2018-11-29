package manager

import (
	"github.com/elastic/beats/libbeat/common"
)

type ServerConfig struct {
	Network        string         `config:"network"`
	Address        string         `config:"address"`
	WorkDir        string         `config:"work.dir"`
	RegistryConfig *common.Config `config:"registry"`
}

var DefaultServerConfig = ServerConfig{
	Network: "tcp",
	Address: ":3000",
	WorkDir: "/tmp",
}
