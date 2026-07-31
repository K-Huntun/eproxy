package bpf

import (
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
	"github.com/eproxy/pkg/cgroups"
	"github.com/sirupsen/logrus"
)

type BPFManager struct {
	ebpffile   string
	cglinks    []link.Link
	collection *ebpf.Collection
	service    *ebpf.Map
	endpoint   *ebpf.Map
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
	CheckOrMountBtfFS()
	spec, err := ebpf.LoadCollectionSpec(bm.ebpffile)
	if err != nil {
		return err
	}
	bm.collection, err = ebpf.NewCollectionWithOptions(spec, ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{PinPath: EProxyPath()},
	})
	if err != nil {
		return err
	}
	programs := []struct {
		name   string
		attach ebpf.AttachType
	}{
		{
			name:   "connect4",
			attach: ebpf.AttachCGroupInet4Connect,
		},
		{
			name:   "sendmsg4",
			attach: ebpf.AttachCGroupUDP4Sendmsg,
		},
	}
	for _, prog := range programs {
		program := bm.collection.Programs[prog.name]
		if program == nil {
			continue
		}
		cglink, attachErr := link.AttachCgroup(link.CgroupOptions{
			Path:    cgroups.GetCgroupRoot(),
			Program: program,
			Attach:  prog.attach,
		})
		if attachErr != nil {
			bm.Close()
			return attachErr
		}
		bm.cglinks = append(bm.cglinks, cglink)
	}
	bm.service = bm.collection.Maps["eproxy_lb4_services"]
	bm.endpoint = bm.collection.Maps["eproxy_lb4_backends"]
	logrus.Info("maps: ", bm.collection.Maps)
	return err
}

func (bm *BPFManager) Link() link.Link {
	if len(bm.cglinks) == 0 {
		return nil
	}
	return bm.cglinks[0]
}

func (bm *BPFManager) ServiceMap() *ebpf.Map {
	return bm.service
}

func (bm *BPFManager) EndpointMap() *ebpf.Map {
	return bm.endpoint
}

func (bm *BPFManager) Close() error {
	var retErr error
	for _, cglink := range bm.cglinks {
		if cglink == nil {
			continue
		}
		if err := cglink.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}
	bm.cglinks = nil
	if bm.collection != nil {
		bm.collection.Close()
	}
	return retErr
}
