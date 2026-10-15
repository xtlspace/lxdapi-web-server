package ipv6

import (
	"fmt"
	"net"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv6"

	"lxdapi/pkg/logger"
)

// nser - IPv6 Neighbor Solicitation tool
// Sends an IPv6 Neighbor Solicitation (NS) packet so that the gateway learns the
// neighbor for the given source IPv6 and forwards traffic to it.
//
// The sendNS implementation is based on github.com/kkqy-go/nser (BSD-3-Clause).
// Original: https://github.com/kkqy-go/nser
//
// sendNS requires root/administrator privileges to use raw sockets (ip6:58).

// NserIPv6 sends an IPv6 Neighbor Solicitation from srcIP toward dstIP (the gateway)
// on the given network interface. It corresponds to the nser CLI invocation:
//
//	nser -iface <iface> -src <srcIP> -dst <dstIP>
func NserIPv6(iface string, srcIP, dstIP string) error {
	sourceIP := net.ParseIP(strings.TrimSpace(srcIP))
	if sourceIP == nil {
		return fmt.Errorf("无效的源IPv6地址: %s", srcIP)
	}
	targetIP := net.ParseIP(strings.TrimSpace(dstIP))
	if targetIP == nil {
		return fmt.Errorf("无效的目标IPv6地址: %s", dstIP)
	}
	return sendNS(iface, sourceIP, targetIP)
}

// AnnounceNeighbor 对单个容器 IPv6 地址执行邻居学习：
// proxyNDP 为真时先添加内核代理NDP条目（等价 ip -6 neigh add proxy），
// 随后发送邻居请求。代理条目添加失败仅记录告警，不影响邻居请求发送。
func AnnounceNeighbor(iface, addr, gateway string, proxyNDP bool) error {
	if proxyNDP {
		if err := AddProxyNDP(iface, addr); err != nil {
			logger.Warn("添加IPv6代理NDP条目失败 %s dev %s: %v", addr, iface, err)
		}
	}
	return NserIPv6(iface, addr, gateway)
}

// sendNS builds and sends a Neighbor Solicitation packet using native Go raw sockets.
func sendNS(ifaceName string, sourceIP, targetIP net.IP) error {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return fmt.Errorf("failed to get interface %s: %w", ifaceName, err)
	}

	conn, err := net.ListenPacket("ip6:58", "::")
	if err != nil {
		return fmt.Errorf("failed to listen for icmpv6 packets: %w. Check for root privileges", err)
	}
	defer conn.Close()

	rawConn := ipv6.NewPacketConn(conn)

	if err := rawConn.SetMulticastHopLimit(255); err != nil {
		return fmt.Errorf("failed to set multicast hop limit: %w", err)
	}

	solicitedNodeAddr := net.IP{
		0xff, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x01, 0xff, targetIP[13], targetIP[14], targetIP[15],
	}

	icmpv6Layer := &layers.ICMPv6{TypeCode: layers.CreateICMPv6TypeCode(layers.ICMPv6TypeNeighborSolicitation, 0)}
	ipv6LayerForChecksum := &layers.IPv6{
		SrcIP:      sourceIP,
		DstIP:      solicitedNodeAddr,
		NextHeader: layers.IPProtocolICMPv6,
	}
	icmpv6Layer.SetNetworkLayerForChecksum(ipv6LayerForChecksum)

	nsLayer := &layers.ICMPv6NeighborSolicitation{
		TargetAddress: targetIP,
		Options: []layers.ICMPv6Option{
			{Type: layers.ICMPv6OptSourceAddress, Data: iface.HardwareAddr},
		},
	}

	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
	err = gopacket.SerializeLayers(buffer, opts, icmpv6Layer, nsLayer)
	if err != nil {
		return fmt.Errorf("failed to serialize icmpv6 ns layer: %w", err)
	}
	packetData := buffer.Bytes()

	wcm := ipv6.ControlMessage{
		Src:     sourceIP,
		IfIndex: iface.Index,
	}
	ipAddr := &net.IPAddr{IP: solicitedNodeAddr}

	if _, err = rawConn.WriteTo(packetData, &wcm, ipAddr); err != nil {
		return fmt.Errorf("failed to write packet: %w", err)
	}

	logger.OK("nser: 已发送IPv6邻居请求 %s -> 网关 %s (iface %s)", sourceIP, targetIP, ifaceName)
	return nil
}
