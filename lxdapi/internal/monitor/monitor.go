package monitor

import (
	"context"
	"fmt"
	"lxdapi/internal/core"
	"lxdapi/internal/db"
	"lxdapi/internal/ipv6"
	"lxdapi/internal/lxc"
	"lxdapi/models"
	"lxdapi/pkg/logger"
	"lxdapi/pkg/plugin"
	"strings"
	"sync/atomic"
	"time"

	"github.com/tidwall/gjson"
)

type Monitor struct {
	lxcClient *lxc.Client
	interval  time.Duration
	neighbor  atomic.Value
}

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

	for {
		rows := m.loadRows()
		m.autoResetTraffic()

		for i := range rows {
			select {
			case <-ctx.Done():
				logger.Info("流量监控已停止")
				return
			default:
			}

			m.process(&rows[i])
		}

		if !m.wait(ctx, m.interval) {
			logger.Info("流量监控已停止")
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
	if err := db.DB.Select("name", "status", "traffic_limit", "traffic_usage", "rx_bytes", "tx_bytes", "locked").
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
		case "running":
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
		updates["locked"] = true
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
		if !strings.HasPrefix(ip, cfg.Prefix) || seen[ip] {
			continue
		}
		seen[ip] = true

		if err := ipv6.NserIPv6(cfg.Iface, ip, cfg.Gateway); err != nil {
			logger.Warn("容器 %s IPv6邻居请求发送失败 %s -> %s: %v", containerName, ip, cfg.Gateway, err)
			continue
		}
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

type containerUsageSnapshot struct {
	Status    string
	MemUsage  float64
	DiskUsage float64
	RxBytes   uint64
	TxBytes   uint64
	IPv6s     []string
}

// getContainerState 使用 gjson 流式按需提取 LXD state 响应字段，避免完整反序列化开销。
func (m *Monitor) getContainerState(ctx context.Context, name string) (*containerUsageSnapshot, error) {
	data, err := m.lxcClient.GetRaw(ctx, fmt.Sprintf("/1.0/instances/%s/state", name))
	if err != nil {
		return nil, err
	}

	stateBytes := gjson.GetBytes(data, "metadata")

	snap := &containerUsageSnapshot{
		Status: stateBytes.Get("status").String(),
	}

	snap.MemUsage = stateBytes.Get("memory.usage").Float()

	if root := stateBytes.Get("disk.root.usage"); root.Exists() {
		snap.DiskUsage = root.Float()
	}

	// 仅采集 eth0 网卡
	if eth0 := stateBytes.Get("network.eth0"); eth0.Exists() {
		counters := eth0.Get("counters")
		snap.RxBytes = counters.Get("bytes_received").Uint()
		snap.TxBytes = counters.Get("bytes_sent").Uint()

		eth0.Get("addresses").ForEach(func(_, v gjson.Result) bool {
			if v.Get("family").String() == "inet6" {
				addr := v.Get("address").String()
				if addr != "" && !strings.HasPrefix(addr, "fe80:") {
					snap.IPv6s = append(snap.IPv6s, addr)
				}
			}
			return true
		})
	}

	return snap, nil
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
	Locked        bool      `json:"locked"`
	LastReset     time.Time `json:"last_reset"`
}

func (m *Monitor) GetTraffic(containerName string) (*TrafficInfo, error) {
	var container models.Container
	if err := db.DB.Where("name = ?", containerName).
		Select("name", "traffic_usage", "traffic_limit", "rx_bytes", "tx_bytes", "locked", "last_reset").
		First(&container).Error; err != nil {
		return nil, err
	}
	return &TrafficInfo{
		ContainerName: container.Name,
		RxBytes:       container.RxBytes,
		TxBytes:       container.TxBytes,
		TrafficUsage:  container.TrafficUsage,
		TrafficLimit:  container.TrafficLimit,
		Locked:        container.Locked,
		LastReset:     container.LastReset,
	}, nil
}

func (m *Monitor) autoResetTraffic() {
	now := time.Now()
	firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var containers []models.Container
	// 具备补偿机制：记录到本月首日之前未重置或从未重置的都视为待重置
	if err := db.DB.Select("name", "status", "locked", "last_reset").
		Where("last_reset < ? OR last_reset IS NULL", firstDayOfMonth).
		Find(&containers).Error; err != nil {
		logger.Error("查询待重置流量记录失败: %v", err)
		return
	}

	for _, container := range containers {
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
		}

		if err := db.DB.Model(&models.Container{}).Where("name = ?", container.Name).Updates(updates).Error; err != nil {
			logger.Error("重置容器流量失败 %s: %v", container.Name, err)
			continue
		}

		logger.OK("容器流量已按月自动重置: %s", container.Name)
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
	if err := db.DB.Select("traffic_limit").Where("name = ?", containerName).First(&container).Error; err != nil {
		return fmt.Errorf("容器不存在: %v", err)
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