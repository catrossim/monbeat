package beater

import (
	"fmt"
	"sync"

	"github.com/elastic/beats/libbeat/cfgfile"

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
	modules []staticModule
	config  config.Config
	client  beat.Client
}

type staticModule struct {
	connector *base.Connector
	module    *base.ModuleWrapper
}

// New creates an instance of monbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Monbeat{
		done:   make(chan struct{}),
		config: c,
	}
	for _, mCfg := range c.Modules {
		connector, err := base.NewConnector(b.Publisher, mCfg, nil)
		if err != nil {
			logp.Error(err)
			continue
		}

		module, err := base.NewModule(mCfg, base.Registry)
		if err != nil {
			logp.Error(err)
			continue
		}

		bt.modules = append(bt.modules, staticModule{
			connector: connector,
			module:    module,
		})
	}

	return bt, nil
}

// Run starts monbeat.
func (bt *Monbeat) Run(b *beat.Beat) error {
	logp.Info("Monbeat is running! Hit CTRL-C to stop it.")
	var wg sync.WaitGroup

	modules := bt.modules
	for _, mod := range modules {
		client, err := mod.connector.Connect()
		if err != nil {
			return err
		}
		r := base.NewRunner(mod.module, client)
		wg.Add(1)
		go r.Start()
		go func() {
			defer wg.Done()
			<-bt.done
			r.Stop()
		}()
	}

	logp.Info("Monitoring...")
	if bt.config.ConfigModules.Enabled() {
		moduleReloader := cfgfile.NewReloader(b.Publisher, bt.config.ConfigModules)
		factory := base.NewFactory()
		if err := moduleReloader.Check(factory); err != nil {
			return err
		}
		go moduleReloader.Run(factory)
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-bt.done
			moduleReloader.Stop()
		}()
	}

	if bt.config.Manager.Enabled() {
		m, err := manager.NewManager(bt.config.Manager)
		if err != nil {
			logp.Error(err)
		} else {
			wg.Add(1)
			go m.Start()
			go func() {
				defer wg.Done()
				<-bt.done
				m.Stop()
			}()
		}
	}
	wg.Wait()
	return nil
}

// Stop stops monbeat.
func (bt *Monbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
