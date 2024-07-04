// Copyright (c) 2016-2017 ByteDance, Inc. All rights reserved.
package main

import (
	"flag"
	"github.com/eproxy/pkg/bpf"
	"github.com/eproxy/pkg/kubernetes/controller"
	"github.com/eproxy/pkg/kubernetes/informers"
	"github.com/eproxy/pkg/manager"
	"github.com/eproxy/pkg/utils/signals"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

var (
	Prof       bool
	configfile string
	kubeconfig string
	ebpffile   string
	help       bool
	version    bool
)

func ParseCommand() {
	flag.BoolVar(&help, "h", false, "help")
	flag.BoolVar(&Prof, "p", false, "pprof")
	flag.BoolVar(&version, "v", false, "version")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "kubeconfig path")
	flag.StringVar(&ebpffile, "ebpf", "", "ebpf file path")
	flag.StringVar(&configfile, "f", "", "config file")

	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(1)
	}

	if Prof {
		go http.ListenAndServe("localhost:6061", nil)
	}
}

func main() {
	ParseCommand()
	StopCh := signals.SetupSignalHandler()
	bm := bpf.NewBPFManager(ebpffile)
	err := bm.LoadAndAttach()
	if err != nil {
		logrus.Fatal(err)
		return
	}
	ServiceManager := manager.NewServiceManager(bm.ServiceMap(), bm.EndpointMap())

	k8sresource := informers.NewResources(kubeconfig)
	controller := controller.NewController(ServiceManager,
		k8sresource.KubernetetsClient(),
		k8sresource.ServiceInformer(),
		k8sresource.EndpointSliceInfomer())
	go controller.Run(1, StopCh)
	k8sresource.StartListenEventFromKubernetes(StopCh)
	for {
		select {
		case <-StopCh:
			logrus.Info("stop eproxy,close ebpf")
			bm.Close()
			return
		default:
			time.Sleep(10 * time.Second)
		}
	}
}
