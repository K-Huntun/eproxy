// Copyright (c) 2016-2017 ByteDance, Inc. All rights reserved.
package bpf

type pad2uint8 [2]uint8

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
	LookUpElemSerivceMap(key EndpointKey) EndpointValue
	DeleteElemSerivceMap(Key EndpointKey) error
	UpdateElemSerivceMap(Key EndpointKey, value EndpointValue) error
}
