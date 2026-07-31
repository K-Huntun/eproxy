package manager

import (
	"encoding/binary"
	"net"
	"sort"
	"strings"
)

var ignoredNodePortInterfacePrefixes = []string{
	"lo",
	"docker",
	"cni",
	"veth",
	"lxc",
	"flannel",
	"cali",
	"kube-ipvs0",
	"dummy",
	"virbr",
}

func LocalNodeIPv4Addrs() ([]uint32, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	addrSet := map[uint32]struct{}{}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if shouldIgnoreNodePortInterface(iface.Name) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch value := addr.(type) {
			case *net.IPNet:
				ip = value.IP
			case *net.IPAddr:
				ip = value.IP
			default:
				continue
			}
			ipv4 := ip.To4()
			if ipv4 == nil || !ipv4.IsGlobalUnicast() {
				continue
			}
			addrSet[binary.LittleEndian.Uint32(ipv4)] = struct{}{}
		}
	}

	addrs := make([]uint32, 0, len(addrSet))
	for addr := range addrSet {
		addrs = append(addrs, addr)
	}
	sort.Slice(addrs, func(i, j int) bool {
		return addrs[i] < addrs[j]
	})
	return addrs, nil
}

func shouldIgnoreNodePortInterface(name string) bool {
	for _, prefix := range ignoredNodePortInterfacePrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
