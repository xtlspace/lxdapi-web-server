package ipv6

import (
	"fmt"
	"net"
	"net/netip"
	"strings"

	"github.com/vishvananda/netlink"

	"lxdapi/pkg/logger"
)

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
// 仅当主开关与代理NDP开关均开启、且地址匹配前缀时执行，否则不做任何操作。
func CleanupProxyNDPForAddress(addr string) {
	cfg, err := GetNeighborConfig()
	if err != nil || !cfg.Enabled || !cfg.ProxyNDP || cfg.Iface == "" || cfg.Prefix == "" {
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
	if err != nil || !cfg.Enabled || !cfg.ProxyNDP || cfg.Iface == "" || cfg.Prefix == "" {
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
// prefix 支持两种写法：
//   - 标准 CIDR，如 "2001:db8:1::/64"；
//   - 省略长度的地址前缀，如 "2001:b030:a81d:a00::"（按已写出的 hextet 数推断长度）。
func InPrefix(addr, prefix string) bool {
	a, ok := parseNetipAddr(addr)
	if !ok {
		return false
	}

	p := strings.TrimSpace(prefix)
	if p == "" {
		return false
	}

	if pr, err := netip.ParsePrefix(p); err == nil {
		return pr.Masked().Contains(a)
	}

	base := strings.TrimSuffix(p, "::")
	base = strings.TrimSuffix(base, ":")
	if base == "" {
		return false
	}

	bits := 0
	for _, h := range strings.Split(base, ":") {
		if h == "" {
			continue
		}
		bits += 16
	}
	if bits > 128 {
		bits = 128
	}

	baseAddr, ok := parseNetipAddr(p)
	if !ok {
		return false
	}
	pr := netip.PrefixFrom(baseAddr, bits)
	if !pr.IsValid() {
		return false
	}
	return pr.Masked().Contains(a)
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
