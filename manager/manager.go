package manager

import (
	"net"

	"github.com/catrossim/monbeat/manager/pb"
	"github.com/catrossim/monbeat/manager/server"

	"google.golang.org/grpc"

	"github.com/elastic/beats/libbeat/logp"

	"github.com/elastic/beats/libbeat/common"
	"github.com/pkg/errors"
)

type manager struct {
	config *ServerConfig
	logger *logp.Logger
}

func NewManager(cfg *common.Config) (*manager, error) {
	config := DefaultServerConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, errors.Wrapf(err, "failed to unpack config of manager")
	}
	return &manager{
		config: &config,
		logger: logp.NewLogger("manager"),
	}, nil
}

func (m *manager) Run() error {
	c, err := net.Listen(m.config.Network, m.config.Address)
	if err != nil {
		m.logger.Error(err)
		return err
	}
	defer c.Close()
	s := grpc.NewServer()
	rs, err := server.NewServer(m.config.WorkDir)
	if err != nil {
		m.logger.Error(err)
		return err
	}
	pb.RegisterRemoteServer(s, rs)
	if err = s.Serve(c); err != nil {
		return errors.Wrapf(err, "failed to serve tcp connection")
	}
	return nil
}

// func main() {
// 	m := &manager{
// 		config: &DefaultServerConfig,
// 	}

// 	err := m.Run()
// 	if err != nil {
// 		logp.Error(err)
// 		return
// 	}
// }
