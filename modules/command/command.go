package command

import (
	"github.com/catrossim/monbeat/base"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
)

func init() {
	base.Registry.AddModule("command", New)
}

type CommandModule struct {
	base.BaseModule
	config *CommandsConfig
	done   chan struct{}
	out    chan *common.MapStr
	logger *logp.Logger
}

func (cm *CommandModule) Done() {
	close(cm.done)
	close(cm.out)
}

func (cm *CommandModule) Out() chan *common.MapStr {
	return cm.out
}

func (cm *CommandModule) DoneChannel() chan struct{} {
	return cm.done
}

func New(bm base.BaseModule) (base.Module, error) {
	config := DefaultCommandConfig
	if err := bm.UnpackConfig(config); err != nil {
		return nil, err
	}
	return &CommandModule{
		bm,
		config,
		make(chan struct{}),
		make(chan *common.MapStr),
		logp.NewLogger("command"),
	}, nil
}

func (cm *CommandModule) Monitor() error {
	for _, rawCfg := range cm.config.Cmds {
		cmdCfg := DefaultCmdConfig
		if err := rawCfg.Unpack(cmdCfg); err != nil {
			cm.ErrorChannel() <- err
			continue
		}
		watcher, err := NewCmdWatcher(cmdCfg.Cmd, cmdCfg.Internal, cm.out, cm.logger, cm.ErrorChannel())
		if err != nil {
			cm.logger.Error(err)
			cm.ErrorChannel() <- err
			continue
		}
		go func() {
			err := watcher.Watch(cm.done)
			if err != nil {
				cm.logger.Error(err)
				cm.ErrorChannel() <- err
			}
		}()
	}
	return nil
}
