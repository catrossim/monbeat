package base

import (
	"errors"
	"fmt"
	"strings"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/elastic/beats/libbeat/common"
)

type ModuleWrapper struct {
	module Module
	status Scode
}

func (mw *ModuleWrapper) run(done chan struct{}) chan beat.Event {
	out := make(chan beat.Event)
	switch m := mw.module.(type) {
	case Monitor:
		r := NewReporter(mw, out, done)
		go mw.start(m, r)
	default:
		error := fmt.Errorf("module %s is not supported", m.Name())
		logp.Error(error)
	}
	return out
}

func (mw *ModuleWrapper) start(mon Monitor, reporter Reporter) {
	go mon.Monitor()
	for {
		select {
		case <-reporter.Done():
			return
		case event := <-mon.Out():
			event.Put("module", mw.module.Name())
			reporter.Event(event)
		case err := <-mw.module.ErrorChannel():
			reporter.Error(err)
		}
	}
}

func NewModule(config *common.Config, r *Register) (*ModuleWrapper, error) {
	if !config.Enabled() {
		return nil, errors.New("module disabled error")
	}

	bm, err := newBaseModuleFromConfig(config)
	logp.Debug("module", "New module from config")
	if err != nil {
		return nil, err
	}

	module, err := createModule(r, bm)
	if err != nil {
		return nil, err
	}
	logp.Debug("module", "Create module %s", bm.Name())
	return &ModuleWrapper{
		module,
		StatusStandby,
	}, nil
}

func newBaseModuleFromConfig(rawConfig *common.Config) (BaseModule, error) {
	baseModule := BaseModule{
		config:    DefaultModuleConfig(),
		rawConfig: rawConfig,
		errorChan: make(chan error),
	}
	err := rawConfig.Unpack(&baseModule.config)
	if err != nil {
		return baseModule, err
	}

	if baseModule.config.Timeout == 0 {
		baseModule.config.Timeout = baseModule.config.Period
	}

	baseModule.name = strings.ToLower(baseModule.config.Module)

	return baseModule, nil
}

func createModule(r *Register, bm BaseModule) (Module, error) {
	f := r.moduleFactory(bm.Name())
	if f == nil {
		logp.Debug("module", "Module %s was created by default.", bm.Name())
		f = DefaultModuleFactory
	}
	return f(bm)
}
