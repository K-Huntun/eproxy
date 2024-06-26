package manager

import (
	"github.com/cilium/ebpf/link"
	"github.com/eproxy/pkg/bpf"
	"github.com/eproxy/pkg/cache"
	"github.com/eproxy/pkg/utils"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"sync"
)

const (
	LabelServiceName = "kubernetes.io/service-name"
)

type serviceManager struct {
	services map[string]*cache.Service
	lock     sync.RWMutex
	bpfMap   *bpf.ServiceBPF
	link     link.Link
}

func (s *serviceManager) OnAddEndpointSlice(endpointSlice *discovery.EndpointSlice) {

}

func (s *serviceManager) OnUpdateEndpointSlice(old *discovery.EndpointSlice, new *discovery.EndpointSlice) {
	if new.Labels == nil || len(new.Labels) == 0 {
		return
	}
	// TODO check change
	var needDelete = true
	svcname := new.Labels[LabelServiceName]
	service, ok := s.services[svcname+"/"+new.Namespace]
	if !ok {
		needDelete = false
		service = &cache.Service{
			Name:      svcname,
			Namespace: new.Namespace,
		}
	}
	eps := make([]uint32, 0, len(new.Endpoints))
	for _, ep := range new.Endpoints {
		if ep.Conditions.Ready != nil && *ep.Conditions.Ready {
			for _, ip := range ep.Addresses {
				if ret := utils.IPString2Int32(ip); ret == 0 {
					eps = append(eps, ret)
				}
			}
		}
	}
	if needDelete {
		s.bpfMap.DeleteService(service)
	}
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
	svc := &cache.Service{
		Name:      service.Name,
		Namespace: service.Namespace,
		//TODO 适配其他
		IpAddress: service.Spec.ClusterIP,
	}
	for _, port := range service.Spec.Ports {
		p := cache.Ports{
			Protocol: port.Protocol,
			Port:     uint16(port.Port),
		}
		svc.Ports.Add(p)
	}
	s.services[svc.Name] = svc
}

func (s *serviceManager) OnUpdateService(service *v1.Service) {
	// service not update
}

func (s *serviceManager) OnDeleteService(service *v1.Service) {
	// service not delete
}

var _ = &serviceManager{}

func NewServiceManager(link link.Link) *serviceManager {
	return &serviceManager{link: link}
}
