package base

import (
	"errors"
	"time"

	"github.com/elastic/beats/libbeat/logp"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
)

type Reporter interface {
	Event(e *common.MapStr) bool
	Error(err error) bool
	Done() <-chan struct{}
}

type baseReporter struct {
	module *ModuleWrapper
	out    chan beat.Event
	done   chan struct{}
}

func NewReporter(module *ModuleWrapper, out chan beat.Event, done chan struct{}) *baseReporter {
	return &baseReporter{
		module,
		out,
		done,
	}
}

func (br *baseReporter) Event(e *common.MapStr) bool {
	if e == nil {
		err := errors.New("empty event error")
		logp.Error(err)
		br.Error(err)
		return false
	}

	event := beat.Event{
		Timestamp: time.Now().UTC(),
		Fields:    *e,
	}

	return writeEvent(event, br.out, br.done)
}

func (br *baseReporter) Error(err error) bool {
	e := &common.MapStr{
		"error": err.Error(),
	}
	return br.Event(e)
}

func (br *baseReporter) Done() <-chan struct{} {
	return br.done
}

func writeEvent(event beat.Event, out chan beat.Event, done chan struct{}) bool {
	select {
	case <-done:
		return false
	case out <- event:
		return true
	}
}
