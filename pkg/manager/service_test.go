package manager

import (
	"fmt"
	"testing"
)

func TestIpParse(t *testing.T) {
	var testip = "192.168.56.2"
	fmt.Println(IP(testip).Address())
	var testip2 = "fe80::8cee:d1ff:feb6:123d"
	fmt.Println(IP(testip2).Address())
}
