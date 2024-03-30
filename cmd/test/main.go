// Copyright (c) 2016-2017 ByteDance, Inc. All rights reserved.
package main

import (
	"fmt"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
	"github.com/eproxy/pkg/cgroups"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)
	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal(err)
	}
	// mount group2
	cgroups.CheckOrMountCgrpFS("")
	//
	coll, err := ebpf.LoadCollection(os.Args[1])
	if err != nil {
		log.Fatalf("loading objects: %v", err)
	}
	defer coll.Close()
	/*
		5.7 以下 RawAttachProgram
		5.7以上 AttachRawLink
	*/
	// Attach ebpf program to a cgroupv2
	fmt.Println(coll.Programs["connect4"].FD())
	time.Sleep(1 * time.Second)

	connectorlink, err := link.AttachCgroup(link.CgroupOptions{
		Path:    cgroups.GetCgroupRoot(),
		Program: coll.Programs["connect4"],
		Attach:  ebpf.AttachCGroupInet4Connect,
	})
	defer connectorlink.Close()

	for {
		select {
		case <-stopper:
		default:
			time.Sleep(2 * time.Second)
			fmt.Println(connectorlink.Info())
		}
	}
}
