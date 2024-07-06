package manager

import (
	"github.com/eproxy/pkg/utils/set"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"math/big"
	"net"
)

type Ports struct {
	Protocol   v1.Protocol
	Port       uint16
	TargetPort uint16
}

type IP string

type Service struct {
	Name      string
	Namespace string
	ServiceId uint16
	IpAddress IP
	Ports     set.Set[Ports]
	Endpoints []uint32
}

func NewService(service *v1.Service, endpointSlice *discoveryv1.EndpointSlice) *Service {
	svc := &Service{}
	svc.Name = service.Name
	svc.Namespace = service.Namespace
	svc.ServiceId = 0
	svc.IpAddress = IP(service.Spec.ClusterIP)
	for _, port := range service.Spec.Ports {
		p := Ports{
			Protocol: port.Protocol,
			Port:     uint16(port.Port),
		}
		svc.Ports.Add(p)
	}
	return svc
}

func (s *Service) ServiceKey() string {
	return s.Namespace + "/" + s.Name
}

func (i IP) Address() uint32 {
	return uint32(big.NewInt(0).SetBytes(net.ParseIP(string(i)).To4()).Int64())
}
