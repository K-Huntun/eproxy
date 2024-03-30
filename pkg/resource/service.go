// Copyright (c) 2016-2017 ByteDance, Inc. All rights reserved.
package resource

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"time"
)

type ServiceInformer struct {
	lw                *cache.ListWatch
	indexer           cache.Indexer
	controller        cache.Controller
	eventhandler      cache.ResourceEventHandler
	resyncPeriod      time.Duration
	watchErrorHandler cache.WatchErrorHandler
	started, stopped  bool
}

func (informer *ServiceInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, resyncPeriod time.Duration) {
	informer.eventhandler = handler
	informer.resyncPeriod = resyncPeriod
}

func (informer *ServiceInformer) GetStore() cache.Store {
	return informer.indexer
}

func (informer *ServiceInformer) GetController() cache.Controller {
	return informer.controller
}

func (informer *ServiceInformer) LastSyncResourceVersion() string {
	return informer.controller.LastSyncResourceVersion()
}

func (informer *ServiceInformer) SetWatchErrorHandler(handler cache.WatchErrorHandler) error {
	informer.watchErrorHandler = handler
	return nil
}

func (informer *ServiceInformer) AddIndexers(indexers cache.Indexers) error {
	return informer.indexer.AddIndexers(indexers)
}

func (informer *ServiceInformer) GetIndexer() cache.Indexer {
	return informer.indexer
}

func convertToCustomService(obj interface{}) interface{} {
	switch concreteObj := obj.(type) {
	case *v1.Service:
		p := &v1.Service{
			TypeMeta: concreteObj.TypeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name:      concreteObj.Name,
				Namespace: concreteObj.Namespace,
				Labels:    concreteObj.Labels,
			},
			Spec: v1.ServiceSpec{
				Ports: concreteObj.Spec.Ports,
				Type:  concreteObj.Spec.Type,
			},
		}
		*concreteObj = v1.Service{}
		return p
	case cache.DeletedFinalStateUnknown:
		service, ok := concreteObj.Obj.(*v1.Service)
		if !ok {
			return obj
		}
		dfsu := cache.DeletedFinalStateUnknown{
			Key: concreteObj.Key,
			Obj: &v1.Service{
				TypeMeta: service.TypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:              service.Name,
					Namespace:         service.Namespace,
					ResourceVersion:   service.ResourceVersion,
					DeletionTimestamp: service.DeletionTimestamp,
					Annotations:       service.Annotations,
					OwnerReferences:   service.OwnerReferences,
					Labels:            service.Labels,
				},
				Spec: v1.ServiceSpec{
					Ports: service.Spec.Ports,
					Type:  service.Spec.Type,
				},
				Status: v1.ServiceStatus{},
			},
		}
		*service = v1.Service{}
		return dfsu
	default:
		return obj
	}
}

func (informer *ServiceInformer) AddEventHandler(handle cache.ResourceEventHandler) {
	informer.eventhandler = handle
}

func (informer *ServiceInformer) HasSynced() bool {
	if informer.controller == nil {
		return false
	}
	return informer.controller.HasSynced()
}

func ComparesLabes(map1, map2 map[string]string) bool {
	return false
}

func (informer *ServiceInformer) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	fifo := cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{
		KnownObjects:          informer.indexer,
		EmitDeltaTypeReplaced: true,
	})
	cfg := &cache.Config{
		Queue:             fifo,
		ListerWatcher:     informer.lw,
		ObjectType:        &v1.Service{},
		FullResyncPeriod:  0,
		RetryOnError:      false,
		WatchErrorHandler: informer.watchErrorHandler,

		Process: func(obj interface{}) error {
			for _, d := range obj.(cache.Deltas) {
				var obj interface{}
				obj = convertToCustomService(d.Object)
				switch d.Type {
				case cache.Sync, cache.Added, cache.Updated, cache.Replaced:
					if _, exists, err := informer.indexer.Get(obj); err == nil && exists {
						// 无需关注 service变化
						// Selector将由 Endpoint支持
						if err := informer.indexer.Update(obj); err != nil {
							return err
						}
					} else {
						if err := informer.indexer.Add(obj); err != nil {
							return err
						}
						//informer.eventhandler.OnAdd(obj)
					}
				case cache.Deleted:
					if err := informer.indexer.Delete(obj); err != nil {
						return err
					}
					//informer.eventhandler.OnDelete(obj)
				}
			}
			return nil
		},
	}

	func() {
		informer.controller = cache.New(cfg)
		informer.started = true
	}()
	defer func() {
		informer.stopped = true // Don't want any new listeners
	}()
	informer.controller.Run(stopCh)
}

//func MetaNamespaceIndexFunc(obj interface{}) ([]string, error) {
//	meta, err := meta.Accessor(obj)
//	if err != nil {
//		return []string{""}, fmt.Errorf("object has no meta: %v", err)
//	}
//	return []string{meta.GetNamespace()}, nil
//}

var _ cache.SharedIndexInformer = &ServiceInformer{}

func defaultCustomServiceInformer(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	indexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return client.CoreV1().Services(v1.NamespaceAll).List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return client.CoreV1().Services(v1.NamespaceAll).Watch(context.TODO(), options)
		},
	}
	return &ServiceInformer{
		lw:           lw,
		resyncPeriod: resyncPeriod,
		indexer:      indexer,
	}
}
