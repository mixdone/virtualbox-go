package virtualbox

import (
	"testing"
)

func checkEquality(n *NatNetwork, nat *NatNetwork) bool {
	if (n.NetName == nat.NetName) && (n.Network == nat.Network) && (n.DHCP == nat.DHCP) && (n.Ipv6 == nat.Ipv6) &&
		(n.Enabled == nat.Enabled) && (len(n.PortForward4) == len(nat.PortForward4)) && (len(n.PortForward6) == len(nat.PortForward6)) {
		c := 0
		for i := 0; i < len(n.PortForward4); i++ {
			if (n.PortForward4[i].Name == nat.PortForward4[i].Name) && (string(n.PortForward4[i].Protocol) == string(nat.PortForward4[i].Protocol)) &&
				(n.PortForward4[i].HostIP == nat.PortForward4[i].HostIP) && (n.PortForward4[i].HostPort == nat.PortForward4[i].HostPort) &&
				(n.PortForward4[i].GuestIP == nat.PortForward4[i].GuestIP) && (n.PortForward4[i].GuestPort == nat.PortForward4[i].GuestPort) {
				c++
			}
		}
		for i := 0; i < len(n.PortForward6); i++ {
			if (n.PortForward6[i].Name == nat.PortForward6[i].Name) && (string(n.PortForward6[i].Protocol) == string(nat.PortForward6[i].Protocol)) &&
				(n.PortForward6[i].HostIP == nat.PortForward6[i].HostIP) && (n.PortForward6[i].HostPort == nat.PortForward6[i].HostPort) &&
				(n.PortForward6[i].GuestIP == nat.PortForward6[i].GuestIP) && (n.PortForward6[i].GuestPort == nat.PortForward6[i].GuestPort) {
				c++
			}
		}
		if c == (len(n.PortForward4) + len(n.PortForward6)) {
			return true
		}
	}
	return false
}

func TestNatNetwork(t *testing.T) {
	vb := NewVBox(Config{})

	nat := NatNetwork{}
	nat.NetName = "TestNatNet"
	nat.Network = "192.168.10.0/24"
	nat.Enabled = true
	nat.DHCP = true
	nat.Ipv6 = false
	nat.PortForward4 = make([]PortForwarding, 0, 5)
	nat.PortForward6 = make([]PortForwarding, 0, 5)

	var rule1 PortForwarding
	rule1.Name = "rule1"
	rule1.Protocol = TCP
	rule1.HostIP = ""
	rule1.HostPort = 1024
	rule1.GuestIP = "192.168.10.5"
	rule1.GuestPort = 22

	var rule2 PortForwarding
	rule2.Name = "rule2"
	rule2.Protocol = TCP
	rule2.HostIP = ""
	rule2.HostPort = 1022
	rule2.GuestIP = "192.168.11.5"
	rule2.GuestPort = 25
	nat.PortForward4 = append(nat.PortForward4, rule1, rule2)

	if err := vb.AddNatNet(&nat); err != nil {
		t.Fatalf("Failed creating NAT network %v", err)
	}

	natnws, err := vb.ListNatNets()
	if err != nil {
		t.Fatalf("Failed getting list of all NAT networks %v", err)
	}

	ok := false
	for _, n := range natnws {
		if ok = checkEquality(&n, &nat); ok {
			t.Logf("NAT network with name %s created", nat.NetName)
		}
	}
	if !ok {
		t.Fatalf("The created NAT network with name %s could not be found in the list of all NAT networks", nat.NetName)
	}

	nat.Ipv6 = true
	nat.Network = "192.160.0.0/24"
	if err := vb.ModifyNatNet(&nat, []string{"ipv6", "network"}); err != nil {
		t.Fatalf("Failed to modify NAT network %v", err)
	}

	nat.PortForward4 = make([]PortForwarding, 0, 5)
	nat.PortForward6 = make([]PortForwarding, 0, 5)

	var rule3 PortForwarding
	rule3.Name = "rule3"
	rule3.Protocol = UDP
	rule3.HostIP = ""
	rule3.HostPort = 1030
	rule3.GuestIP = "192.168.13.5"
	rule3.GuestPort = 27

	if err = vb.AddAllPortForwNat(&nat, []PortForwarding{rule3}, "--port-forward-4"); err != nil {
		t.Fatalf("Failed to add all port forwarding %v", err)
	}

	if err = vb.DeleteAllPortForwNat(&nat, []PortForwarding{rule1, rule2}, "--port-forward-4"); err != nil {
		t.Fatalf("Failed to delete all port forwarding %v", err)
	}

	nat.PortForward4 = append(nat.PortForward4, rule3)

	var rule4 PortForwarding
	rule4.Name = "rule4"
	rule4.Protocol = UDP
	rule4.HostIP = ""
	rule4.HostPort = 1024
	rule4.GuestIP = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	rule4.GuestPort = 22

	nat.PortForward6 = append(nat.PortForward6, rule4)

	if err = vb.AddAllPortForwNat(&nat, []PortForwarding{rule4}, "--port-forward-6"); err != nil {
		t.Fatalf("Failed to add all port forwarding %v", err)
	}

	natnws, err = vb.ListNatNets()
	if err != nil {
		t.Fatalf("Failed getting list of all NAT networks %v", err)
	}

	ok = false
	for _, n := range natnws {
		if ok = checkEquality(&n, &nat); ok {
			t.Logf("NAT network with name %s modified", nat.NetName)
		}
	}
	if !ok {
		t.Fatalf("The NAT network with name %s has not been modified", nat.NetName)
	}

	if err := vb.StartNatNet(&nat); err != nil {
		t.Fatalf("Failed starting NAT network %v", err)
	}
	t.Logf("NAT network with name %s started", nat.NetName)

	if err := vb.StopNatNet(&nat); err != nil {
		t.Fatalf("Failed stopping NAT network %v", err)
	}
	t.Logf("NAT network with name %s stopped", nat.NetName)

	if err := vb.RemoveNatNet(&nat); err != nil {
		t.Fatalf("Failed removing NAT network %v", err)
	}
	t.Logf("NAT network with name %s removed", nat.NetName)
}
