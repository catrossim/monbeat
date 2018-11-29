package base

import (
	"github.com/joeshaw/multierror"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/cfgfile"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/common/cfgwarn"
)

// Factory creates new Runner instances from configuration objects.
// It is used to register and reload modules.
type Factory struct {
}

// NewFactory creates new Reloader instance for the given config
func NewFactory() *Factory {
	return &Factory{}
}

// Create creates a new metricbeat module runner reporting events to the passed pipeline.
func (r *Factory) Create(p beat.Pipeline, c *common.Config, meta *common.MapStrPointer) (cfgfile.Runner, error) {
	var errs multierror.Errors

	err := cfgwarn.CheckRemoved5xSettings(c, "filters")
	if err != nil {
		errs = append(errs, err)
	}
	connector, err := NewConnector(p, c, meta)
	if err != nil {
		errs = append(errs, err)
	}
	w, err := NewModule(c, Registry)
	if err != nil {
		errs = append(errs, err)
	}

	if err := errs.Err(); err != nil {
		return nil, err
	}

	client, err := connector.Connect()
	if err != nil {
		return nil, err
	}

	mr := NewRunner(w, client)
	return mr, nil
}

// CheckConfig checks if a config is valid or not
func (r *Factory) CheckConfig(config *common.Config) error {
	_, err := NewModule(config, Registry)
	if err != nil {
		return err
	}

	return nil
}
