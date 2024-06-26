// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package cgroups

import (
	"github.com/sirupsen/logrus"
	"sync"

	"github.com/eproxy/pkg/defaults"
)

var (
	// Path to where cgroup is mounted
	cgroupRoot = defaults.DefaultCgroupRoot

	// Only mount a single instance
	cgrpMountOnce sync.Once
)

// setCgroupRoot will set the path to mount cgroupv2
func setCgroupRoot(path string) {
	cgroupRoot = path
}

// GetCgroupRoot returns the path for the cgroupv2 mount
func GetCgroupRoot() string {
	return cgroupRoot
}

// CheckOrMountCgrpFS this checks if the cilium cgroup2 root mount point is
// mounted and if not mounts it. If mapRoot is "" it will mount the default
// location. It is harmless to have multiple cgroupv2 root mounts so unlike
// BPFFS case we simply mount at the cilium default regardless if the system
// has another mount created by systemd or otherwise.
func CheckOrMountCgrpFS(mapRoot string) {
	cgrpMountOnce.Do(func() {
		if mapRoot == "" {
			mapRoot = cgroupRoot
		}

		if err := cgrpCheckOrMountLocation(mapRoot); err != nil {
			logrus.Errorf("Failed to mount cgroupv2. Any functionality that needs cgroup (e.g.: socket-based LB) will not work.")
			//log.WithError(err).
			//	Warn("Failed to mount cgroupv2. Any functionality that needs cgroup (e.g.: socket-based LB) will not work.")
		} else {
			logrus.Infof("Mounted cgroupv2 filesystem at %s", mapRoot)
		}
	})
}
