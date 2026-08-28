package ipv6

import (
	"fmt"
	"lxdapi/pkg/logger"
	"os/exec"
	"strings"
)

func (m *Manager) addPortDNAT(publicIP string, publicPort, publicPortEnd int, containerIP string, containerPort, containerPortEnd int, protocol, iface string) error {
	var portRange, containerPortRange string
	if publicPortEnd > 0 && publicPortEnd != publicPort {
		portRange = fmt.Sprintf("%d-%d", publicPort, publicPortEnd)
		containerPortRange = fmt.Sprintf("[%s]:%d-%d", containerIP, containerPort, containerPortEnd)
	} else {
		portRange = fmt.Sprintf("%d", publicPort)
		containerPortRange = fmt.Sprintf("[%s]:%d", containerIP, containerPort)
	}

	protos := []string{}
	if protocol == "both" {
		protos = []string{"tcp", "udp"}
	} else {
		protos = []string{protocol}
	}

	for _, proto := range protos {
		args := []string{"add", "rule", "inet", "lxdnat", "prerouting"}
		if publicIP != "" {
			args = append(args, "ip6", "daddr", publicIP)
		}
		if iface != "" {
			args = append(args, "iifname", iface)
		}
		args = append(args, proto, "dport", portRange, "dnat", "to", containerPortRange)

		cmd := exec.Command("nft", args...)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("nft添加IPv6 DNAT规则失败: %v, output: %s", err, string(output))
		}
		logger.Info("nft添加IPv6端口转发: [%s]:%s(%s) -> %s", publicIP, portRange, proto, containerPortRange)
	}

	return nil
}

func (m *Manager) removePortDNAT(publicIP string, publicPort, publicPortEnd int, containerIP string, containerPort, containerPortEnd int, protocol, iface string) error {
	protos := []string{}
	if protocol == "both" {
		protos = []string{"tcp", "udp"}
	} else {
		protos = []string{protocol}
	}

	var deleted int
	for _, proto := range protos {
		protoDeleted := 0
		groups := m.buildPortDNATMatchGroups(publicIP, publicPort, publicPortEnd, "", 0, 0, proto, iface, false)
		for {
			handles := m.findNftHandlesByGroups("lxdnat", "prerouting", groups)
			if len(handles) == 0 {
				break
			}
			for _, handle := range handles {
				cmd := exec.Command("nft", "delete", "rule", "inet", "lxdnat", "prerouting", "handle", handle)
				if err := cmd.Run(); err != nil {
					logger.Error("nft删除IPv6端口转发失败: %v", err)
					return err
				}
				logger.Info("nft删除IPv6端口转发: %s %s handle %s", publicIP, proto, handle)
				deleted++
				protoDeleted++
			}
		}
		if protoDeleted == 0 && containerIP != "" {
			groups := m.buildPortDNATMatchGroups(publicIP, publicPort, publicPortEnd, containerIP, containerPort, containerPortEnd, proto, iface, true)
			handles := m.findNftHandlesByGroups("lxdnat", "prerouting", groups)
			for _, handle := range handles {
				cmd := exec.Command("nft", "delete", "rule", "inet", "lxdnat", "prerouting", "handle", handle)
				if err := cmd.Run(); err != nil {
					logger.Error("nft删除IPv6端口转发失败: %v", err)
					return err
				}
				logger.Info("nft删除IPv6端口转发: %s %s handle %s", publicIP, proto, handle)
				deleted++
				protoDeleted++
			}
		}
	}

	if deleted == 0 && len(protos) > 0 {
		return fmt.Errorf("nft删除IPv6端口转发失败: 未找到匹配规则")
	}
	return nil
}

func (m *Manager) buildPortDNATMatchGroups(publicIP string, publicPort, publicPortEnd int, containerIP string, containerPort, containerPortEnd int, protocol, iface string, includeTarget bool) [][]string {
	portRange := fmt.Sprintf("dport %d", publicPort)
	if publicPortEnd > 0 && publicPortEnd != publicPort {
		portRange = fmt.Sprintf("dport %d-%d", publicPort, publicPortEnd)
	}

	base := []string{}
	if publicIP != "" {
		base = append(base, publicIP)
	}
	if protocol != "" {
		base = append(base, protocol)
	}
	if publicPort > 0 {
		base = append(base, portRange)
	}

	if includeTarget && containerIP != "" {
		base = append(base, containerIP)
		if containerPort > 0 {
			if containerPortEnd > 0 && containerPortEnd != containerPort {
				base = append(base, fmt.Sprintf(":%d-%d", containerPort, containerPortEnd))
			} else {
				base = append(base, fmt.Sprintf(":%d", containerPort))
			}
		}
	}

	if iface == "" {
		return [][]string{base}
	}

	return [][]string{
		append(append([]string{}, base...), fmt.Sprintf("iifname %s", iface)),
		append(append([]string{}, base...), fmt.Sprintf("iifname \"%s\"", iface)),
	}
}

func (m *Manager) findNftHandlesByGroups(table, chain string, groups [][]string) []string {
	cmd := exec.Command("nft", "-a", "list", "chain", "inet", table, chain)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	seen := map[string]bool{}
	var handles []string
	for _, line := range lines {
		for _, matches := range groups {
			allMatch := true
			for _, match := range matches {
				if match != "" && !strings.Contains(line, match) {
					allMatch = false
					break
				}
			}
			if !allMatch {
				continue
			}
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "handle" && i+1 < len(parts) && !seen[parts[i+1]] {
					seen[parts[i+1]] = true
					handles = append(handles, parts[i+1])
				}
			}
		}
	}
	return handles
}
