// Copyright (c) 2016-2017 ByteDance, Inc. All rights reserved.
package bpf

import v1 "k8s.io/api/core/v1"

// Service4Key 必须和bpf代码对齐
type Service4Key struct {
	ServiceIP   uint32
	ServicePort uint16
	Proto       uint8
	Pad         Pad2uint8
}

// Service4Value 必须和bpf代码对齐
type Service4Value struct {
	ServiceID uint16
	Count     uint16
	Pad       Pad2uint8
}

func ParseProto(proto v1.Protocol) uint8 {
	switch proto {
	case v1.ProtocolTCP:
		return 1
	case v1.ProtocolUDP:
		return 2
	case v1.ProtocolSCTP:
		return 3
	}
	return 0
}

// Endpoint4Key 必须和bpf代码对齐
type Endpoint4Key struct {
	EndpointID uint32 // 前16位 serviceid,后16位endpointid
	Pad        Pad2uint8
}

// Endpoint4Value 必须和bpf代码对齐
type Endpoint4Value struct {
	EndpointIP   uint32
	EndpointPort uint16
	Proto        uint8
	Pad          Pad2uint8
}

type Endpoint4 struct {
	key   Endpoint4Key
	value Endpoint4Value
}

type Pad2uint8 [2]uint8

type ServiceMap interface {
	LookUpElemSerivceMap(key ServiceKey) ServiceValue
	DeleteElemSerivceMap(Key ServiceKey) error
	UpdateElemSerivceMap(Key ServiceKey, value ServiceValue) error
}

type ServiceKey interface {
}

type ServiceValue interface {
}

type EndpointKey interface {
}

type EndpointValue interface {
}

type EndpointMap interface {
	LookUpElemEndpointMap(key EndpointKey) EndpointValue
	DeleteElemEndpointMap(Key EndpointKey) error
	UpdateElemEndpointMap(Key EndpointKey, value EndpointValue) error
}
