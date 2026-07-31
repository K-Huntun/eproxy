package manager

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/eproxy/pkg/bpf"
	"github.com/eproxy/pkg/utils"
	"github.com/sirupsen/logrus"
)

const (
	LabelServiceName = "kubernetes.io/service-name"
)

type ServiceManager struct {
	services        map[string]*Service
	cacheSerivceId  map[uint16]bool
	lock            sync.RWMutex
	serviceMap      *ebpf.Map
	endpointsMap    *ebpf.Map
	nodeAddressFunc func() ([]uint32, error)
}

func (s *ServiceManager) DeleteService(serviceKey string) error {
	svc, ok := s.services[serviceKey]
	if !ok {
		logrus.Info("service not found,key: ", serviceKey)
		return nil
	}

	for _, port := range svc.Ports {
		for _, key := range s.frontendKeys(svc, port) {
			if s.serviceMap == nil {
				continue
			}
			if err := s.serviceMap.Delete(key); err != nil && !errors.Is(err, ebpf.ErrKeyNotExist) {
				logrus.Error("error deleting service map(service):", err)
				return err
			}
		}
		for index := range svc.Endpoints {
			key := bpf.Endpoint4Key{
				EndpointID: uint32(port.ServiceID)<<16 | uint32(index+1),
			}
			if s.endpointsMap == nil {
				continue
			}
			if err := s.endpointsMap.Delete(key); err != nil && !errors.Is(err, ebpf.ErrKeyNotExist) {
				logrus.Error("error deleting service map(endpoint):", err)
				return err
			}
		}
		s.releaseServiceID(port.ServiceID)
	}
	s.lock.Lock()
	delete(s.services, serviceKey)
	s.lock.Unlock()
	return nil
}

func (s *ServiceManager) UpdateService(svc *Service) error {
	old, ok := s.services[svc.ServiceKey()]
	if !ok {
		logrus.Info("service not found,key: ", svc.ServiceKey(), ",add svc to bpf")
		return s.AppendService(svc)
	}
	oldIDs := make(map[string]uint16, len(old.Ports))
	for _, port := range old.Ports {
		oldIDs[s.portIdentity(port)] = port.ServiceID
	}
	for i := range svc.Ports {
		svc.Ports[i].ServiceID = oldIDs[s.portIdentity(svc.Ports[i])]
	}
	if err := s.DeleteService(old.ServiceKey()); err != nil {
		return err
	}
	return s.AppendService(svc)
}

func (s *ServiceManager) AppendService(svc *Service) error {
	if len(svc.NodePortIPs) == 0 && s.nodeAddressFunc != nil {
		nodePortIPs, err := s.nodeAddressFunc()
		if err != nil {
			return err
		}
		svc.NodePortIPs = nodePortIPs
	}

	for i := range svc.Ports {
		if svc.Ports[i].ServiceID == 0 {
			serviceID, err := s.allocateServiceID()
			if err != nil {
				return err
			}
			svc.Ports[i].ServiceID = serviceID
		} else {
			s.cacheSerivceId[svc.Ports[i].ServiceID] = true
		}
		logrus.Infof("service(%s/%s) port(%d/%s) id is: %d", svc.Namespace, svc.Name, svc.Ports[i].Port, svc.Ports[i].Protocol, svc.Ports[i].ServiceID)

		value := bpf.Service4Value{
			ServiceID: svc.Ports[i].ServiceID,
			Count:     uint16(len(svc.Endpoints)),
		}

		for _, key := range s.frontendKeys(svc, svc.Ports[i]) {
			if s.serviceMap == nil {
				logrus.Info("service map not initialized")
				continue
			}
			if err := s.serviceMap.Update(key, value, ebpf.UpdateAny); err != nil {
				logrus.Error("error Append service map(service):", err)
				return err
			}
		}

		for index, endpointIP := range svc.Endpoints {
			key := bpf.Endpoint4Key{
				EndpointID: uint32(svc.Ports[i].ServiceID)<<16 | uint32(index+1),
			}
			value := bpf.Endpoint4Value{
				EndpointIP:   endpointIP,
				EndpointPort: utils.LittleEndianPort(svc.Ports[i].TargetPort),
				Proto:        bpf.ParseProto(svc.Ports[i].Protocol),
			}
			if s.endpointsMap == nil {
				logrus.Info("endpoints map not initialized")
				continue
			}
			if err := s.endpointsMap.Update(key, value, ebpf.UpdateAny); err != nil {
				logrus.Error("error Append service map(endpoints):", err)
				return err
			}
		}
	}

	s.lock.Lock()
	s.services[svc.ServiceKey()] = svc
	s.lock.Unlock()
	return nil
}

func (s *ServiceManager) frontendKeys(svc *Service, port Port) []bpf.Service4Key {
	keys := make([]bpf.Service4Key, 0, 1+len(svc.NodePortIPs))
	proto := bpf.ParseProto(port.Protocol)

	if svc.ClusterIP != nil {
		if clusterIP := binary.LittleEndian.Uint32(svc.ClusterIP.To4()); clusterIP != 0 && port.Port != 0 {
			keys = append(keys, bpf.Service4Key{
				ServiceIP:   clusterIP,
				ServicePort: utils.LittleEndianPort(port.Port),
				Proto:       proto,
			})
		}
	}

	if port.NodePort == 0 {
		return keys
	}

	for _, nodeIP := range svc.NodePortIPs {
		if nodeIP == 0 {
			continue
		}
		keys = append(keys, bpf.Service4Key{
			ServiceIP:   nodeIP,
			ServicePort: utils.LittleEndianPort(port.NodePort),
			Proto:       proto,
		})
	}

	return keys
}

func (s *ServiceManager) portIdentity(port Port) string {
	return fmt.Sprintf("%s/%d/%d/%s", port.Protocol, port.Port, port.NodePort, port.Name)
}

func (s *ServiceManager) allocateServiceID() (uint16, error) {
	for i := 1; i < 65535; i++ {
		serviceID := uint16(i)
		if _, ok := s.cacheSerivceId[serviceID]; ok {
			continue
		}
		s.cacheSerivceId[serviceID] = true
		return serviceID, nil
	}

	return 0, errors.New("no available service id")
}

func (s *ServiceManager) releaseServiceID(serviceID uint16) {
	if serviceID == 0 {
		return
	}
	delete(s.cacheSerivceId, serviceID)
}

var _ = &ServiceManager{}

func NewServiceManager(service, endpoint *ebpf.Map) *ServiceManager {
	return &ServiceManager{
		serviceMap:      service,
		endpointsMap:    endpoint,
		services:        make(map[string]*Service),
		cacheSerivceId:  make(map[uint16]bool),
		nodeAddressFunc: LocalNodeIPv4Addrs,
	}
}
