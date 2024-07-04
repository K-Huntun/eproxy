// Copyright (c) 2016-2017 ByteDance, Inc. All rights reserved.

// Licensed under the MIT license;
package controller

import (
	"github.com/eproxy/pkg/manager"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/client-go/informers/core/v1"
	discoveryv1 "k8s.io/client-go/informers/discovery/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
	listersv1 "k8s.io/client-go/listers/discovery/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"strings"
)

const svccontrollerName = "svc-controller"

type Controller struct {
	BaseController
	cluster          string
	serviceManager   *manager.ServiceManager
	KubernetesClient kubernetes.Interface
	serviceLister    v1.ServiceLister
	endpointsLister  listersv1.EndpointSliceLister
}

func NewController(service *manager.ServiceManager, k8sClient kubernetes.Interface, serviceinformer corev1.ServiceInformer, endpointinformer discoveryv1.EndpointSliceInformer) BController {
	ctl := &Controller{
		BaseController: BaseController{
			Workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), svccontrollerName),
			Synced:    []cache.InformerSynced{serviceinformer.Informer().HasSynced, endpointinformer.Informer().HasSynced},
			Name:      svccontrollerName,
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
		AddFunc: ctl.Enqueue,
		UpdateFunc: func(old, new interface{}) {
			ctl.Enqueue(new)
		},
		DeleteFunc: ctl.Enqueue,
	})
	return ctl
}

func (c *Controller) handler(key string) error {
	keyArr := strings.Split(key, "/")
	if len(keyArr) != 3 {
		logrus.Errorf("invalid key: %s", key)
		return nil
	}
	if keyArr[0] == "services" {
		return c.ServiceHandler(keyArr[1], keyArr[2])
	}
	if keyArr[0] == "endpoints" {
		return c.ServiceHandler(keyArr[1], keyArr[2])
	}
	logrus.Errorf("unsupport key: %s", key)
	return nil
}

func (c *Controller) ServiceHandler(name string, namespace string) error {
	service, err := c.serviceLister.Services(namespace).Get(name)
	// TODO 创建新的svc
	bsvc := manager.NewService()
	if err != nil || !service.ObjectMeta.DeletionTimestamp.IsZero() {
		logrus.Info("svc is Deleted name:", name, ",namespace: ", namespace, ",err:", err)
		c.serviceManager.DeleteService(bsvc)
		return nil
	}
	logrus.Info("one svc had change,", c.cluster, "/", namespace, "/", name)
	c.serviceManager.UpdateService(bsvc)
	return nil
}

func (c *Controller) EndpointHandler(name string, namespace string) error {
	endpointSlice, err := c.endpointsLister.EndpointSlices(namespace).Get(name)
	// TODO 创建新的svc
	bsvc := manager.NewService()
	if err != nil || !endpointSlice.ObjectMeta.DeletionTimestamp.IsZero() {
		logrus.Info("endpointSlice is Deleted name:", name, ",namespace: ", namespace, ",err:", err)
		c.serviceManager.DeleteService(bsvc)
		return nil
	}
	logrus.Info("one endpointSlice had change,", c.cluster, "/", namespace, "/", name)
	c.serviceManager.UpdateService(bsvc)
	return nil
}
