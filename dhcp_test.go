package virtualbox

import (
	"testing"
)

func TestDhcpServer(t *testing.T) {
	vb := NewVBox(Config{})

	dhcp1 := DHCPServer{
		IPAddress:      "10.0.2.1",
		LowerIPAddress: "10.0.2.2",
		UpperIPAddress: "10.0.2.254",
		NetworkName:    "NatNetwork",
		NetworkMask:    "255.255.255.0",
		Enabled:        true,
	}

	if _, err := vb.AddDHCPServer(dhcp1); err != nil {
		t.Fatalf("add dhcp failed: %s", err.Error())
	}

	t.Log("dhcp server created")

	dhcp2, err := vb.DHCPInfo(dhcp1.NetworkName)
	if err != nil {
		t.Fatalf("info failed: %s", err.Error())
	}

	t.Log("get info")

	if !compare(dhcp1, *dhcp2) {
		t.Fatalf("not equals")
	}

	t.Log("correct info")

	if err := vb.RemoveDHCPServer(dhcp1.NetworkName); err != nil {
		t.Fatalf("remove failed: %s", err.Error())
	}

	t.Log("dhcp server removed")
}

func compare(dhcp1, dhcp2 DHCPServer) bool {
	if dhcp1.IPAddress != dhcp2.IPAddress ||
		dhcp1.Enabled != dhcp2.Enabled ||
		dhcp1.LowerIPAddress != dhcp2.LowerIPAddress ||
		dhcp1.UpperIPAddress != dhcp2.UpperIPAddress ||
		dhcp1.NetworkMask != dhcp2.NetworkMask ||
		dhcp1.NetworkName != dhcp2.NetworkName {
		return false
	}

	return true
}
