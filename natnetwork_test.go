package virtualbox

import (
	"testing"
)

func TestNatNetwork(t *testing.T) {
	vb := NewVBox(Config{})
	nat := NatNetwork{}
	nat.NetName = "TestNatNet"
	nat.Network = "192.168.10.0/24"
	nat.Enabled = true
	nat.DHCP = true
	nat.Ipv6 = false
	nat.Loopback4 = ""
	nat.Loopback6 = ""
	nat.PortForward4 = ""
	nat.PortForward6 = ""

	if err := vb.AddNatNet(nat); err != nil {
		t.Fatalf("Failed creating NAT network %v", err)
	}

	natnws, err := vb.listNatNets()
	if err != nil {
		t.Fatalf("Failed getting list of all NAT networks %v", err)
	}

	ok := false
	for _, n := range natnws {
		if (n.NetName == "TestNatNet") && (n.Network == "192.168.10.0/24") && n.DHCP && !n.Ipv6 && n.Enabled {
			t.Logf("NAT network with name %s created", nat.NetName)
			ok = true
		}
	}
	if !ok {
		t.Fatalf("The created NAT network with name %s could not be found in the list of all NAT networks", nat.NetName)
	}

	nat.Ipv6 = true
	nat.PortForward4 = "ssh:tcp:[]:1022:[192.168.10.5]:22"
	nat.Network = "192.160.0.0/24"
	if err := vb.ModifyNatNet(nat, []string{"ipv6", "portforward4", "network"}); err != nil {
		t.Fatalf("Failed to modify NAT network %v", err)
	}

	natnws, err = vb.listNatNets()
	if err != nil {
		t.Fatalf("Failed getting list of all NAT networks %v", err)
	}
	ok = false
	for _, n := range natnws {
		if (n.NetName == "TestNatNet") && (n.Network == "192.160.0.0/24") && n.DHCP && n.Ipv6 && n.Enabled {
			t.Logf("NAT network with name %s modified", nat.NetName)
			ok = true
		}
	}
	if !ok {
		t.Fatalf("The NAT network with name %s has not been modified", nat.NetName)
	}

	if err := vb.StartNatNet(nat); err != nil {
		t.Fatalf("Failed starting NAT network %v", err)
	}
	t.Logf("NAT network with name %s started", nat.NetName)

	if err := vb.StopNatNet(nat); err != nil {
		t.Fatalf("Failed stopping NAT network %v", err)
	}
	t.Logf("NAT network with name %s stopped", nat.NetName)

	if err := vb.RemoveNatNet(nat); err != nil {
		t.Fatalf("Failed removing NAT network %v", err)
	}
	t.Logf("NAT network with name %s removed", nat.NetName)
}
