// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"github.com/elastic/beats/libbeat/common"
)

type Config struct {
	Modules []*common.Config `config:"modules"`
	DataDir string           `config:"data.dir"`
	Manager *common.Config   `config:"manager"`
}

var DefaultConfig = Config{}
