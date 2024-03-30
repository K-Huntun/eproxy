// Copyright (c) 2016-2017 ByteDance, Inc. All rights reserved.
package bpf

import (
	v1 "k8s.io/api/core/v1"
)

// Service4Key 必须和bpf代码对齐
type Service4Key struct {
	ServiceIP    uint32
	ServicePort  uint16
	Backend_slot uint8
	Proto        uint8
	Pad          pad2uint8
}

// Service4Value 必须和bpf代码对齐
type Service4Value struct {
	BackendID uint32
	Count     uint16
	Pad       pad2uint8
}

func parseProto(proto v1.Protocol) uint8 {
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
