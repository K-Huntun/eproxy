package bpf

import (
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
	"github.com/eproxy/pkg/cgroups"
)

type BPFManager struct {
	ebpffile   string
	cglink     link.Link
	collection *ebpf.Collection
}

func NewBPFManager(file string) *BPFManager {
	return &BPFManager{
		ebpffile: file,
	}
}

func (bm *BPFManager) LoadAndAttach() error {
	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		return err
	}
	// mount group2
	cgroups.CheckOrMountCgrpFS("")
	spec, err := ebpf.LoadCollectionSpec(bm.ebpffile)
	if err != nil {
		return err
	}
	bm.collection, err = ebpf.NewCollectionWithOptions(spec, ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{PinPath: (EProxyPath())},
	})
	if err != nil {
		return err
	}
	// Attach ebpf program to a cgroupv2
	//fmt.Println(coll.Programs["connect4"].FD())
	bm.cglink, err = link.AttachCgroup(link.CgroupOptions{
		Path:    cgroups.GetCgroupRoot(),
		Program: bm.collection.Programs["connect4"],
		Attach:  ebpf.AttachCGroupInet4Connect,
	})
	return err
}

func (bm *BPFManager) Link() link.Link {
	return bm.cglink
}

func (bm *BPFManager) Close() error {
	err := bm.cglink.Close()
	bm.collection.Close()
	return err
}
