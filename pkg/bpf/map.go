// Copyright (c) 2016-2017 ByteDance, Inc. All rights reserved.
package bpf

import (
	"github.com/cilium/ebpf"
	"github.com/sirupsen/logrus"
)

type Service struct {
	ipv6        bool
	service_map *ebpf.Map
}

func (s *Service) IsIpv6() bool {
	return s.ipv6
}

func (s *Service) LookUpElemSerivceMap(key ServiceKey) ServiceValue {
	value := Service4Value{}
	s.service_map.Lookup(key, &value)
	return &value
}

func (s *Service) DeleteElemSerivceMap(Key ServiceKey) error {
	err := s.service_map.Delete(Key)
	return err
}

func (s *Service) UpdateElemSerivceMap(Key ServiceKey, value ServiceValue) error {
	err := s.service_map.Update(Key, value, ebpf.UpdateAny)
	return err
}

var _ ServiceMap = &Service{}

// TODO no use
type Endpoint struct {
	ipv6         bool
	endpoint_map *ebpf.Map
}

func (e *Endpoint) LookUpElemEndpointMap(key EndpointKey) EndpointValue {
	return nil
}

func (e *Endpoint) DeleteElemEndpointMap(Key EndpointKey) error {
	return nil
}

func (e *Endpoint) UpdateElemEndpointMap(Key EndpointKey, value EndpointValue) error {
	return nil
}

var _ EndpointMap = &Endpoint{}

type emptyMap struct{}

func (e emptyMap) LookUpElemSerivceMap(key ServiceKey) ServiceValue {
	logrus.Error("EndpointMap looks up empty map")
	return nil
}

func (e emptyMap) DeleteElemSerivceMap(Key ServiceKey) error {
	logrus.Error("EndpointMap deletes empty map")
	return nil
}

func (e emptyMap) UpdateElemSerivceMap(Key ServiceKey, value ServiceValue) error {
	logrus.Error("EndpointMap updates empty map")
	return nil
}

func (e emptyMap) LookUpElemEndpointMap(key EndpointKey) EndpointValue {
	logrus.Error("EndpointMap looks up empty map")
	return nil
}

func (e emptyMap) DeleteElemEndpointMap(Key EndpointKey) error {
	logrus.Error("EndpointMap deletes empty map")
	return nil
}

func (e emptyMap) UpdateElemEndpointMap(Key EndpointKey, value EndpointValue) error {
	logrus.Error("EndpointMap updates empty map")
	return nil
}

var _ EndpointMap = &emptyMap{}
var _ ServiceMap = &emptyMap{}
