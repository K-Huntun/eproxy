// Copyright (c) 2016-2017 ByteDance, Inc. All rights reserved.
package informers

import (
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"k8s.io/client-go/informers"
	v1 "k8s.io/client-go/informers/core/v1"
	discoveryv1 "k8s.io/client-go/informers/discovery/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

const defaultResync = 0 * time.Second

type Resources struct {
	kubernetesClient kubernetes.Interface
	informers        informers.SharedInformerFactory
}

func (resources *Resources) KubernetetsClient() kubernetes.Interface {
	return resources.kubernetesClient
}

func (resources *Resources) ServiceInformer() v1.ServiceInformer {
	return resources.informers.Core().V1().Services()
}

func (resources *Resources) EndpointSliceInfomer() discoveryv1.EndpointSliceInformer {
	return resources.informers.Discovery().V1().EndpointSlices()
}

func (resources *Resources) StartListenEventFromKubernetes(stopCh <-chan struct{}) {
	resources.informers.Start(stopCh)
}

func NewResources(kubeconfig string) *Resources {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	resources := &Resources{}
	resources.kubernetesClient = kubernetes.NewForConfigOrDie(config)
	resources.informers = informers.NewSharedInformerFactory(resources.kubernetesClient, defaultResync)
	resources.informers.InformerFor(&discovery.EndpointSlice{}, defaultCustomEndpointSliceInformer)
	resources.informers.InformerFor(&corev1.Service{}, defaultCustomServiceInformer)
	return resources
}
