package command

import (
	"time"

	"github.com/elastic/beats/libbeat/common"
)

type CommandsConfig struct {
	Cmds []*common.Config `config:"cmds"`
}

type CmdConfig struct {
	Cmd      string        `config:"cmd"`
	Internal time.Duration `config:"internal"`
}

var DefaultCommandConfig = &CommandsConfig{}

var DefaultCmdConfig = &CmdConfig{
	Internal: 30 * time.Minute,
}
