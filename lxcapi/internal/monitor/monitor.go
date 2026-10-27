package monitor

import (
	"context"
	"fmt"
	"lxcapi/internal/core"
	"lxcapi/internal/db"
	"lxcapi/internal/ipv6"
	"lxcapi/internal/lxc"
	"lxcapi/models"
	"lxcapi/pkg/logger"
	"lxcapi/pkg/plugin"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tidwall/gjson"
)

type Monitor struct {
	lxcClient *lxc.Client
	interval  time.Duration
	neighbor  atomic.Value

	lastResetMonth int
	lastResetDay   int

	proxyMu      sync.Mutex
	cycleDesired map[string]struct{}
	proxyMissing map[string]int
}

// proxyNDPStaleGrace 为一个代理NDP条目连续缺席多少个监控周期后才被清理，
// 用于规避单次采集抖动导致的短暂断网。
const proxyNDPStaleGrace = 2

var GlobalMonitor *Monitor

func InitMonitor() error {
	cfg := core.GlobalConfig.Traffic

	GlobalMonitor = &Monitor{
		lxcClient: lxc.DefaultClient(),
		interval:  time.Duration(cfg.Interval) * time.Second,
	}

	GlobalMonitor.reloadNeighborConfig()

	logger.OK("流量监控器初始化成功")
	return nil
}

func (m *Monitor) Start(ctx context.Context) {
	if m == nil {
		return
	}

	logger.Info("流量监控已启动，采样间隔: %v", m.interval)

	// 启动时全量补重置一次未重置流量
	m.autoResetTraffic()

	for {
		rows := m.loadRows()

		m.beginProxyCycle()

		for i := range rows {
			select {
			case <-ctx.Done():
				return
			default:
			}

			m.process(&rows[i])

			if i < len(rows)-1 {
				if !m.wait(ctx, m.interval) {
					return
				}
			}
		}

		// 本周期结束后对账：补齐缺失的代理NDP条目，清理连续缺席的残留条目
		m.reconcileProxyNDP()

		// 仅跨月时才重新查询并重置
		m.maybeAutoResetTraffic()

		if !m.wait(ctx, m.interval) {
			return
		}
	}
}

// wait 等待指定时长，期间可被 ctx 中断，返回 false 表示已取消。
func (m *Monitor) wait(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (m *Monitor) loadRows() []models.Container {
	var containers []models.Container
	if err := db.DB.Select("name", "status", "traffic_limit", "traffic_usage", "rx_bytes", "tx_bytes", "locked", "traffic_over_action", "traffic_over_limitspeed").
		Find(&containers).Error; err != nil {
		logger.Error("获取容器列表失败: %v", err)
		return nil
	}
	return containers
}

// process 按预定义三分支处理单个容器。
func (m *Monitor) process(c *models.Container) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	state, err := m.getContainerState(ctx, c.Name)
	if err != nil {
		return
	}

	switch state.Status {
	case "Running":
		switch c.Status {
		case "frozen":
			m.stopFrozenContainer(c.Name)
		case "running", "stopped", "limited":
			if c.Status == "stopped" {
				db.DB.Model(&models.Container{}).Where("name = ?", c.Name).Update("status", "running")
			}
			m.collectRunning(c, state)
		}
	case "Stopped":
		if c.Status == "running" {
			m.restartContainer(c.Name)
		}
	}
}

// stopFrozenContainer DB frozen 但实际 running，强制停止保持 frozen。
func (m *Monitor) stopFrozenContainer(name string) {
	logger.Warn("检测到数据库frozen但实际running，强制停止容器: %s", name)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.lxcClient.StopContainer(ctx, name); err != nil {
		logger.Warn("强制停止容器失败 %s: %v", name, err)
	}
}

// restartContainer DB running 但实际 stopped，重新启动。
func (m *Monitor) restartContainer(name string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.lxcClient.StartContainer(ctx, name); err != nil {
		logger.Warn("检测到容器意外停止后自动重启失败 %s: %v", name, err)
		return
	}
	db.DB.Model(&models.Container{}).Where("name = ?", name).Update("status", "running")
	logger.OK("检测到容器意外停止，已自动重启: %s", name)
}

// collectRunning DB running 且实际 running：计算流量增量写库，超限则冻结，否则采集内存硬盘。
func (m *Monitor) collectRunning(c *models.Container, state *containerUsageSnapshot) {
	newRx := state.RxBytes
	newTx := state.TxBytes

	deltaRx := trafficDelta(newRx, uint64(c.RxBytes))
	deltaTx := trafficDelta(newTx, uint64(c.TxBytes))
	incrementGB := float64(deltaRx+deltaTx) / (1024 * 1024 * 1024)

	totalGB := c.TrafficUsage + incrementGB

	updates := map[string]interface{}{
		"rx_bytes":      int64(newRx),
		"tx_bytes":      int64(newTx),
		"traffic_usage": totalGB,
	}

	if c.TrafficLimit > 0 && totalGB >= float64(c.TrafficLimit) {
		action := strings.ToLower(strings.TrimSpace(c.TrafficOverAction))
		if action == "" {
			action = "frozen"
		}

		updates["locked"] = true

		if action == "limited" {
			if c.Status != "limited" {
				updates["status"] = "limited"
				db.DB.Model(&models.Container{}).Where("name = ?", c.Name).Updates(updates)
				logger.Warn("容器 %s 流量超限: %.4fGB / %dGB，已触发限速 %dMbit", c.Name, totalGB, c.TrafficLimit, c.TrafficOverLimitSpeed)
				m.handleLimited(c)
			} else {
				db.DB.Model(&models.Container{}).Where("name = ?", c.Name).Updates(updates)
			}
			return
		}

		updates["status"] = "frozen"
		db.DB.Model(&models.Container{}).Where("name = ?", c.Name).Updates(updates)

		logger.Warn("容器 %s 流量超限: %.4fGB / %dGB", c.Name, totalGB, c.TrafficLimit)
		m.handleOverLimit(c.Name, totalGB, c.TrafficLimit)
		return
	}

	updates["memory_usage_raw"] = uint64(state.MemUsage)
	updates["disk_usage_raw"] = uint64(state.DiskUsage)
	updates["last_sync"] = time.Now()
	db.DB.Model(&models.Container{}).Where("name = ?", c.Name).Updates(updates)

	m.handleIPv6NeighborRequests(c.Name, state)
}

func trafficDelta(cur, base uint64) uint64 {
	if cur >= base {
		return cur - base
	}
	return cur
}

func (m *Monitor) handleIPv6NeighborRequests(containerName string, state *containerUsageSnapshot) {
	if len(state.IPv6s) == 0 {
		return
	}

	cfg := m.neighbor.Load().(models.IPv6NeighborConfig)
	if !cfg.Enabled || cfg.Iface == "" || cfg.Gateway == "" || cfg.Prefix == "" {
		return
	}

	seen := make(map[string]bool)
	for _, ip := range state.IPv6s {
		if !ipv6.InPrefix(ip, cfg.Prefix) || seen[ip] {
			continue
		}
		seen[ip] = true

		m.markDesired(ip)

		if err := ipv6.AnnounceNeighbor(cfg.Iface, ip, cfg.Gateway); err != nil {
			logger.Warn("容器 %s IPv6邻居请求发送失败 %s -> %s: %v", containerName, ip, cfg.Gateway, err)
			continue
		}
	}
}

func (m *Monitor) beginProxyCycle() {
	m.proxyMu.Lock()
	m.cycleDesired = make(map[string]struct{})
	m.proxyMu.Unlock()
}

func (m *Monitor) markDesired(addr string) {
	m.proxyMu.Lock()
	if m.cycleDesired != nil {
		m.cycleDesired[addr] = struct{}{}
	}
	m.proxyMu.Unlock()
}

// reconcileProxyNDP 在每个监控周期结束时对账内核代理NDP条目：
//   - 主开关或代理NDP开关关闭时，清理本前缀的全部条目；
//   - 否则补齐“应存在但内核缺失”的条目；
//   - 对“本前缀内、且连续缺席达到 proxyNDPStaleGrace 个周期”的条目执行清理，
//     前缀外的条目一律不触碰。
func (m *Monitor) reconcileProxyNDP() {
	cfg := m.neighbor.Load().(models.IPv6NeighborConfig)

	if !cfg.Enabled {
		m.resetProxyState()
		if cfg.Iface != "" && cfg.Prefix != "" {
			if err := ipv6.FlushProxyNDP(cfg.Iface, cfg.Prefix); err != nil {
				logger.Warn("清理IPv6代理NDP条目失败 dev %s: %v", cfg.Iface, err)
			}
		}
		return
	}

	if cfg.Iface == "" || cfg.Gateway == "" || cfg.Prefix == "" {
		return
	}

	m.proxyMu.Lock()
	desired := make(map[string]struct{}, len(m.cycleDesired))
	for k := range m.cycleDesired {
		desired[k] = struct{}{}
	}
	m.proxyMu.Unlock()

	cur, err := ipv6.ListProxyNDP(cfg.Iface)
	if err != nil {
		logger.Warn("查询IPv6代理NDP条目失败 dev %s: %v", cfg.Iface, err)
		return
	}

	current := make(map[string]struct{}, len(cur))
	for _, ip := range cur {
		current[ip.String()] = struct{}{}
	}

	m.proxyMu.Lock()
	if m.proxyMissing == nil {
		m.proxyMissing = make(map[string]int)
	}

	// 方向A：清理残留（仅限本前缀，连续缺席达到宽限才删）
	for key := range current {
		if !ipv6.InPrefix(key, cfg.Prefix) {
			continue
		}
		if _, ok := desired[key]; ok {
			delete(m.proxyMissing, key)
			continue
		}
		m.proxyMissing[key]++
		if m.proxyMissing[key] >= proxyNDPStaleGrace {
			if err := ipv6.DelProxyNDP(cfg.Iface, key); err != nil {
				logger.Warn("清理残留代理NDP条目失败 %s dev %s: %v", key, cfg.Iface, err)
			} else {
				logger.Info("已清理残留IPv6代理NDP条目 %s dev %s", key, cfg.Iface)
			}
			delete(m.proxyMissing, key)
		}
	}

	// 清理计数中已不在内核里的地址
	for key := range m.proxyMissing {
		if _, ok := current[key]; !ok {
			delete(m.proxyMissing, key)
		}
	}
	m.proxyMu.Unlock()

	// 方向B：补齐应存在但内核缺失的条目
	for key := range desired {
		if _, ok := current[key]; ok {
			continue
		}
		if err := ipv6.AddProxyNDP(cfg.Iface, key); err != nil {
			logger.Warn("补齐IPv6代理NDP条目失败 %s dev %s: %v", key, cfg.Iface, err)
		}
	}
}

func (m *Monitor) resetProxyState() {
	m.proxyMu.Lock()
	m.proxyMissing = nil
	m.proxyMu.Unlock()
}

// refreshProxyNDP 供配置保存后立即生效：查询当前运行中的容器并补发邻居请求、
// 添加代理NDP条目，无需等待下一个监控周期。
func (m *Monitor) refreshProxyNDP() {
	cfg := m.neighbor.Load().(models.IPv6NeighborConfig)
	if !cfg.Enabled || cfg.Iface == "" || cfg.Gateway == "" || cfg.Prefix == "" {
		return
	}

	for _, c := range m.loadRows() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		state, err := m.getContainerState(ctx, c.Name)
		cancel()
		if err != nil || state.Status != "Running" {
			continue
		}
		m.handleIPv6NeighborRequests(c.Name, state)
	}
}

func (m *Monitor) reloadNeighborConfig() {
	cfg, err := ipv6.GetNeighborConfig()
	if err != nil {
		cfg = models.IPv6NeighborConfig{}
	}
	m.neighbor.Store(cfg)
}

func ReloadNeighborConfig() {
	if GlobalMonitor != nil {
		GlobalMonitor.reloadNeighborConfig()
	}
}

// RefreshProxyNDP 在配置保存后立即对运行中容器生效（补发邻居请求、添加代理NDP条目）。
func RefreshProxyNDP() {
	if GlobalMonitor != nil {
		GlobalMonitor.refreshProxyNDP()
	}
}

type containerUsageSnapshot struct {
	Status    string
	MemUsage  float64
	DiskUsage float64
	RxBytes   uint64
	TxBytes   uint64
	IPv6s     []string
}

// getContainerState 使用 gjson 单次扫描（GetManyBytes）提取 LXD state 响应字段，
// 避免多次 Get 对 metadata 子串的重复词法解析。
func (m *Monitor) getContainerState(ctx context.Context, name string) (*containerUsageSnapshot, error) {
	data, err := m.lxcClient.GetRaw(ctx, fmt.Sprintf("/1.0/instances/%s/state", name))
	if err != nil {
		return nil, err
	}

	res := gjson.GetManyBytes(data,
		"metadata.status",
		"metadata.memory.usage",
		"metadata.disk.root.usage",
		"metadata.network.eth0.counters.bytes_received",
		"metadata.network.eth0.counters.bytes_sent",
		"metadata.network.eth0.addresses",
		"metadata.network.eth1.addresses",
	)

	snap := &containerUsageSnapshot{Status: res[0].String()}

	snap.MemUsage = res[1].Float()

	if res[2].Exists() {
		snap.DiskUsage = res[2].Float()
	}

	snap.RxBytes = res[3].Uint()
	snap.TxBytes = res[4].Uint()

	// 采集 eth0/eth1 网卡；addresses 存在才解析过滤 inet6
	for _, i := range []int{5, 6} {
		if addrs := res[i]; addrs.Exists() {
			addrs.ForEach(func(_, v gjson.Result) bool {
				if v.Get("family").String() == "inet6" {
					addr := v.Get("address").String()
					if addr != "" && !strings.HasPrefix(addr, "fe80:") {
						snap.IPv6s = append(snap.IPv6s, addr)
					}
				}
				return true
			})
		}
	}

	return snap, nil
}

func (m *Monitor) handleLimited(c *models.Container) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	speed := c.TrafficOverLimitSpeed
	if speed <= 0 {
		speed = 1
	}
	limit := fmt.Sprintf("%dMbit", speed)

	if err := m.lxcClient.UpdateContainerConfig(ctx, c.Name, 0, "", "", limit, limit, "", "", "", 0, nil, nil, nil); err != nil {
		logger.Error("容器限速失败 %s: %v", c.Name, err)
		return
	}

	logger.OK("容器流量超限已限速至 %dMbit: %s", speed, c.Name)

	if mgr := plugin.GetManager(); mgr != nil {
		mgr.GetHookManager().TriggerAsync(plugin.HookTrafficOverLimit, ctx, map[string]interface{}{
			"name":    c.Name,
			"current": uint64(c.TrafficUsage * 1024 * 1024 * 1024),
			"limit":   uint64(c.TrafficLimit) * 1024 * 1024 * 1024,
			"action":  "limited",
			"speed":   speed,
		})
	}
}

// restoreSpeed 恢复容器原始入站/出站带宽（流量重置或解除限速时调用）
func (m *Monitor) restoreSpeed(container *models.Container) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ingress := ""
	egress := ""
	if container.Ingress > 0 {
		ingress = fmt.Sprintf("%dMbit", container.Ingress)
	}
	if container.Egress > 0 {
		egress = fmt.Sprintf("%dMbit", container.Egress)
	}
	if ingress == "" && egress == "" {
		return
	}

	if err := m.lxcClient.UpdateContainerConfig(ctx, container.Name, 0, "", "", ingress, egress, "", "", "", 0, nil, nil, nil); err != nil {
		logger.Error("恢复容器带宽失败 %s: %v", container.Name, err)
		return
	}
	logger.OK("已恢复容器原始带宽: %s", container.Name)
}

func (m *Monitor) handleOverLimit(containerName string, current float64, limit int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.lxcClient.StopContainer(ctx, containerName); err != nil {
		logger.Error("自动停止超限容器失败 %s: %v", containerName, err)
		return
	}

	logger.OK("容器因流量超限已锁定并停止: %s", containerName)

	if mgr := plugin.GetManager(); mgr != nil {
		mgr.GetHookManager().TriggerAsync(plugin.HookTrafficOverLimit, ctx, map[string]interface{}{
			"name":    containerName,
			"current": uint64(current * 1024 * 1024 * 1024),
			"limit":   uint64(limit) * 1024 * 1024 * 1024,
		})
	}
}

type TrafficInfo struct {
	ContainerName string    `json:"container_name"`
	RxBytes       int64     `json:"rx_bytes"`
	TxBytes       int64     `json:"tx_bytes"`
	TrafficUsage  float64   `json:"traffic_usage"`
	TrafficLimit  int       `json:"traffic_limit"`
	TrafficResetCycle string `json:"traffic_reset_cycle"`
	TrafficOverAction string `json:"traffic_over_action"`
	TrafficOverLimitSpeed int `json:"traffic_over_limitspeed"`
	Locked        bool      `json:"locked"`
	LastReset     time.Time `json:"last_reset"`
}

func (m *Monitor) GetTraffic(containerName string) (*TrafficInfo, error) {
	var container models.Container
	if err := db.DB.Where("name = ?", containerName).
		Select("name", "traffic_usage", "traffic_limit", "rx_bytes", "tx_bytes", "locked", "last_reset",
			"traffic_reset_cycle", "traffic_over_action", "traffic_over_limitspeed").
		First(&container).Error; err != nil {
		return nil, err
	}
	return &TrafficInfo{
		ContainerName: container.Name,
		RxBytes:       container.RxBytes,
		TxBytes:       container.TxBytes,
		TrafficUsage:  container.TrafficUsage,
		TrafficLimit:  container.TrafficLimit,
		TrafficResetCycle: container.TrafficResetCycle,
		TrafficOverAction: container.TrafficOverAction,
		TrafficOverLimitSpeed: container.TrafficOverLimitSpeed,
		Locked:        container.Locked,
		LastReset:     container.LastReset,
	}, nil
}

// maybeAutoResetTraffic 仅在跨月时才真正执行重置查询，其余周期零开销快速返回。
func (m *Monitor) maybeAutoResetTraffic() {
	now := time.Now()
	curMonth := now.Year()*100 + int(now.Month())
	curDay := curMonth*100 + now.Day()
	if m.lastResetMonth == curMonth && m.lastResetDay == curDay {
		return
	}
	m.autoResetTraffic()
}

func (m *Monitor) autoResetTraffic() {
	now := time.Now()
	firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	m.lastResetMonth = now.Year()*100 + int(now.Month())
	m.lastResetDay = m.lastResetMonth*100 + now.Day()

	var containers []models.Container
	// 具备补偿机制：记录到本月首日之前未重置或从未重置的都视为待重置
	if err := db.DB.Select("name", "status", "locked", "last_reset", "traffic_reset_cycle", "traffic_over_action", "ingress", "egress").
		Find(&containers).Error; err != nil {
		logger.Error("查询待重置流量记录失败: %v", err)
		return
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for _, container := range containers {
		needsReset := false
		if container.TrafficResetCycle == "daily" {
			if container.LastReset.Before(today) || container.LastReset.IsZero() {
				needsReset = true
			}
		} else {
			if container.LastReset.Before(firstDayOfMonth) || container.LastReset.IsZero() {
				needsReset = true
			}
		}
		if !needsReset {
			continue
		}

		if container.Status == "frozen" && !container.Locked {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		state, err := m.getContainerState(ctx, container.Name)
		cancel()

		if err != nil {
			logger.Error("重置容器流量时获取网卡状态失败 %s: %v", container.Name, err)
			continue
		}

		updates := map[string]interface{}{
			"rx_bytes":      int64(state.RxBytes),
			"tx_bytes":      int64(state.TxBytes),
			"traffic_usage": 0.0,
			"locked":        false,
			"last_reset":    now,
		}
		if container.Status == "frozen" {
			updates["status"] = "stopped"
		} else if container.Status == "limited" {
			updates["status"] = "running"
			m.restoreSpeed(&container)
		}

		if err := db.DB.Model(&models.Container{}).Where("name = ?", container.Name).Updates(updates).Error; err != nil {
			logger.Error("重置容器流量失败 %s: %v", container.Name, err)
			continue
		}

		logger.OK("容器流量已自动重置: %s", container.Name)
	}
}

func (m *Monitor) ResetTraffic(containerName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	state, err := m.getContainerState(ctx, containerName)
	cancel()
	if err != nil {
		return fmt.Errorf("获取容器流量数据失败")
	}

	var container models.Container
	if err := db.DB.Select("name", "status", "traffic_limit", "traffic_reset_cycle", "ingress", "egress").
		Where("name = ?", containerName).First(&container).Error; err != nil {
		return fmt.Errorf("容器不存在: %v", err)
	}

	if container.Status == "limited" {
		m.restoreSpeed(&container)
	}

	err = db.DB.Model(&models.Container{}).Where("name = ?", containerName).Updates(map[string]interface{}{
		"rx_bytes":      int64(state.RxBytes),
		"tx_bytes":      int64(state.TxBytes),
		"traffic_usage": 0.0,
		"locked":        false,
		"last_reset":    time.Now(),
	}).Error

	if err != nil {
		return fmt.Errorf("重置流量统计失败: %v", err)
	}

	runCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.lxcClient.StartContainer(runCtx, containerName); err != nil {
		if strings.Contains(err.Error(), "already running") {
			db.DB.Model(&models.Container{}).Where("name = ?", containerName).Update("status", "running")
			logger.Info("容器已在运行中，跳过自动开机: %s", containerName)
		} else {
			db.DB.Model(&models.Container{}).Where("name = ?", containerName).Update("status", "stopped")
			logger.Error("重置流量后自动开机失败 %s: %v", containerName, err)
		}
	} else {
		db.DB.Model(&models.Container{}).Where("name = ?", containerName).Update("status", "running")
		logger.OK("重置流量后已自动开机: %s", containerName)
	}

	logger.OK("流量统计已重置: %s (重置日期: 每月1号)", containerName)
	return nil
}