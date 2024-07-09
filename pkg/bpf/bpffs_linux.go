package bpf

import (
	"fmt"
	"github.com/eproxy/pkg/defaults"
	"github.com/eproxy/pkg/mountinfo"
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
	MkdirBPF(filepath.Join(defaults.BPFFSRoot, "eproxy"))
	return nil
}
func MkdirBPF(path string) error {
	return os.MkdirAll(path, 0755)
}

func checkOrMountDefaultLocations() error {
	// Check whether /sys/fs/bpf has a BPFFS mount.
	mounted, bpffsInstance, err := mountinfo.IsMountFS(mountinfo.FilesystemTypeBPFFS, defaults.BPFFSRoot)
	if err != nil {
		return err
	}

	// If /sys/fs/bpf is not mounted at all, we should mount
	// BPFFS there.
	if !mounted {
		if err := mountFS(false); err != nil {
			return err
		}

		return nil
	}

	if !bpffsInstance {
		// If /sys/fs/bpf has a mount but with some other filesystem
		// than BPFFS, it means that Cilium is running inside container
		// and /sys/fs/bpf is not mounted on host. We should mount BPFFS
		// in /run/cilium/bpffs automatically. This will allow operation
		// of Cilium but will result in unmounting of the filesystem
		// when the pod is restarted. This in turn will cause resources
		// such as the connection tracking table of the BPF programs to
		// be released which will cause all connections into local
		// containers to be dropped. User is going to be warned.
		log.Warnf("BPF filesystem is going to be mounted automatically "+
			"in %s. However, it probably means that Cilium is running "+
			"inside container and BPFFS is not mounted on the host. "+
			"for more information, see: https://cilium.link/err-bpf-mount",
			defaults.BPFFSRootFallback,
		)

		cMounted, cBpffsInstance, err := mountinfo.IsMountFS(mountinfo.FilesystemTypeBPFFS, defaults.BPFFSRoot)
		if err != nil {
			return err
		}
		if !cMounted {
			if err := mountFS(false); err != nil {
				return err
			}
		} else if !cBpffsInstance {
			log.Fatalf("%s is mounted but has a different filesystem than BPFFS", defaults.BPFFSRootFallback)
		}
	}

	log.Infof("Detected mounted BPF filesystem at %s", defaults.BPFFSRoot)

	return nil
}

func CheckOrMountBtfFS() {
	if err := checkOrMountDefaultLocations(); err != nil {
		log.Errorf("Failed to mount btffs:%v", err)
	} else {
		log.Infof("Mounted btffs filesystem ")
	}
}
