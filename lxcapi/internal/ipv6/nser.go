package ipv6

import (
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/vishvananda/netlink"
	"golang.org/x/net/ipv6"

	"lxcapi/pkg/logger"
)

// nser - IPv6 Neighbor Solicitation tool
// Sends an IPv6 Neighbor Solicitation (NS) packet so that the gateway learns the
// neighbor for the given source IPv6 and forwards traffic to it.
//
// The sendNS implementation is based on github.com/kkqy-go/nser (BSD-3-Clause).
// Original: https://github.com/kkqy-go/nser

// AnnounceNeighbor 对单个容器 IPv6 地址执行邻居学习：
// 始终先添加内核代理NDP条目（等价 ip -6 neigh add proxy），
// 随后发送邻居请求。代理条目添加失败仅记录告警，不影响邻居请求发送。
func AnnounceNeighbor(iface, addr, gateway string) error {
	if err := AddProxyNDP(iface, addr); err != nil {
		logger.Warn("添加IPv6代理NDP条目失败 %s dev %s: %v", addr, iface, err)
	}
	return NserIPv6(iface, addr, gateway)
}

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

// proxy NDP 条目管理：等价于
//
//	ip -6 neigh add proxy <addr> dev <iface>
//	ip -6 neigh del proxy <addr> dev <iface>
//	ip -6 neigh show proxy dev <iface>
//
// 需要 root / CAP_NET_ADMIN 权限。所有操作均为 netlink 调用，不 spawn 子进程。

// AddProxyNDP 为指定网卡添加（或替换）一条 IPv6 代理邻居条目。幂等。
func AddProxyNDP(iface, addr string) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return fmt.Errorf("网卡 %s 不存在: %w", iface, err)
	}

	ip := net.ParseIP(strings.TrimSpace(addr))
	if ip == nil || ip.To16() == nil {
		return fmt.Errorf("无效的IPv6地址: %s", addr)
	}

	neigh := &netlink.Neigh{
		LinkIndex: link.Attrs().Index,
		Family:    netlink.FAMILY_V6,
		IP:        ip,
		Flags:     netlink.NTF_PROXY,
	}
	if err := netlink.NeighSet(neigh); err != nil {
		return fmt.Errorf("添加代理NDP条目失败 %s dev %s: %w", ip, iface, err)
	}
	return nil
}

// DelProxyNDP 删除指定网卡的 IPv6 代理邻居条目。
func DelProxyNDP(iface, addr string) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return fmt.Errorf("网卡 %s 不存在: %w", iface, err)
	}

	ip := net.ParseIP(strings.TrimSpace(addr))
	if ip == nil || ip.To16() == nil {
		return fmt.Errorf("无效的IPv6地址: %s", addr)
	}

	neigh := &netlink.Neigh{
		LinkIndex: link.Attrs().Index,
		Family:    netlink.FAMILY_V6,
		IP:        ip,
		Flags:     netlink.NTF_PROXY,
	}
	if err := netlink.NeighDel(neigh); err != nil {
		return fmt.Errorf("删除代理NDP条目失败 %s dev %s: %w", ip, iface, err)
	}
	return nil
}

// ListProxyNDP 返回指定网卡上现有的全部 IPv6 代理邻居地址。
func ListProxyNDP(iface string) ([]net.IP, error) {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return nil, fmt.Errorf("网卡 %s 不存在: %w", iface, err)
	}

	neighs, err := netlink.NeighProxyList(link.Attrs().Index, netlink.FAMILY_V6)
	if err != nil {
		return nil, fmt.Errorf("查询代理NDP条目失败 dev %s: %w", iface, err)
	}

	ips := make([]net.IP, 0, len(neighs))
	for _, n := range neighs {
		if n.IP == nil || n.IP.To4() != nil {
			continue
		}
		ips = append(ips, n.IP)
	}
	return ips, nil
}

// FlushProxyNDP 删除指定网卡上所有匹配给定前缀的 IPv6 代理邻居条目。
// 前缀外的条目一律保留（保护外部手动添加的条目）。
func FlushProxyNDP(iface, prefix string) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return fmt.Errorf("网卡 %s 不存在: %w", iface, err)
	}

	neighs, err := netlink.NeighProxyList(link.Attrs().Index, netlink.FAMILY_V6)
	if err != nil {
		return fmt.Errorf("查询代理NDP条目失败 dev %s: %w", iface, err)
	}

	var firstErr error
	for _, n := range neighs {
		if n.IP == nil || n.IP.To4() != nil {
			continue
		}
		if !InPrefix(n.IP.String(), prefix) {
			continue
		}
		del := &netlink.Neigh{
			LinkIndex: link.Attrs().Index,
			Family:    netlink.FAMILY_V6,
			IP:        n.IP,
			Flags:     netlink.NTF_PROXY,
		}
		if err := netlink.NeighDel(del); err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("删除代理NDP条目失败 %s dev %s: %w", n.IP, iface, err)
			}
		}
	}
	return firstErr
}

// CleanupProxyNDPForAddress 按当前配置清理单个地址的代理NDP条目：
func CleanupProxyNDPForAddress(addr string) {
	cfg, err := GetNeighborConfig()
	if err != nil || !cfg.Enabled || cfg.Iface == "" || cfg.Prefix == "" {
		return
	}
	if !InPrefix(addr, cfg.Prefix) {
		return
	}
	if err := DelProxyNDP(cfg.Iface, addr); err != nil {
		logger.Warn("清理IPv6代理NDP条目失败 %s dev %s: %v", addr, cfg.Iface, err)
	}
}

// EnsureProxyNDPForAddress 按当前配置为单个地址添加代理NDP条目（幂等）。
func EnsureProxyNDPForAddress(addr string) {
	cfg, err := GetNeighborConfig()
	if err != nil || !cfg.Enabled || cfg.Iface == "" || cfg.Prefix == "" {
		return
	}
	if !InPrefix(addr, cfg.Prefix) {
		return
	}
	if err := AddProxyNDP(cfg.Iface, addr); err != nil {
		logger.Warn("添加IPv6代理NDP条目失败 %s dev %s: %v", addr, cfg.Iface, err)
	}
}

// InPrefix 判断 IPv6 地址是否落在给定前缀内。
func InPrefix(addr, prefix string) bool {
	a, ok := parseNetipAddr(addr)
	if !ok {
		return false
	}

	p := strings.TrimSpace(prefix)
	if p == "" {
		return false
	}
	if i := strings.IndexByte(p, '%'); i >= 0 {
		p = p[:i]
	}

	if strings.Contains(p, "/") {
		pr, err := netip.ParsePrefix(p)
		if err != nil {
			return false
		}
		return pr.Masked().Contains(a)
	}

	a16 := a.As16()
	var ah [8]uint16
	for i := 0; i < 8; i++ {
		ah[i] = uint16(a16[i*2])<<8 | uint16(a16[i*2+1])
	}

	if strings.Contains(p, "::") {
		parts := strings.SplitN(p, "::", 2)
		left, ok1 := parseHextets(parts[0])
		right, ok2 := parseHextets(parts[1])
		if !ok1 || !ok2 {
			return false
		}
		if len(left)+len(right) > 8 {
			return false
		}
		for i, h := range left {
			if ah[i] != h {
				return false
			}
		}
		for i, h := range right {
			if ah[8-len(right)+i] != h {
				return false
			}
		}
		return true
	}

	toks, ok := parseHextets(p)
	if !ok || len(toks) == 0 || len(toks) > 8 {
		return false
	}
	for i, h := range toks {
		if ah[i] != h {
			return false
		}
	}
	return true
}

func parseHextets(s string) ([]uint16, bool) {
	s2 := strings.Trim(s, ":")
	if s2 == "" {
		return []uint16{}, true
	}
	parts := strings.Split(s2, ":")
	out := make([]uint16, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			return nil, false
		}
		v, err := strconv.ParseUint(p, 16, 16)
		if err != nil {
			return nil, false
		}
		out = append(out, uint16(v))
	}
	return out, true
}

func parseNetipAddr(s string) (netip.Addr, bool) {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '%'); i >= 0 {
		s = s[:i]
	}
	a, err := netip.ParseAddr(s)
	if err != nil {
		return netip.Addr{}, false
	}
	return a.WithZone(""), true
}
