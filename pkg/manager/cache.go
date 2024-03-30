package manager

import (
	"github.com/eproxy/pkg/bpf"
	"github.com/eproxy/pkg/set"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"sync"
)

const (
	LabelServiceName = "kubernetes.io/service-name"
)

type Ports struct {
	Protocol v1.Protocol
	Port     uint16
}

type Service struct {
	name      string
	namespace string
	id        string
	IpAddress string
	Ports     set.Set[Ports]
	Endpoints []string
}

type serviceManager struct {
	services map[string]*Service
	lock     sync.RWMutex
	bpfMap   *bpf.ServiceBPF
}

func (s *serviceManager) OnAddEndpointSlice(endpointSlice *discovery.EndpointSlice) {

}

func (s *serviceManager) OnUpdateEndpointSlice(old *discovery.EndpointSlice, new *discovery.EndpointSlice) {
	if new.Labels == nil || len(new.Labels) == 0 {
		return
	}
	// TODO check change

	svcname := new.Labels[LabelServiceName]
	service, ok := s.services[svcname+"/"+new.Namespace]
	if !ok {
		service = &Service{
			name:      svcname,
			namespace: new.Namespace,
		}
	}
	eps := make([]string, 0, len(new.Endpoints))
	for _, ep := range new.Endpoints {
		if ep.Conditions.Ready != nil && *ep.Conditions.Ready {
			for _, ip := range ep.Addresses {
				eps = append(eps, ip)
			}
		}
	}
	s.bpfMap.DeleteService(service)
	service.Endpoints = eps
	s.bpfMap.AppendService(service)
	s.services[svcname+"/"+new.Namespace] = service
}

func (s *serviceManager) OnDeleteEndpointSlice(endpointSlice *discovery.EndpointSlice) {
	svcname := endpointSlice.Labels[LabelServiceName]
	service, ok := s.services[svcname+"/"+endpointSlice.Namespace]
	if !ok {
		return
	}
	s.bpfMap.DeleteService(service)
	delete(s.services, svcname+"/"+endpointSlice.Namespace)
}

func (s *serviceManager) OnAddService(service *v1.Service) {
	svc := &Service{
		name:      service.Name,
		namespace: service.Namespace,
		//TODO 适配其他
		IpAddress: service.Spec.ClusterIP,
	}
	for _, port := range service.Spec.Ports {
		p := Ports{
			Protocol: port.Protocol,
			Port:     uint16(port.Port),
		}
		svc.Ports.Add(p)
	}
	s.services[svc.name] = svc
}

func (s *serviceManager) OnUpdateService(service *v1.Service) {
	// service not update
}

func (s *serviceManager) OnDeleteService(service *v1.Service) {
	// service not delete
}

var _ = &serviceManager{}

func NewServiceManager() *serviceManager {
	return &serviceManager{}
}
