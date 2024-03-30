// Copyright (c) 2016-2017 ByteDance, Inc. All rights reserved.
package resource

import (
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"k8s.io/client-go/informers"
	v1 "k8s.io/client-go/informers/core/v1"
	discoveryv1 "k8s.io/client-go/informers/discovery/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"time"
)

const defaultResync = 0 * time.Second

type Resources struct {
	informers            informers.SharedInformerFactory
	ServiceInformer      v1.ServiceInformer
	EndpointSliceInfomer discoveryv1.EndpointSliceInformer
}

func NewResources(client kubernetes.Interface) *Resources {
	resources := &Resources{}
	informers := informers.NewSharedInformerFactory(client, defaultResync)
	informers.InformerFor(&discovery.EndpointSlice{}, defaultCustomEndpointSliceInformer)
	informers.InformerFor(&corev1.Service{}, defaultCustomServiceInformer)
	resources.ServiceInformer = informers.Core().V1().Services()
	resources.EndpointSliceInfomer = informers.Discovery().V1().EndpointSlices()
	return resources
}

func (resources *Resources) SetServiceHandler(handler cache.ResourceEventHandler) {
	resources.ServiceInformer.Informer().AddEventHandler(handler)
}

func (resources *Resources) SetEndpointHandler(handler cache.ResourceEventHandler) {
	resources.EndpointSliceInfomer.Informer().AddEventHandler(handler)
}

func (resources *Resources) StartListenEventFromKubernetes(stopCh <-chan struct{}) {
	resources.ServiceInformer.Informer().Run(stopCh)
	resources.EndpointSliceInfomer.Informer().Run(stopCh)
}
