// Copyright (c) 2016-2017 ByteDance, Inc. All rights reserved.

// Licensed under the MIT license;
package controller

import (
	"strings"

	"github.com/eproxy/pkg/defaults"
	"github.com/eproxy/pkg/manager"
	"github.com/sirupsen/logrus"
	discovery "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	corev1 "k8s.io/client-go/informers/core/v1"
	discoveryv1 "k8s.io/client-go/informers/discovery/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	listersv1 "k8s.io/client-go/listers/discovery/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	LabelServiceName = "kubernetes.io/service-name"
)

type Controller struct {
	BaseController
	cluster          string
	serviceManager   *manager.ServiceManager
	KubernetesClient kubernetes.Interface
	serviceLister    corelisters.ServiceLister
	endpointsLister  listersv1.EndpointSliceLister
}

func NewController(service *manager.ServiceManager, k8sClient kubernetes.Interface, serviceinformer corev1.ServiceInformer, endpointinformer discoveryv1.EndpointSliceInformer) BController {
	ctl := &Controller{
		BaseController: BaseController{
			Workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), defaults.SvccontrollerName),
			Synced:    []cache.InformerSynced{endpointinformer.Informer().HasSynced, serviceinformer.Informer().HasSynced}, //serviceinformer.Informer().HasSynced
			Name:      defaults.SvccontrollerName,
		},
		KubernetesClient: k8sClient,
		serviceManager:   service,
		serviceLister:    serviceinformer.Lister(),
		endpointsLister:  endpointinformer.Lister(),
	}
	ctl.Handler = ctl.handler
	logrus.Info("Setting up event handlers")
	serviceinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.Enqueue,
		UpdateFunc: func(old, new interface{}) {
			ctl.Enqueue(new)
		},
		DeleteFunc: ctl.Enqueue,
	})
	endpointinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.enqueueServiceByEndpointSlice,
		UpdateFunc: func(old, new interface{}) {
			ctl.enqueueServiceByEndpointSlice(new)
		},
		DeleteFunc: ctl.enqueueServiceByEndpointSlice,
	})
	return ctl
}

func (c *Controller) handler(key string) error {
	keyArr := strings.Split(key, "/")
	if len(keyArr) != 3 {
		logrus.Errorf("invalid key: %s", key)
		return nil
	}
	if keyArr[0] == ServiceType {
		return c.ServiceHandler(keyArr[2], keyArr[1])
	}
	if keyArr[0] == EndpointSliceType {
		return c.EndpointHandler(keyArr[2], keyArr[1])
	}
	logrus.Errorf("unsupport key: %s", key)
	return nil
}

func (c *Controller) ServiceHandler(name string, namespace string) error {
	logrus.Info("Service Handler handle one event, namespace: ", namespace, ",name: ", name)
	svc, err := c.serviceLister.Services(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return c.serviceManager.DeleteService(namespace + "/" + name)
		}
		logrus.Error("can't get service ", name, err)
		return err
	}
	if svc.DeletionTimestamp != nil {
		return c.serviceManager.DeleteService(namespace + "/" + name)
	}

	selector := labels.Set{LabelServiceName: name}.AsSelectorPreValidated()
	endpointSlices, err := c.endpointsLister.EndpointSlices(namespace).List(selector)
	if err != nil {
		logrus.Error("can't list endpointslices for service ", name, err)
		return err
	}

	bsvc := manager.NewService(svc, endpointSlices, nil)
	return c.serviceManager.UpdateService(bsvc)
}

func (c *Controller) EndpointHandler(namespace string, name string) error {
	logrus.Info("EndpointSlice Handler handle one event, namespace: ", namespace, ",name: ", name)
	endpointSlice, err := c.endpointsLister.EndpointSlices(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		logrus.Error("can't get endpointSlice ", name, err)
		return err
	}
	svcname := endpointSlice.Labels[LabelServiceName]
	if svcname == "" {
		return nil
	}
	return c.ServiceHandler(svcname, namespace)
}

func (c *Controller) enqueueServiceByEndpointSlice(obj interface{}) {
	endpointSlice := endpointSliceFromObj(obj)
	if endpointSlice == nil {
		return
	}
	svcname := endpointSlice.Labels[LabelServiceName]
	if svcname == "" {
		return
	}
	c.Workqueue.Add(ServiceType + "/" + endpointSlice.Namespace + "/" + svcname)
}

func endpointSliceFromObj(obj interface{}) *discovery.EndpointSlice {
	switch value := obj.(type) {
	case *discovery.EndpointSlice:
		return value
	case cache.DeletedFinalStateUnknown:
		endpointSlice, ok := value.Obj.(*discovery.EndpointSlice)
		if !ok {
			return nil
		}
		return endpointSlice
	default:
		return nil
	}
}
