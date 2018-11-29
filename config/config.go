// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"github.com/elastic/beats/libbeat/common"
)

type Config struct {
	Modules       []*common.Config `config:"modules"`
	ConfigModules *common.Config   `config:"config.modules"`
	DataDir       string           `config:"config.data.dir"`
	Manager       *common.Config   `config:"config.manager"`
}

var DefaultConfig = Config{}
