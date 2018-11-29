package registry

import (
	"fmt"
	"net"

	"github.com/pkg/errors"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/hashicorp/consul/api"
)

type ConsulRegister struct {
	config ConsulConfig
	sAddr  ServiceAddress
	logger *logp.Logger
}

type ServiceAddress struct {
	IP   string
	Port int
}

func New(config *common.Config, sAddr ServiceAddress) (*ConsulRegister, error) {
	cfg := DefaultConsulConfig
	if err := config.Unpack(&cfg); err != nil {
		return nil, err
	}
	if sAddr.IP == "" {
		sAddr.IP = localIP()
	}
	if sAddr.Port == 0 {
		return nil, errors.Errorf("service port should not be 0")
	}
	return &ConsulRegister{
		config: cfg,
		sAddr:  sAddr,
		logger: logp.NewLogger("consul"),
	}, nil
}

func (cr *ConsulRegister) Register() error {
	config := api.DefaultConfig()
	config.Address = cr.config.Address
	client, err := api.NewClient(config)
	if err != nil {
		return err
	}

	agent := client.Agent()
	reg := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%v-%v-%v", cr.config.Service, cr.sAddr.IP, cr.sAddr.Port),
		Name:    fmt.Sprintf("grpc.health.v1.%v", cr.config.Service),
		Tags:    cr.config.Tag,
		Port:    cr.sAddr.Port,
		Address: cr.sAddr.IP,
		Check: &api.AgentServiceCheck{
			Interval: cr.config.Interval.String(),
			GRPC:     fmt.Sprintf("%v:%v/%v", cr.sAddr.IP, cr.sAddr.Port, cr.config.Service),
			DeregisterCriticalServiceAfter: cr.config.DeregisterCriticalServiceAfter.String(),
		},
	}
	if err := agent.ServiceRegister(reg); err != nil {
		return err
	}
	cr.logger.Debugf("Register to consul registry: %s", cr.config.Address)
	return nil
}

func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
