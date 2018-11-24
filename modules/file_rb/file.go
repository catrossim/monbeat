package file_rb

import (
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
	done   chan struct{}
	out    chan *common.MapStr
	logger *logp.Logger
}

func (fm *FileModule) DoneChannel() chan struct{} {
	return fm.done
}
func (fm *FileModule) Out() chan *common.MapStr {
	return fm.out
}
func (fm *FileModule) Done() {
	close(fm.out)
	close(fm.done)
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
		make(chan struct{}),
		make(chan *common.MapStr),
		logp.NewLogger("file_rb"),
	}, nil
}

func (fm *FileModule) Monitor() error {
	fm.logger.Debugf("%d targets are detected.", len(fm.cfg.Paths))

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
		go watcher.Watch(fm.done)

		fm.logger.Debugf("Monitor %s started.", pathCfg.Path)

	}
	return nil
}
