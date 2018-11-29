package file_rb

import (
	"sync"

	"github.com/catrossim/monbeat/base"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
)

func init() {
	base.Registry.AddModule("file_rb", New)
}

type FileModule struct {
	base.BaseModule
	cfg    *FileConfig
	out    chan *common.MapStr
	logger *logp.Logger
}

func (fm *FileModule) Out() chan *common.MapStr {
	return fm.out
}
func (fm *FileModule) Done() {
	close(fm.out)
}

func New(bm base.BaseModule) (base.Module, error) {
	// read config
	cfg := DefaultFileConfig
	if err := bm.UnpackConfig(cfg); err != nil {
		return nil, err
	}

	return &FileModule{
		bm,
		cfg,
		make(chan *common.MapStr),
		logp.NewLogger("file_rb"),
	}, nil
}

func (fm *FileModule) Monitor(done chan struct{}) error {
	fm.logger.Debugf("%d targets are detected.", len(fm.cfg.Paths))
	var wg sync.WaitGroup
	for _, path := range fm.cfg.Paths {
		pathCfg := DefaultPathConfig
		err := path.Unpack(pathCfg)
		if err != nil {
			fm.logger.Error("Error in unpakcing config.")
			fm.ErrorChannel() <- err
			continue
		}
		watcher, err := NewWatcher(pathCfg.Path, pathCfg.Internal, pathCfg.FullTextEnabled, fm.out, fm.logger, fm.ErrorChannel())
		if err != nil {
			fm.ErrorChannel() <- err
			return err
		}
		wg.Add(1)
		go watcher.Watch()
		go func() {
			defer wg.Done()
			<-done
			watcher.Close()
		}()
		fm.logger.Debugf("Monitor %s started.", pathCfg.Path)

	}
	wg.Wait()
	fm.Done()
	return nil
}
