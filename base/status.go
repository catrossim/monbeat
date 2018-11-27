package base

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

type Scode int32

const (
	StatusEnabled Scode = 1 << iota
	StatusDisabled
	StatusStandby
	StatusRunning
	StatusStopped
)

type ModuleStatus struct {
	status Scode
	lock   sync.Mutex
}

type StatusCenter struct {
	statusMap map[string]*ModuleStatus
	lock      sync.Mutex
}

var StatusRegistry = NewStatusCenter()

func NewStatusCenter() *StatusCenter {
	return &StatusCenter{
		statusMap: make(map[string]*ModuleStatus, 10),
	}
}

func (sc *StatusCenter) Add(name string) error {
	if _, ok := sc.statusMap[name]; !ok {
		sc.statusMap[name] = &ModuleStatus{
			status: StatusStandby,
		}
		return nil
	}
	return errors.New(fmt.Sprintf("%s has already registered.", name))
}

func (sc *StatusCenter) Update(name string, sCode Scode) error {
	if smap, ok := sc.statusMap[name]; ok {
		smap.lock.Lock()
		defer smap.lock.Unlock()
		smap.status = sCode
		return nil
	}
	return errors.New(fmt.Sprintf("%s has not registered yet.", name))
}

func (sc *StatusCenter) Enable(name string) error {
	return sc.Update(name, StatusEnabled)
}

func (sc *StatusCenter) Disable(name string) error {
	return sc.Update(name, StatusDisabled)
}

func (sc *StatusCenter) Remove(name string) {
	if smap, ok := sc.statusMap[name]; ok {
		smap.lock.Lock()
		defer smap.lock.Unlock()
		delete(sc.statusMap, name)
	}
}

func (sc *StatusCenter) Get(name string) (Scode, error) {
	if smap, ok := sc.statusMap[name]; ok {
		return smap.status, nil
	}
	return -1, errors.New(fmt.Sprintf("%s has not registered yet.", name))
}

func (sc *StatusCenter) IsEnabled(name string) bool {
	if smap, ok := sc.statusMap[name]; ok {
		if smap.status == StatusEnabled {
			return true
		}
	}
	return false
}
