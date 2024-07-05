package manager

import (
	"github.com/eproxy/pkg/utils/set"
	v1 "k8s.io/api/core/v1"
)

type Ports struct {
	Protocol   v1.Protocol
	Port       uint16
	TargetPort uint16
}

type Service struct {
	Name      string
	Namespace string
	ServiceId uint16
	IpAddress string
	Ports     set.Set[Ports]
	Endpoints []uint32
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) ServiceKey() string {
	return s.Namespace + "/" + s.Name
}
