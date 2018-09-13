package kit

import (
	"errors"
	"strconv"
	"strings"

	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	stdconsul "github.com/hashicorp/consul/api"
)

func SimpleConsulRegistrar(id, name, addr string) (sd.Registrar, error) {
	client, err := stdconsul.NewClient(stdconsul.DefaultConfig())
	if err != nil {
		return nil, err
	}
	ss := strings.SplitN(addr, ":", 2)
	if len(ss) != 2 {
		return nil, errors.New("invalid service address")
	}
	addr = ss[0]
	port, err := strconv.Atoi(ss[1])
	if err != nil {
		return nil, errors.New("invalid port number")
	}
	registration := &stdconsul.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Address: addr,
		Port:    port,
	}
	registrar := consul.NewRegistrar(consul.NewClient(client), registration, DefaultLogger)
	return registrar, nil
}

func SimpleConsulInstancer(service string) (sd.Instancer, error) {
	client, err := stdconsul.NewClient(stdconsul.DefaultConfig())
	if err != nil {
		return nil, err
	}
	instancer := consul.NewInstancer(consul.NewClient(client), DefaultLogger, service, nil, true)
	return instancer, nil
}
