package beater

import (
	"fmt"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/catrossim/monbeat/base"
	"github.com/catrossim/monbeat/config"
	"github.com/catrossim/monbeat/manager"
)

// Monbeat configuration.
type Monbeat struct {
	done    chan struct{}
	modules []*base.ModuleWrapper
	config  config.Config
	client  beat.Client
}

// New creates an instance of monbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}
	logp.Info("Read inputs. num: %d", len(c.Modules))
	bt := &Monbeat{
		done:   make(chan struct{}),
		config: c,
	}
	for _, mCfg := range c.Modules {
		module, err := base.NewModule(mCfg, base.Registry)
		if err != nil {
			logp.Error(err)
			continue
		}
		bt.modules = append(bt.modules, module)
	}

	return bt, nil
}

// Run starts monbeat.
func (bt *Monbeat) Run(b *beat.Beat) error {
	logp.Info("Monbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	modules := bt.modules
	for _, mod := range modules {
		runner, err := base.NewRunner(mod, bt.client)
		if err != nil {
			return err
		}
		go runner.Run()
	}

	logp.Info("Monitoring...")

	if bt.config.Manager.Enabled() {
		m, err := manager.NewManager(bt.config.Manager)
		if err != nil {
			logp.Error(err)
		} else {
			go m.Run()
		}
	}
	<-bt.done
	return nil
}

// Stop stops monbeat.
func (bt *Monbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
