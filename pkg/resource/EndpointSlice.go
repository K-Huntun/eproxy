package resource

import (
	"context"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"time"
)

const (
	// LabelServiceProxyName indicates that an alternative service
	// proxy will implement this Service.
	LabelServiceProxyName = "service.kubernetes.io/service-proxy-name"
	ProxyName             = "eproxy"
)

type EndpointSliceInformer struct {
	lw                *cache.ListWatch
	indexer           cache.Indexer
	controller        cache.Controller
	eventhandler      cache.ResourceEventHandler
	resyncPeriod      time.Duration
	watchErrorHandler cache.WatchErrorHandler
	started, stopped  bool
}

func (informer *EndpointSliceInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, resyncPeriod time.Duration) {
	informer.eventhandler = handler
	informer.resyncPeriod = resyncPeriod
}

func (informer *EndpointSliceInformer) GetStore() cache.Store {
	return informer.indexer
}

func (informer *EndpointSliceInformer) GetController() cache.Controller {
	return informer.controller
}

func (informer *EndpointSliceInformer) LastSyncResourceVersion() string {
	return informer.controller.LastSyncResourceVersion()
}

func (informer *EndpointSliceInformer) SetWatchErrorHandler(handler cache.WatchErrorHandler) error {
	informer.watchErrorHandler = handler
	return nil
}

func (informer *EndpointSliceInformer) AddIndexers(indexers cache.Indexers) error {
	return informer.indexer.AddIndexers(indexers)
}

func (informer *EndpointSliceInformer) GetIndexer() cache.Indexer {
	return informer.indexer
}

func convertToCustomEndpointSlice(obj interface{}) interface{} {
	switch concreteObj := obj.(type) {
	case *discovery.EndpointSlice:
		p := &discovery.EndpointSlice{
			TypeMeta: concreteObj.TypeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name:              concreteObj.Name,
				Namespace:         concreteObj.Namespace,
				ResourceVersion:   concreteObj.ResourceVersion,
				DeletionTimestamp: concreteObj.DeletionTimestamp,
				Annotations:       concreteObj.Annotations,
				OwnerReferences:   concreteObj.OwnerReferences,
				Labels:            concreteObj.Labels,
			},
		}
		*concreteObj = discovery.EndpointSlice{}
		return p
	case cache.DeletedFinalStateUnknown:
		EndpointSlice, ok := concreteObj.Obj.(*discovery.EndpointSlice)
		if !ok {
			return obj
		}
		dfsu := cache.DeletedFinalStateUnknown{
			Key: concreteObj.Key,
			Obj: &discovery.EndpointSlice{
				TypeMeta: EndpointSlice.TypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:              EndpointSlice.Name,
					Namespace:         EndpointSlice.Namespace,
					ResourceVersion:   EndpointSlice.ResourceVersion,
					DeletionTimestamp: EndpointSlice.DeletionTimestamp,
					Annotations:       EndpointSlice.Annotations,
					OwnerReferences:   EndpointSlice.OwnerReferences,
					Labels:            EndpointSlice.Labels,
				},
			},
		}
		*EndpointSlice = discovery.EndpointSlice{}
		return dfsu
	default:
		return obj
	}
}

func (informer *EndpointSliceInformer) AddEventHandler(handle cache.ResourceEventHandler) {
	informer.eventhandler = handle
}

func (informer *EndpointSliceInformer) HasSynced() bool {
	if informer.controller == nil {
		return false
	}
	return informer.controller.HasSynced()
}

func IsEndpointSliceReady(EndpointSlice *discovery.EndpointSlice) bool {
	if EndpointSlice.DeletionTimestamp != nil {
		return false
	}
	// Read状态为true
	return false
}

func (informer *EndpointSliceInformer) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	fifo := cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{
		KnownObjects:          informer.indexer,
		EmitDeltaTypeReplaced: true,
	})
	cfg := &cache.Config{
		Queue:             fifo,
		ListerWatcher:     informer.lw,
		ObjectType:        &discovery.EndpointSlice{},
		FullResyncPeriod:  0,
		RetryOnError:      false,
		WatchErrorHandler: informer.watchErrorHandler,

		Process: func(obj interface{}) error {
			for _, d := range obj.(cache.Deltas) {
				var obj interface{}
				obj = convertToCustomEndpointSlice(d.Object)
				switch d.Type {
				case cache.Sync, cache.Added, cache.Updated, cache.Replaced:
					if old, exists, err := informer.indexer.Get(obj); err == nil && exists {
						if err := informer.indexer.Update(obj); err != nil {
							return err
						}
						informer.eventhandler.OnUpdate(old, obj)
					} else {
						if err := informer.indexer.Add(obj); err != nil {
							return err
						}
						informer.eventhandler.OnAdd(obj)
					}
				case cache.Deleted:
					if err := informer.indexer.Delete(obj); err != nil {
						return err
					}
					informer.eventhandler.OnDelete(obj)
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

var _ cache.SharedIndexInformer = &EndpointSliceInformer{}

func defaultCustomEndpointSliceInformer(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	eProxyName, err := labels.NewRequirement(LabelServiceProxyName, selection.Equals, []string{ProxyName})
	if err != nil {
		panic(err)
	}
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*eProxyName)
	indexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.LabelSelector = labelSelector.String()
			return client.DiscoveryV1().EndpointSlices(v1.NamespaceAll).List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.LabelSelector = labelSelector.String()
			return client.DiscoveryV1().EndpointSlices(v1.NamespaceAll).Watch(context.TODO(), options)
		},
	}
	return &EndpointSliceInformer{
		lw:           lw,
		resyncPeriod: resyncPeriod,
		indexer:      indexer,
	}
}
