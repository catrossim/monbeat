package manager

import (
	"net"
	"strconv"
	"strings"

	"github.com/catrossim/monbeat/manager/registry"

	"github.com/catrossim/manager/pb"
	"github.com/catrossim/manager/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

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

func (m *manager) Start() error {
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
	grpc_health_v1.RegisterHealthServer(s, &server.HealthImpl{})

	tokens := strings.Split(m.config.Address, ":")
	var port int
	if len(tokens) < 2 {
		port = 80
	} else {
		port, err = strconv.Atoi(tokens[1])
		if err != nil {
			return errors.Wrap(err, "invalid port format")
		}
	}
	registry, err := registry.New(m.config.RegistryConfig, registry.ServiceAddress{
		IP:   tokens[0],
		Port: port,
	})

	if err != nil {
		return err
	}
	if err := registry.Register(); err != nil {
		return err
	}
	if err = s.Serve(c); err != nil {
		return errors.Wrapf(err, "failed to serve tcp connection")
	}
	return nil
}

func (m *manager) Stop() {

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
