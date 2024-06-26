package utils

import (
	"bytes"
	"encoding/binary"
	"net"
)

func IPString2Int16(ip string) uint16 {
	var ipret uint16
	IP := net.ParseIP(ip)
	if IP == nil {
		return 0
	}
	if err := binary.Read(bytes.NewBuffer(IP.To4()), binary.BigEndian, &ipret); err != nil {
		return 0
	}
	return ipret
}
