package manager

import (
	"github.com/cilium/ebpf"
	"github.com/eproxy/pkg/bpf"
	"github.com/eproxy/pkg/utils"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"math/big"
	"net"
	"sync"
)

const (
	LabelServiceName = "kubernetes.io/service-name"
)

type ServiceManager struct {
	services     map[string]*Service
	lock         sync.RWMutex
	serviceMap   *ebpf.Map
	endpointsMap *ebpf.Map
}

func (s *ServiceManager) DeleteService(serviceKey string) error {
	svc, ok := s.services[serviceKey]
	if !ok {
		logrus.Info("service not found,key: ", serviceKey)
		return nil
	}
	svc.Ports.Iter(func(port Ports) error {
		key := bpf.Service4Key{
			ServiceIP:   uint32(big.NewInt(0).SetBytes(net.ParseIP(svc.IpAddress).To4()).Int64()),
			ServicePort: port.Port,
			Proto:       bpf.ParseProto(port.Protocol),
			Pad:         bpf.Pad2uint8{},
		}
		if err := s.serviceMap.Delete(key); err != nil {
			logrus.Error("error deleting service map(service):", err)
			return err
		}
		for index, _ := range svc.Endpoints {
			key := bpf.Endpoint4Key{
				EndpointID: uint32(svc.ServiceId)<<16 | uint32(index),
				Pad:        bpf.Pad2uint8{},
			}
			if err := s.endpointsMap.Delete(key); err != nil {
				logrus.Error("error deleting service map(endpoint):", err)
				return err
			}
		}
		return nil
	})
	s.lock.Lock()
	delete(s.services, serviceKey)
	s.lock.Unlock()
	return nil
}

func (s *ServiceManager) UpdateService(svc *Service) {

}

func (s *ServiceManager) AppendService(svc *Service) {
	svc.Ports.Iter(func(port Ports) error {
		key := bpf.Service4Key{
			ServiceIP:   uint32(big.NewInt(0).SetBytes(net.ParseIP(svc.IpAddress).To4()).Int64()),
			ServicePort: port.Port,
			Proto:       bpf.ParseProto(port.Protocol),
			Pad:         bpf.Pad2uint8{},
		}
		value := bpf.Service4Value{
			ServiceID: svc.ServiceId,
			Count:     uint16(len(svc.Endpoints)),
			Pad:       bpf.Pad2uint8{},
		}

		if err := s.serviceMap.Update(key, value, ebpf.UpdateAny); err != nil {
			logrus.Error("error Append service map(service):", err)
			return err
		}
		for index, Eip := range svc.Endpoints {
			key := bpf.Endpoint4Key{
				EndpointID: uint32(svc.ServiceId)<<16 | uint32(index),
				Pad:        bpf.Pad2uint8{},
			}
			value := bpf.Endpoint4Value{
				EndpointIP:   Eip,
				EndpointPort: port.TargetPort,
				Pad:          bpf.Pad2uint8{},
			}
			if err := s.endpointsMap.Update(key, value, ebpf.UpdateAny); err != nil {
				logrus.Error("error Append service map(endpoints):", err)
				return err
			}
		}
		return nil
	})
}

func (s *ServiceManager) OnUpdateEndpointSlice(old *discovery.EndpointSlice, new *discovery.EndpointSlice) {
	logrus.Info("UpdateEndpointSlice, Name: ", new.Name)
	if new.Labels == nil || len(new.Labels) == 0 {
		return
	}
	var needDelete = true
	svcname := new.Labels[LabelServiceName]
	service, ok := s.services[svcname+"/"+new.Namespace]
	if !ok {
		needDelete = false
		service = &Service{
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
		s.DeleteService(service.ServiceKey())
	}
	service.Endpoints = eps
	s.AppendService(service)
	s.services[svcname+"/"+new.Namespace] = service
}

func (s *ServiceManager) OnDeleteEndpointSlice(endpointSlice *discovery.EndpointSlice) {
	logrus.Info("DeleteEndpointSlice, Name: ", endpointSlice.Name)
	svcname := endpointSlice.Labels[LabelServiceName]
	service, ok := s.services[svcname+"/"+endpointSlice.Namespace]
	if !ok {
		return
	}
	s.DeleteService(service.ServiceKey())
	delete(s.services, svcname+"/"+endpointSlice.Namespace)
}

func (s *ServiceManager) OnAddService(service *v1.Service) {
	logrus.Info("OnAddService, Name: ", service.Name)
	svc := &Service{
		Name:      service.Name,
		Namespace: service.Namespace,
		IpAddress: service.Spec.ClusterIP,
	}
	for _, port := range service.Spec.Ports {
		p := Ports{
			Protocol: port.Protocol,
			Port:     uint16(port.Port),
		}
		svc.Ports.Add(p)
	}
	s.services[svc.Name] = svc
}

var _ = &ServiceManager{}

func NewServiceManager(service, endpoint *ebpf.Map) *ServiceManager {
	return &ServiceManager{
		serviceMap:   service,
		endpointsMap: endpoint,
		services:     make(map[string]*Service),
	}
}
