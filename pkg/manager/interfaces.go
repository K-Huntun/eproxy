package manager

import (
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
)

type ServiceHandler interface {
	OnAddService(service *v1.Service)
	OnUpdateService(service *v1.Service)
	OnDeleteService(service *v1.Service)
}

type EndpointSliceHandler interface {
	OnAddEndpointSlice(endpointSlice *discovery.EndpointSlice)
	OnUpdateEndpointSlice(old *discovery.EndpointSlice, new *discovery.EndpointSlice)
	OnDeleteEndpointSlice(endpointSlice *discovery.EndpointSlice)
}
