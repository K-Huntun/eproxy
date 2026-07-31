package manager

import (
	"encoding/binary"
	"net"
	"testing"

	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestNewServiceAggregatesEndpointsAndNodePorts(t *testing.T) {
	ready := true
	tcp := v1.ProtocolTCP

	svc := &v1.Service{
		Spec: v1.ServiceSpec{
			ClusterIP: "10.0.0.1",
			Ports: []v1.ServicePort{
				{
					Name:       "dns",
					Protocol:   v1.ProtocolUDP,
					Port:       53,
					TargetPort: intstr.FromInt(5353),
					NodePort:   30053,
				},
				{
					Name:       "web",
					Protocol:   v1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromString("web"),
					NodePort:   30080,
				},
			},
		},
	}

	slices := []*discoveryv1.EndpointSlice{
		{
			Ports: []discoveryv1.EndpointPort{
				{Name: strPtr("web"), Protocol: &tcp, Port: int32Ptr(8080)},
			},
			Endpoints: []discoveryv1.Endpoint{
				{Conditions: discoveryv1.EndpointConditions{Ready: &ready}, Addresses: []string{"192.168.1.10"}},
			},
		},
		{
			Ports: []discoveryv1.EndpointPort{
				{Name: strPtr("web"), Protocol: &tcp, Port: int32Ptr(8080)},
			},
			Endpoints: []discoveryv1.Endpoint{
				{Conditions: discoveryv1.EndpointConditions{Ready: &ready}, Addresses: []string{"192.168.1.11", "192.168.1.10"}},
			},
		},
	}

	nodePortIPs := []uint32{
		binary.LittleEndian.Uint32(net.ParseIP("172.16.0.10").To4()),
	}

	got := NewService(svc, slices, nodePortIPs)

	if len(got.Endpoints) != 2 {
		t.Fatalf("expected 2 unique endpoints, got %d", len(got.Endpoints))
	}
	if got.Ports[0].TargetPort != 5353 {
		t.Fatalf("expected udp target port 5353, got %d", got.Ports[0].TargetPort)
	}
	if got.Ports[1].TargetPort != 8080 {
		t.Fatalf("expected tcp target port 8080 from endpointslice, got %d", got.Ports[1].TargetPort)
	}
	if len(got.NodePortIPs) != 1 || got.NodePortIPs[0] != nodePortIPs[0] {
		t.Fatalf("unexpected node port IPs: %#v", got.NodePortIPs)
	}
}

func strPtr(value string) *string {
	return &value
}

func int32Ptr(value int32) *int32 {
	return &value
}
