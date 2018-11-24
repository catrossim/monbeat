package base

import (
	"time"

	"github.com/elastic/beats/libbeat/common"
)

type ModuleConfig struct {
	Period  time.Duration `config:"period"`
	Enabled bool          `config:"enabled"`
	Timeout time.Duration `config:"timeout"`
	Module  string        `config:"module"`
}

var defaultModuleConfig = ModuleConfig{
	Enabled: true,
	Period:  time.Second * 10,
}

func DefaultModuleConfig() ModuleConfig {
	return defaultModuleConfig
}

type Module interface {
	Name() string
	Config() ModuleConfig
	UnpackConfig(to interface{}) error
	ErrorChannel() chan error
}

type BaseModule struct {
	name      string
	config    ModuleConfig
	rawConfig *common.Config
	errorChan chan error
}

func (bm *BaseModule) Name() string {
	return bm.name
}

func (bm *BaseModule) Config() ModuleConfig {
	return bm.config
}

func (bm *BaseModule) UnpackConfig(to interface{}) error {
	return bm.rawConfig.Unpack(to)
}

func (bm *BaseModule) ErrorChannel() chan error {
	return bm.errorChan
}

type Monitor interface {
	Monitor() error
	Out() chan *common.MapStr
	DoneChannel() chan struct{}
	Done()
}
