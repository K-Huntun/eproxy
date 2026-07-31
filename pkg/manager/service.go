package manager

import (
	"encoding/binary"
	"net"
	"sort"

	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Port struct {
	Name       string
	Protocol   v1.Protocol
	Port       uint16
	TargetPort uint16
	NodePort   uint16
	ServiceID  uint16
}

type Service struct {
	Name        string
	Namespace   string
	ClusterIP   net.IP
	Ports       []Port
	Endpoints   []uint32
	NodePortIPs []uint32
}

func NewService(service *v1.Service, endpointSlices []*discoveryv1.EndpointSlice, nodePortIPs []uint32) *Service {
	svc := &Service{}
	svc.Name = service.Name
	svc.Namespace = service.Namespace
	svc.ClusterIP = net.ParseIP(service.Spec.ClusterIP).To4()
	svc.Ports = make([]Port, 0, len(service.Spec.Ports))
	svc.Endpoints = collectReadyIPv4Endpoints(endpointSlices)
	svc.NodePortIPs = append([]uint32(nil), nodePortIPs...)

	for _, port := range service.Spec.Ports {
		svc.Ports = append(svc.Ports, Port{
			Name:       port.Name,
			Protocol:   port.Protocol,
			Port:       uint16(port.Port),
			TargetPort: resolveTargetPort(port, endpointSlices),
			NodePort:   uint16(port.NodePort),
		})
	}
	return svc
}

func resolveTargetPort(port v1.ServicePort, endpointSlices []*discoveryv1.EndpointSlice) uint16 {
	if port.TargetPort.Type == intstr.Int && port.TargetPort.IntValue() > 0 {
		return uint16(port.TargetPort.IntValue())
	}

	for _, endpointSlice := range endpointSlices {
		for _, endpointPort := range endpointSlice.Ports {
			if endpointPort.Port == nil {
				continue
			}
			if endpointPort.Protocol != nil && *endpointPort.Protocol != port.Protocol {
				continue
			}
			if endpointPort.Name != nil && *endpointPort.Name == port.Name {
				return uint16(*endpointPort.Port)
			}
		}
	}

	if port.TargetPort.Type == intstr.String && port.TargetPort.StrVal != "" {
		for _, endpointSlice := range endpointSlices {
			for _, endpointPort := range endpointSlice.Ports {
				if endpointPort.Port == nil || endpointPort.Name == nil {
					continue
				}
				if endpointPort.Protocol != nil && *endpointPort.Protocol != port.Protocol {
					continue
				}
				if *endpointPort.Name == port.TargetPort.StrVal {
					return uint16(*endpointPort.Port)
				}
			}
		}
	}

	if len(endpointSlices) == 1 && len(endpointSlices[0].Ports) == 1 && endpointSlices[0].Ports[0].Port != nil {
		endpointPort := endpointSlices[0].Ports[0]
		if endpointPort.Protocol == nil || *endpointPort.Protocol == port.Protocol {
			return uint16(*endpointPort.Port)
		}
	}

	if port.Port > 0 {
		return uint16(port.Port)
	}

	return 0
}

func collectReadyIPv4Endpoints(endpointSlices []*discoveryv1.EndpointSlice) []uint32 {
	endpointSet := map[uint32]struct{}{}

	for _, endpointSlice := range endpointSlices {
		for _, endpoint := range endpointSlice.Endpoints {
			if endpoint.Conditions.Ready != nil && !*endpoint.Conditions.Ready {
				continue
			}
			for _, ip := range endpoint.Addresses {
				parsedIP := net.ParseIP(ip).To4()
				if parsedIP == nil {
					continue
				}
				ipv4 := binary.LittleEndian.Uint32(parsedIP)
				if ipv4 == 0 {
					continue
				}
				endpointSet[ipv4] = struct{}{}
			}
		}
	}

	endpoints := make([]uint32, 0, len(endpointSet))
	for endpoint := range endpointSet {
		endpoints = append(endpoints, endpoint)
	}
	sort.Slice(endpoints, func(i, j int) bool {
		return endpoints[i] < endpoints[j]
	})
	return endpoints
}

func (s *Service) ServiceKey() string {
	return s.Namespace + "/" + s.Name
}
