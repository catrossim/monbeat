package base

import (
	"fmt"
	"strings"
	"sync"

	"github.com/elastic/beats/libbeat/logp"
)

const initSize = 20

var Registry = NewRegister()

type ModuleFactory func(base BaseModule) (Module, error)

var DefaultModuleFactory = func(base BaseModule) (Module, error) {
	return &base, nil
}

type Register struct {
	lock    sync.RWMutex
	modules map[string]ModuleFactory
}

func NewRegister() *Register {
	return &Register{
		modules: make(map[string]ModuleFactory, initSize),
	}
}

func (r *Register) AddModule(name string, factory ModuleFactory) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if name == "" {
		return fmt.Errorf("Module name should not be empty.")
	}
	name = strings.ToLower(name)
	if factory == nil {
		return fmt.Errorf("One module factory is needed for module [%s]", name)
	}
	if _, exists := r.modules[name]; exists {
		return fmt.Errorf("Module [%s] is already exist.", name)
	}
	r.modules[name] = factory
	logp.Info("Module %s is registered.", name)
	return nil
}

func (r *Register) moduleFactory(name string) ModuleFactory {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.modules[name]
}

func (r *Register) Modules() ([]string, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	modules := make([]string, len(r.modules))
	for name := range r.modules {
		modules = append(modules, name)
	}
	return modules, nil
}
