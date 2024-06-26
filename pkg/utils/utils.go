package utils

import (
	"bytes"
	"encoding/binary"
	"net"
)

func IPString2Int32(ip string) uint32 {
	var ipret uint32
	IP := net.ParseIP(ip)
	if IP == nil {
		return 0
	}
	if err := binary.Read(bytes.NewBuffer(IP.To4()), binary.BigEndian, &ipret); err != nil {
		return 0
	}
	return ipret
}
