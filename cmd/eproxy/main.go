// Copyright (c) 2016-2017 ByteDance, Inc. All rights reserved.
package main

import (
	"flag"
	"github.com/eproxy/pkg/bpf"
	"github.com/eproxy/pkg/manager"
	"github.com/eproxy/pkg/resource"
	"github.com/eproxy/pkg/signals"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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
	var client *kubernetes.Clientset
	StopCh := signals.SetupSignalHandler()
	bm := bpf.NewBPFManager(ebpffile)
	err := bm.LoadAndAttach()
	if err != nil {
		logrus.Fatal(err)
		return
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		logrus.Fatal(err)
		return
	}
	if client, err = kubernetes.NewForConfig(config); err != nil {
		logrus.Error("create k8s client error: ", err)
	}
	k8sresource := resource.NewResources(client)

	svcmgr := manager.NewServiceManager(bm.ServiceMap(), bm.EndpointMap())

	k8sresource.SetEndpointHandler(&resource.EndpointSliceAdapterHandler{svcmgr})
	k8sresource.SetServiceHandler(&resource.ServiceAdapterHandler{svcmgr})

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
