// Copyright (c) 2016-2017 ByteDance, Inc. All rights reserved.
package bpf

import (
	"github.com/cilium/ebpf"
	"github.com/eproxy/pkg/cache"
	"github.com/sirupsen/logrus"
	"math/big"
	"net"
)

type ServiceBPF struct {
	ipv6   bool
	lb4map ebpf.Map
	lb6map ebpf.Map
}

func (s *ServiceBPF) IsIpv6() bool {
	return s.ipv6
}

func (s *ServiceBPF) LookUpElemSerivceMap(key ServiceKey) ServiceValue {
	value := Service4Value{}
	s.lb4map.Lookup(key, &value)
	return &value
}

func (s *ServiceBPF) DeleteElemSerivceMap(Key ServiceKey) error {
	err := s.lb4map.Delete(Key)
	return err
}

func (s *ServiceBPF) UpdateElemSerivceMap(Key ServiceKey, value ServiceValue) error {
	err := s.lb4map.Update(Key, value, ebpf.UpdateAny)
	return err
}

func (s *ServiceBPF) DeleteService(svc *cache.Service) {
	svc.Ports.Iter(func(port cache.Ports) error {
		key := Service4Key{
			ServiceIP:   uint32(big.NewInt(0).SetBytes(net.ParseIP(svc.IpAddress).To4()).Int64()),
			ServicePort: port.Port,
			Proto:       parseProto(port.Protocol),
			Pad:         pad2uint8{},
		}
		if err := s.DeleteElemSerivceMap(key); err != nil {
			logrus.Error("error deleting service map(service):", err)
			return err
		}
		for index, _ := range svc.Endpoints {
			key := Endpoint4Key{
				EndpointID: uint32(svc.ServiceId)<<16 | uint32(index),
				Pad:        pad2uint8{},
			}
			if err := s.DeleteElemSerivceMap(key); err != nil {
				logrus.Error("error deleting service map(endpoint):", err)
				return err
			}
		}
		return nil
	})
}

func (s *ServiceBPF) AppendService(svc *cache.Service) {
	svc.Ports.Iter(func(port cache.Ports) error {
		key := Service4Key{
			ServiceIP:   uint32(big.NewInt(0).SetBytes(net.ParseIP(svc.IpAddress).To4()).Int64()),
			ServicePort: port.Port,
			Proto:       parseProto(port.Protocol),
			Pad:         pad2uint8{},
		}
		value := Service4Value{
			ServiceID: svc.ServiceId,
			Count:     uint16(len(svc.Endpoints)),
			Pad:       pad2uint8{},
		}

		if err := s.UpdateElemSerivceMap(key, value); err != nil {
			logrus.Error("error Append service map(service):", err)
			return err
		}
		for index, Eip := range svc.Endpoints {
			key := Endpoint4Key{
				EndpointID: uint32(svc.ServiceId)<<16 | uint32(index),
				Pad:        pad2uint8{},
			}
			value := Endpoint4Value{
				EndpointIP:   Eip,
				EndpointPort: port.TargetPort,
				Pad:          pad2uint8{},
			}
			if err := s.UpdateElemSerivceMap(key, value); err != nil {
				logrus.Error("error Append service map(endpoints):", err)
				return err
			}
		}
		return nil
	})
}

var _ ServiceMap = &ServiceBPF{}

// TODO no use
type Endpoint struct {
	ipv6            bool
	lb_endpoint_map ebpf.Map
	lb6map          ebpf.Map
	// cache
	service map[EndpointKey]EndpointValue
}

func (e *Endpoint) LookUpElemSerivceMap(key EndpointKey) EndpointValue {
	return nil
}

func (e *Endpoint) DeleteElemSerivceMap(Key EndpointKey) error {
	return nil
}

func (e *Endpoint) UpdateElemSerivceMap(Key EndpointKey, value EndpointValue) error {
	return nil
}

var _ EndpointMap = &Endpoint{}
