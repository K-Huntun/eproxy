package bpf

import (
	"fmt"
	"github.com/eproxy/pkg/defaults"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
)

// CiliumPath returns the bpffs path to be used for Cilium object pins.
func EProxyPath() string {
	return filepath.Join(defaults.BPFFSRoot, "eproxy")
}

// mountFS mounts the BPFFS filesystem into the desired mapRoot directory.
func mountFS(printWarning bool) error {
	if printWarning {
		log.Warning("================================= WARNING ==========================================")
		log.Warning("BPF filesystem is not mounted. This will lead to network disruption when eProxy pods")
		log.Warning("are restarted. Ensure that the BPF filesystem is mounted in the host.")
		log.Warning("====================================================================================")
	}

	log.Infof("Mounting BPF filesystem at %s", defaults.BPFFSRoot)

	mapRootStat, err := os.Stat(defaults.BPFFSRoot)
	if err != nil {
		if os.IsNotExist(err) {
			if err := MkdirBPF(defaults.BPFFSRoot); err != nil {
				return fmt.Errorf("unable to create bpf mount directory: %w", err)
			}
		} else {
			return fmt.Errorf("failed to stat the mount path %s: %w", defaults.BPFFSRoot, err)

		}
	} else if !mapRootStat.IsDir() {
		return fmt.Errorf("%s is a file which is not a directory", defaults.BPFFSRoot)
	}

	if err := unix.Mount(defaults.BPFFSRoot, defaults.BPFFSRoot, "bpf", 0, ""); err != nil {
		return fmt.Errorf("failed to mount %s: %w", defaults.BPFFSRoot, err)
	}
	return nil
}
func MkdirBPF(path string) error {
	return os.MkdirAll(path, 0755)
}
