package file_rb

import (
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/elastic/beats/libbeat/logp"

	"github.com/catrossim/monbeat/utils"
	"github.com/elastic/beats/libbeat/common"
)

// Types of peration
type Op uint32

const (
	Create Op = 1 << iota
	Write
	Remove
	Rename
	Chmod
)

type Event struct {
	Path   string
	Op     Op
	Handle func()
}

type fileWatcher struct {
	path            string
	internal        time.Duration
	fulltextEnabled bool
	dataDir         string
	events          chan *Event
	out             chan *common.MapStr
	err             chan error
	cache           os.FileInfo
	lock            sync.Mutex
	logger          *logp.Logger
}

// NewWatcher creates watcher for tracking fielchanges
func NewWatcher(path string, internal time.Duration, fte bool, out chan *common.MapStr, logger *logp.Logger, err chan error) (*fileWatcher, error) {

	return &fileWatcher{
		path:            path,
		internal:        internal,
		fulltextEnabled: fte,
		events:          make(chan *Event),
		out:             out,
		err:             err,
		logger:          logger,
	}, nil
}

func (fw *fileWatcher) Watch(done chan struct{}) error {
	ticker := time.NewTicker(fw.internal)
	for {
		// start timer
		select {
		case <-done:
			fw.Close()
			return nil
		case <-ticker.C:
		}
		// logic for tracking details
		go fw.watchOnce()

		go func() {
			select {
			case event := <-fw.events:
				fw.logger.Debugf("watcher", "Event [%d] is detected.", event.Op)
				event.Handle()
			case <-done:
				return
			}
		}()
	}
}

func (fw *fileWatcher) Close() {
	close(fw.out)
	close(fw.events)
}

func (fw *fileWatcher) watchOnce() {
	stat, err := os.Stat(fw.path)
	if err != nil {
		if os.IsNotExist(err) {
			fw.logger.Warnf("File %s is not exist.", fw.path)
			if fw.cache != nil {
				// Remove event
				fw.cache = nil
				event := &Event{
					Path: fw.path,
					Op:   Remove,
					Handle: func() {
						result := &common.MapStr{
							"path":   fw.path,
							"action": Remove,
						}
						fw.out <- result
					},
				}
				fw.events <- event
			}
			return
		}
		fw.logger.Error(err)
		fw.err <- err
		return
	}
	if fw.cache == nil {
		// Create event
		fw.cache = stat
		event := &Event{
			Path: fw.path,
			Op:   Create,
			Handle: func() {
				result := &common.MapStr{
					"path":   fw.path,
					"action": Create,
				}
				if fw.fulltextEnabled {
					content, err := ioutil.ReadFile(fw.path)
					if err != nil {
						fw.logger.Error(err)
					} else {
						result.Put("content", string(content))
					}
				}
				fw.out <- result
			},
		}
		fw.events <- event
		return
	}
	preToken, err := utils.GenFileToken([]byte(fw.cache.Name() + fw.cache.ModTime().String()))
	if err != nil {
		fw.logger.Error(err)
		fw.err <- err
		return
	}
	currToken, err := utils.GenFileToken([]byte(stat.Name() + stat.ModTime().String()))
	if err != nil {
		fw.logger.Error(err)
		fw.err <- err
		return
	}
	if preToken == currToken {
		fw.logger.Debugf("No changes were detected for file [%s]. ", fw.path)
		return
	} else {
		//Write event
		fw.cache = stat
		event := &Event{
			Path: fw.path,
			Op:   Write,
			Handle: func() {
				result := &common.MapStr{
					"path":   fw.path,
					"action": Write,
				}
				if fw.fulltextEnabled {
					content, err := ioutil.ReadFile(fw.path)
					if err != nil {
						fw.logger.Error(err)
						fw.err <- err
					} else {
						result.Put("content", string(content))
					}
				}
				fw.out <- result
			},
		}
		fw.events <- event
		return
	}
}
