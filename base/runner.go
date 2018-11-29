package base

import (
	"sync"

	"github.com/elastic/beats/libbeat/beat"
)

type runner struct {
	mod       *ModuleWrapper
	startOnce sync.Once
	stopOnce  sync.Once
	client    beat.Client
	done      chan struct{}
	wg        sync.WaitGroup
}

func NewRunner(mod *ModuleWrapper, client beat.Client) *runner {
	return &runner{
		mod:    mod,
		client: client,
		done:   make(chan struct{}),
	}
}

func (r *runner) Start() {
	r.startOnce.Do(func() {
		output := r.mod.run(r.done)
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			PublishChannels(r.client, output)
		}()
	})
}

func (r *runner) Stop() {
	r.stopOnce.Do(func() {
		close(r.done)
		r.client.Close()
		r.wg.Wait()
	})
}

func (r *runner) String() string {
	return r.mod.module.Name()
}

func PublishChannels(client beat.Client, cs ...<-chan beat.Event) {
	var wg sync.WaitGroup

	// output publishes values from c until c is closed, then calls wg.Done.
	output := func(c <-chan beat.Event) {
		defer wg.Done()
		for event := range c {
			client.Publish(event)
		}
	}

	// Start an output goroutine for each input channel in cs.
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}
	wg.Wait()
}
