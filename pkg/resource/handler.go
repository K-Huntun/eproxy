// Copyright (c) 2016-2017 ByteDance, Inc. All rights reserved.
package resource

import (
	"github.com/eproxy/pkg/manager"
	discovery "k8s.io/api/discovery/v1"
	"k8s.io/client-go/tools/cache"
)

type ServiceAdapterHandler struct {
	manager.ServiceHandler
}

func (s *ServiceAdapterHandler) OnAdd(obj interface{}) {

}

func (s *ServiceAdapterHandler) OnUpdate(oldObj, newObj interface{}) {
	//s.event <- KubernetesEvent{}
}

func (s *ServiceAdapterHandler) OnDelete(obj interface{}) {
}

var _ cache.ResourceEventHandler = &ServiceAdapterHandler{}

type EndpointSliceAdapterHandler struct {
	manager.EndpointSliceHandler
}

func (s *EndpointSliceAdapterHandler) OnAdd(obj interface{}) {
	s.OnAddEndpointSlice(obj.(*discovery.EndpointSlice))
}

func (s *EndpointSliceAdapterHandler) OnUpdate(oldObj, newObj interface{}) {
	s.OnUpdateEndpointSlice(oldObj.(*discovery.EndpointSlice), newObj.(*discovery.EndpointSlice))
}

func (s *EndpointSliceAdapterHandler) OnDelete(obj interface{}) {
	s.OnDeleteEndpointSlice(obj.(*discovery.EndpointSlice))
}

var _ cache.ResourceEventHandler = &EndpointSliceAdapterHandler{}
