package virtualbox

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func (vb *VBox) AddNatNet(nat *NatNetwork) error {
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
	if _, err := vb.manage(args...); err != nil {
		return err
	}

	if len(nat.PortForward4) != 0 {
		if err := vb.AddAllPortForwNat(nat, nat.PortForward4, "--port-forward-4"); err != nil {
			return err
		}
	}
	if len(nat.PortForward6) != 0 {
		if err := vb.AddAllPortForwNat(nat, nat.PortForward6, "--port-forward-6"); err != nil {
			return err
		}
	}

	return nil
}

func (vb *VBox) AddAllPortForwNat(nat *NatNetwork, rule []PortForwarding, flag string) error {
	args := []string{"natnetwork", "modify", "--netname", nat.NetName}
	for i := 0; i < len(rule); i++ {
		args = append(args, flag, fmt.Sprintf("%v:%v:[%v]:%v:[%v]:%v", rule[i].Name, string(rule[i].Protocol),
			rule[i].HostIP, rule[i].HostPort, rule[i].GuestIP, rule[i].GuestPort))
	}
	_, err := vb.manage(args...)
	return err
}

func (vb *VBox) DeleteAllPortForwNat(nat *NatNetwork, rule []PortForwarding, flag string) error {
	args := []string{"natnetwork", "modify", "--netname", nat.NetName}
	for i := 0; i < len(rule); i++ {
		args = append(args, flag, "delete", rule[i].Name)
	}
	_, err := vb.manage(args...)
	return err
}

func (vb *VBox) RemoveNatNet(nat *NatNetwork) error {
	args := []string{"natnetwork", "remove", "--netname", nat.NetName}
	_, err := vb.manage(args...)
	return err
}

func (vb *VBox) StartNatNet(nat *NatNetwork) error {
	args := []string{"natnetwork", "start", "--netname", nat.NetName}
	_, err := vb.manage(args...)
	return err
}

func (vb *VBox) StopNatNet(nat *NatNetwork) error {
	args := []string{"natnetwork", "stop", "--netname", nat.NetName}
	_, err := vb.manage(args...)
	return err
}

func fillRule(ret []string) (PortForwarding, error) {
	var rule PortForwarding
	rule.Name = ret[1]
	if ret[2] == "tcp" {
		rule.Protocol = TCP
	} else {
		rule.Protocol = UDP
	}
	rule.HostIP = ret[3]
	i, err := strconv.Atoi(ret[4])
	if err != nil {
		return rule, err
	}
	rule.HostPort = i
	rule.GuestIP = ret[5]
	i, err = strconv.Atoi(ret[6])
	if err != nil {
		return rule, err
	}
	rule.GuestPort = i
	return rule, err
}

func parseKeyVal(key string, val string, nw *NatNetwork) {
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
	}
}

func (vb *VBox) ListNatNets() ([]NatNetwork, error) {
	out, err := vb.manage("natnetwork", "list")
	if err != nil {
		return nil, err
	}
	var reColonLine = regexp.MustCompile(`([^:]+):\s+(.*)`)
	var rePortForw = regexp.MustCompile(`(.*):(.*):\[(.*)\]:(\d*):\[(.*)\]:(\d*)`)
	r := strings.NewReader(out)
	s := bufio.NewScanner(r)
	var natnws []NatNetwork
	var nw NatNetwork
	for s.Scan() {
		line := s.Text()
		if strings.TrimSpace(line) == "" {
			natnws = append(natnws, nw)
			nw = NatNetwork{}
		}
		res := reColonLine.FindStringSubmatch(strings.TrimSpace(line))
		portforw6 := false
		if res == nil {
			if strings.TrimSpace(line) == "Port-forwarding (ipv4)" {
				var rules []PortForwarding
				if s.Scan() {
					line = s.Text()
				}
				ret := rePortForw.FindStringSubmatch(strings.TrimSpace(line))
				for ret != nil {
					rule, err := fillRule(ret)
					if err != nil {
						return nil, err
					}
					rules = append(rules, rule)
					if s.Scan() {
						line = s.Text()
					}
					ret = rePortForw.FindStringSubmatch(strings.TrimSpace(line))
				}
				nw.PortForward4 = rules
				if strings.TrimSpace(line) == "Port-forwarding (ipv6)" {
					portforw6 = true
				}
			}
			if (strings.TrimSpace(line) == "Port-forwarding (ipv6)") || (portforw6) {
				var rules []PortForwarding
				if strings.TrimSpace(line) == "Port-forwarding (ipv6)" {
					if s.Scan() {
						line = s.Text()
					}
				}
				portforw6 = false
				ret := rePortForw.FindStringSubmatch(strings.TrimSpace(line))
				for ret != nil {
					rule, err := fillRule(ret)
					if err != nil {
						return nil, err
					}
					rules = append(rules, rule)
					if s.Scan() {
						line = s.Text()
					}
					ret = rePortForw.FindStringSubmatch(strings.TrimSpace(line))
				}
				nw.PortForward6 = rules
			}
		} else {
			key, val := res[1], res[2]
			parseKeyVal(key, val, &nw)
		}
	}
	return natnws, nil
}

func (vb *VBox) ModifyNatNet(nat *NatNetwork, parameters []string) error {
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
		default:
			return errors.New("invalid parameter in the arguments")
		}
	}
	_, err := vb.manage(args...)
	return err
}
