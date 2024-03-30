// Copyright (c) 2016-2017 ByteDance, Inc. All rights reserved.
package main

import (
	"flag"
	"github.com/eproxy/pkg/manager"
	"github.com/eproxy/pkg/resource"
	"github.com/eproxy/pkg/signals"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"os"
)

var (
	Prof       bool
	configfile string
	help       bool
	version    bool
)

func ParseCommand() {
	flag.BoolVar(&help, "h", false, "help")
	flag.BoolVar(&Prof, "p", false, "pprof")
	flag.BoolVar(&version, "v", false, "version")
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
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		panic(err)
	}
	if client, err = kubernetes.NewForConfig(config); err != nil {
		logrus.Error("create k8s client error: ", err)
	}
	k8sresource := resource.NewResources(client)
	svcmgr := manager.NewServiceManager()

	k8sresource.SetEndpointHandler(&resource.EndpointSliceAdapterHandler{svcmgr})
	k8sresource.SetServiceHandler(&resource.ServiceAdapterHandler{svcmgr})

	k8sresource.StartListenEventFromKubernetes(StopCh)

}
