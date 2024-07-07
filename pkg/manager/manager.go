package manager

import (
	"encoding/binary"
	"github.com/cilium/ebpf"
	"github.com/eproxy/pkg/bpf"
	"github.com/eproxy/pkg/utils"
	"github.com/sirupsen/logrus"
	discovery "k8s.io/api/discovery/v1"
	"net"
	"sync"
)

const (
	LabelServiceName = "kubernetes.io/service-name"
)

type ServiceManager struct {
	services       map[string]*Service
	cacheSerivceId map[uint16]bool
	lock           sync.RWMutex
	serviceMap     *ebpf.Map
	endpointsMap   *ebpf.Map
}

func (s *ServiceManager) DeleteService(serviceKey string) error {
	svc, ok := s.services[serviceKey]
	if !ok {
		logrus.Info("service not found,key: ", serviceKey)
		return nil
	}
	svc.Ports.Iter(func(port Ports) error {
		key := bpf.Service4Key{
			ServiceIP:   binary.LittleEndian.Uint32(svc.IpAddress.To4()),
			ServicePort: utils.LittleEndianPort(port.Port),
			Proto:       bpf.ParseProto(port.Protocol),
		}
		if err := s.serviceMap.Delete(key); err != nil {
			logrus.Error("error deleting service map(service):", err)
			return err
		}
		for index, _ := range svc.Endpoints {
			key := bpf.Endpoint4Key{
				EndpointID: uint32(svc.ServiceId)<<16 | uint32(index),
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

func (s *ServiceManager) UpdateService(svc *Service) error {
	old, ok := s.services[svc.ServiceKey()]
	if !ok {
		logrus.Info("service not found,key: ", svc.ServiceKey(), ",add svc to bpf")
		s.AppendService(svc)
		return nil
	}
	logrus.Info("update svc to ")
	err := s.DeleteService(old.ServiceKey())
	s.AppendService(svc)
	return err
}

func (s *ServiceManager) AppendService(svc *Service) {
	for i := 1; i < 65535; i++ {
		if _, ok := s.cacheSerivceId[uint16(i)]; !ok {
			s.cacheSerivceId[svc.ServiceId] = true
			svc.ServiceId = uint16(i)
			break
		}
	}
	logrus.Infof("serivce(%s) id is: %d", svc.Name, svc.ServiceId)
	svc.Ports.Iter(func(port Ports) error {
		key := bpf.Service4Key{
			ServiceIP:   binary.LittleEndian.Uint32(svc.IpAddress.To4()),
			ServicePort: utils.LittleEndianPort(port.Port),
			Proto:       bpf.ParseProto(port.Protocol),
		}
		value := bpf.Service4Value{
			ServiceID: svc.ServiceId,
			Count:     uint16(len(svc.Endpoints)),
		}
		if s.serviceMap == nil {
			logrus.Info("service map not initialized")
			return nil
		}
		if err := s.serviceMap.Update(key, value, ebpf.UpdateAny); err != nil {
			logrus.Error("error Append service map(service):", err)
			return err
		}
		for index, Eip := range svc.Endpoints {
			key := bpf.Endpoint4Key{
				EndpointID: uint32(svc.ServiceId)<<16 | uint32(index+1),
			}
			value := bpf.Endpoint4Value{
				EndpointIP:   Eip,
				EndpointPort: utils.LittleEndianPort(port.TargetPort),
			}
			if s.endpointsMap == nil {
				logrus.Info("endpoints map not initialized")
				return nil
			}
			if err := s.endpointsMap.Update(key, value, ebpf.UpdateAny); err != nil {
				logrus.Error("error Append service map(endpoints):", err)
				return err
			}
		}
		return nil
	})
	s.lock.Lock()
	s.services[svc.ServiceKey()] = svc
	s.lock.Unlock()
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
				if ret := binary.LittleEndian.Uint32(net.ParseIP(ip).To4()); ret == 0 {
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

var _ = &ServiceManager{}

func NewServiceManager(service, endpoint *ebpf.Map) *ServiceManager {
	return &ServiceManager{
		serviceMap:     service,
		endpointsMap:   endpoint,
		services:       make(map[string]*Service),
		cacheSerivceId: make(map[uint16]bool),
	}
}
