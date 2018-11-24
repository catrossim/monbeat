package file_rb

import (
	"time"

	"github.com/elastic/beats/libbeat/common"
)

type PathConfig struct {
	Path            string        `config:"path"`
	Internal        time.Duration `config:"internal"`
	FullTextEnabled bool          `config:"fulltext.enable"`
}

var DefaultPathConfig = &PathConfig{
	Internal:        30 * time.Minute,
	FullTextEnabled: false,
}

type FileConfig struct {
	Paths []*common.Config `config:"paths"`
}

var DefaultFileConfig = &FileConfig{}
