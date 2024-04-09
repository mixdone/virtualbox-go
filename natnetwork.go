package virtualbox

import (
	"errors"
	"strings"
)

func (vb *VBox) AddNatNet(nat NatNetwork) error {
	args := []string{"natnetwork", "add", "--netname", nat.NetName, "--network", nat.Network}
	if !nat.Enabled {
		args = append(args, "--disable")
	}
	if !nat.DHCP {
		args = append(args, "--dhcp", "off")
	}
	if nat.Ipv6 {
		args = append(args, "--ipv6", "on")
	}
	if nat.Loopback4 != "" {
		args = append(args, "--loopback-4", nat.Loopback4)
	}
	if nat.Loopback6 != "" {
		args = append(args, "--loopback-6", nat.Loopback6)
	}
	if nat.PortForward4 != "" {
		args = append(args, "--port-forward-4", nat.PortForward4)
	}
	if nat.PortForward6 != "" {
		args = append(args, "--port-forward-6", nat.PortForward6)
	}
	_, err := vb.manage(args...)
	return err
}

func (vb *VBox) RemoveNatNet(nat NatNetwork) error {
	args := []string{"natnetwork", "remove", "--netname", nat.NetName}
	_, err := vb.manage(args...)
	return err
}

func (vb *VBox) StartNatNet(nat NatNetwork) error {
	args := []string{"natnetwork", "start", "--netname", nat.NetName}
	_, err := vb.manage(args...)
	return err
}

func (vb *VBox) StopNatNet(nat NatNetwork) error {
	args := []string{"natnetwork", "stop", "--netname", nat.NetName}
	_, err := vb.manage(args...)
	return err
}

func (vb *VBox) listNatNets() ([]NatNetwork, error) {
	out, err := vb.manage("natnetwork", "list")
	if err != nil {
		return nil, err
	}

	var natnws []NatNetwork

	var nw NatNetwork
	_ = tryParseKeyValues(out, reColonLine, func(key, val string, ok bool) error {
		switch key {
		case "Name":
			nw.NetName = val
		case "Network":
			nw.Network = val
		case "DHCP Server":
			if val == "No" {
				nw.DHCP = false
			} else {
				nw.DHCP = true
			}
		case "IPv6":
			if val == "No" {
				nw.Ipv6 = false
			} else {
				nw.Ipv6 = true
			}
		case "Enabled":
			if val == "No" {
				nw.Enabled = false
			} else {
				nw.Enabled = true
			}
		default:
			if !ok && strings.TrimSpace(val) == "" {
				natnws = append(natnws, nw)
				nw = NatNetwork{}
			}
		}
		return nil
	})
	return natnws, nil
}

func (vb *VBox) ModifyNatNet(nat NatNetwork, parameters []string) error {
	if len(parameters) == 0 {
		return errors.New("no parameters to change")
	}
	args := []string{"natnetwork", "modify", "--netname", nat.NetName}
	for _, s := range parameters {
		switch s {
		case "network":
			args = append(args, "--network", nat.Network)
		case "enabled":
			if nat.Enabled {
				args = append(args, "--enable")
			} else {
				args = append(args, "--disable")
			}
		case "DHCP":
			if nat.DHCP {
				args = append(args, "--dhcp", "on")
			} else {
				args = append(args, "--dhcp", "off")
			}
		case "ipv6":
			if nat.DHCP {
				args = append(args, "--ipv6", "on")
			} else {
				args = append(args, "--ipv6", "off")
			}
		case "loopback4":
			args = append(args, "--loopback-4", nat.Loopback4)
		case "loopback6":
			args = append(args, "--loopback-6", nat.Loopback6)
		case "portforward4":
			args = append(args, "--port-forward-4", nat.PortForward4)
		case "portforward6":
			args = append(args, "--port-forward-6", nat.PortForward6)
		default:
			return errors.New("invalid parameter in the arguments")
		}
	}
	_, err := vb.manage(args...)
	return err
}
