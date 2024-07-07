package manager

import (
	"encoding/binary"
	"github.com/eproxy/pkg/utils/set"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
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
	IpAddress net.IP
	Ports     set.Set[Ports]
	Endpoints []uint32
}

func NewService(service *v1.Service, endpointSlice *discoveryv1.EndpointSlice) *Service {
	svc := &Service{}
	svc.Name = service.Name
	svc.Namespace = service.Namespace
	svc.ServiceId = 0
	svc.IpAddress = net.ParseIP(service.Spec.ClusterIP).To4()
	svc.Ports = set.New[Ports]()
	svc.Endpoints = make([]uint32, 0)
	for _, port := range service.Spec.Ports {
		p := Ports{
			Protocol:   port.Protocol,
			Port:       uint16(port.Port),
			TargetPort: uint16(port.TargetPort.IntValue()),
		}
		svc.Ports.Add(p)
	}
	for _, endpoint := range endpointSlice.Endpoints {
		if *endpoint.Conditions.Ready == true {
			for _, ip := range endpoint.Addresses {
				if ipv4 := binary.LittleEndian.Uint32(net.ParseIP(ip).To4()); ipv4 != 0 {
					svc.Endpoints = append(svc.Endpoints, ipv4)
				}
			}
		}
	}
	return svc
}

func (s *Service) ServiceKey() string {
	return s.Namespace + "/" + s.Name
}
