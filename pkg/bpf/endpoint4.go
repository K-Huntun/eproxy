// Copyright (c) 2016-2017 ByteDance, Inc. All rights reserved.
package bpf

// Endpoint4Key 必须和bpf代码对齐
type Endpoint4Key struct {
	ServiceIP    uint32
	ServicePort  uint16
	Backend_slot uint8
	Proto        uint8
	Node         uint8
	Pad          pad2uint8
}

// Endpoint4Value 必须和bpf代码对齐
type Endpoint4Value struct {
	BackendID uint32
	Count     uint16
	Pad       pad2uint8
}

type Endpoint4 struct {
	key   Endpoint4Key
	value Endpoint4Value
}
