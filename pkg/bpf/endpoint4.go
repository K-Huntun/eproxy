// Copyright (c) 2016-2017 ByteDance, Inc. All rights reserved.
package bpf

// Endpoint4Key 必须和bpf代码对齐
type Endpoint4Key struct {
	EndpointID uint32 // 前16位 serviceid,后16位endpointid
	Pad        pad2uint8
}

// Endpoint4Value 必须和bpf代码对齐
type Endpoint4Value struct {
	EndpointIP   uint32
	EndpointPort uint16
	Proto        uint8
	Pad          pad2uint8
}

type Endpoint4 struct {
	key   Endpoint4Key
	value Endpoint4Value
}
