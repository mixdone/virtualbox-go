package virtualbox

import (
	"fmt"
)

func (vb *VBox) RemoveDHCPServer(netName string) error {
	_, err := vb.manage("dhcpserver", "remove", "--netname", netName)
	return err
}

func (vb *VBox) AddDHCPServer(dhcp DHCPServer) (string, error) {
	args := []string{"dhcpserver", "add", "--netname", dhcp.NetworkName}
	args = append(args, fmt.Sprintf("--ip=%s", dhcp.IPAddress), fmt.Sprintf("--netmask=%s", dhcp.NetworkMask),
		fmt.Sprintf("--lowerip=%s", dhcp.LowerIPAddress), fmt.Sprintf("--upperip=%s", dhcp.UpperIPAddress))

	if dhcp.Enabled {
		args = append(args, "--enable")
	} else {
		args = append(args, "--disable")
	}
	return vb.manage(args...)
}

func (vb *VBox) ModifyDHCPServer(dhcp DHCPServer, parametrs []string) error {
	if len(parametrs) == 0 {
		return nil
	}

	args := []string{"dhcpserver", "modify"}

	for _, s := range parametrs {
		switch s {
		case "netname":
			args = append(args, "--netname", dhcp.NetworkName)
		case "netmask":
			args = append(args, "--netmask", dhcp.NetworkMask)
		case "ip":
			args = append(args, "--ip", dhcp.IPAddress)
		case "lowerip":
			args = append(args, "--lowerip", dhcp.LowerIPAddress)
		case "upperip":
			args = append(args, "--upperip", dhcp.UpperIPAddress)
		case "work":
			tmp := "--disable"
			if dhcp.Enabled {
				tmp = "--enable"
			}
			args = append(args, tmp)
		}
	}

	_, err := vb.manage(args...)
	return err
}

func (vb *VBox) StartDHCPServer(netName string) error {
	_, err := vb.manage("dhcpserver", "start", "--netname", netName)
	return err
}

func (vb *VBox) RestartDHCPServer(netName string) error {
	_, err := vb.manage("dhcpserver", "restart", "--netname", netName)
	return err
}

func (vb *VBox) StopDHCPServer(netName string) error {
	_, err := vb.manage("dhcpserver", "stop", "--netname", netName)
	return err
}

func (vb *VBox) DHCPInfo(netName string) (*DHCPServer, string, [][2]string, error) {
	out, err := vb.manage("list", "dhcpservers")
	if err != nil {
		return nil, out, nil, err
	}

	optionList := make([]([2]string), 0, 20)
	_ = parseKeyValues(out, reColonLine, func(key, val string) error {
		optionList = append(optionList, [2]string{key, val})
		return nil
	})

	dhcp := &DHCPServer{}

	for i := 0; i < len(optionList); i++ {
		fmt.Println(optionList[i][0], ":", optionList[i][1])
		if optionList[i][0] == "NetworkName" && optionList[i][1] == netName {

			dhcp.NetworkName = (optionList[i][1])
			dhcp.IPAddress = (optionList[i+1][1])
			dhcp.LowerIPAddress = (optionList[i+2][1])
			dhcp.UpperIPAddress = (optionList[i+3][1])
			dhcp.NetworkMask = (optionList[i+4][1])

			if (optionList[i+5][1]) == "Yes" {
				dhcp.Enabled = true
			} else {
				dhcp.Enabled = false
			}
		}
	}
	return dhcp, out, nil, nil
}
