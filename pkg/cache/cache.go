package cache

import (
	"github.com/eproxy/pkg/set"
	v1 "k8s.io/api/core/v1"
)

type Ports struct {
	Protocol v1.Protocol
	Port     uint16
}

type Service struct {
	Name      string
	Namespace string
	ServiceId uint16
	IpAddress string
	Ports     set.Set[Ports]
	Endpoints []uint16
}
