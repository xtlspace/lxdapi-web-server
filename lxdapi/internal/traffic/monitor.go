package traffic

import (
	"context"
	"fmt"
	"lxdapi/internal/core"
	"lxdapi/internal/db"
	"lxdapi/internal/ipv6"
	"lxdapi/internal/lxc"
	"lxdapi/models"
	"lxdapi/pkg/format"
	"lxdapi/pkg/logger"
	"lxdapi/pkg/plugin"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

type Monitor struct {
	mu        sync.RWMutex
	lxcClient *lxc.Client
	interval  time.Duration
	pending   map[string]*sampleState
	listAt    time.Time
	neighbor  atomic.Value
	sem       chan struct{} // 限制最大并发采样协程数
}

const (
	maxIdleWait = time.Second
)

type sampleState struct {
	name    string
	nextDue time.Time
	baseCPU float64
	baseAt  time.Time
	hasBase bool
}

var GlobalMonitor *Monitor

func InitMonitor() error {
	cfg := core.GlobalConfig.Traffic

	GlobalMonitor = &Monitor{
		lxcClient: lxc.DefaultClient(),
		interval:  time.Duration(cfg.Interval) * time.Second,
		pending:   make(map[string]*sampleState),
		sem:       make(chan struct{}, 16), // 最多同时运行 16 个采样任务
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

	var periodicAt time.Time
	for {
		select {
		case <-ctx.Done():
			logger.Info("流量监控已停止")
			return
		default:
		}

		now := time.Now()

		if now.Sub(m.listAt) >= m.interval {
			m.refreshSchedule(now)
			m.listAt = now
		}

		if periodicAt.IsZero() || now.Sub(periodicAt) >= m.interval {
			m.autoResetTraffic()
			periodicAt = now
		}

		due, next := m.nextDue(now)
		if due != nil {
			select {
			case m.sem <- struct{}{}:
				go func(s *sampleState) {
					defer func() { <-m.sem }()
					m.sampleContainer(s)
				}(due)
			default:
			}
			continue
		}

		wait := maxIdleWait
		if !next.IsZero() {
			if d := next.Sub(now); d < wait {
				wait = d
			}
		}
		if d := m.interval - now.Sub(m.listAt); d < wait {
			wait = d
		}
		if !periodicAt.IsZero() {
			if d := periodicAt.Add(m.interval).Sub(now); d < wait {
				wait = d
			}
		}
		if wait < 0 {
			wait = 0
		}

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			logger.Info("流量监控已停止")
			return
		case <-timer.C:
		}
	}
}

func (m *Monitor) nextDue(now time.Time) (*sampleState, time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var due *sampleState
	next := time.Time{}
	for _, s := range m.pending {
		if !s.nextDue.After(now) {
			if due == nil || s.nextDue.Before(due.nextDue) {
				due = s
			}
		} else if next.IsZero() || s.nextDue.Before(next) {
			next = s.nextDue
		}
	}
	if due != nil {
		due.nextDue = now.Add(m.interval)
	}
	return due, next
}

func (m *Monitor) refreshSchedule(now time.Time) {
	var containers []models.Container
	if err := db.DB.Select("name", "status").Find(&containers).Error; err != nil {
		logger.Error("获取容器列表失败: %v", err)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	active := make(map[string]bool, len(containers))
	added := make([]string, 0)
	for _, c := range containers {
		active[c.Name] = true
		if _, ok := m.pending[c.Name]; !ok {
			added = append(added, c.Name)
		}
	}

	for name := range m.pending {
		if !active[name] {
			delete(m.pending, name)
		}
	}

	if len(added) == 0 {
		return
	}

	sort.Strings(added)

	total := len(m.pending) + len(added)
	spacing := m.interval / time.Duration(total)
	if spacing <= 0 {
		spacing = time.Millisecond
	}

	slot := now
	for _, name := range added {
		m.pending[name] = &sampleState{name: name, nextDue: slot}
		slot = slot.Add(spacing)
	}
}

func (m *Monitor) sampleContainer(s *sampleState) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	state, err := m.getContainerState(ctx, s.name)
	if err != nil {
		return
	}

	newStatus := strings.ToLower(state.Status)
	db.DB.Model(&models.Container{}).
		Where("name = ? AND status <> ? AND status <> ?", s.name, newStatus, "frozen").
		Update("status", newStatus)

	if state.Status == "Running" {
		var cont models.Container
		if err := db.DB.Where("name = ?", s.name).First(&cont).Error; err == nil && cont.Status == "frozen" {
			logger.Warn("检测到数据库frozen但实际running，强制停止容器: %s", s.name)
			stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := m.lxcClient.StopContainer(stopCtx, s.name)
			cancel()
			if err != nil {
				logger.Warn("强制停止容器失败 %s: %v", s.name, err)
			}
			s.hasBase = false
			return
		}
	}

	if state.Status != "Running" {
		s.hasBase = false
		return
	}

	m.handleIPv6NeighborRequests(s.name, state)

	if m.isTrafficLocked(s.name) {
		m.stopLockedContainer(s.name)
		s.hasBase = false
		return
	}

	if s.hasBase {
		m.collectContainerUsage(s.name, state, s.baseCPU, s.baseAt)
	}

	s.baseCPU = state.CPUUsage
	s.baseAt = time.Now()
	s.hasBase = true
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

type NetworkCounters struct {
	BytesReceived   uint64 `json:"bytes_received"`
	BytesSent       uint64 `json:"bytes_sent"`
	PacketsReceived uint64 `json:"packets_received"`
	PacketsSent     uint64 `json:"packets_sent"`
}

type containerUsageSnapshot struct {
	Status    string
	CPUUsage  float64
	MemUsage  float64
	DiskUsage float64
	RxBytes   uint64
	TxBytes   uint64
	IPv6s     []string
}

func (m *Monitor) getContainerState(ctx context.Context, name string) (*containerUsageSnapshot, error) {
	var raw struct {
		Status  string                 `json:"status"`
		CPU     map[string]interface{} `json:"cpu"`
		Memory  map[string]interface{} `json:"memory"`
		Disk    map[string]interface{} `json:"disk"`
		Network map[string]struct {
			Counters  NetworkCounters `json:"counters"`
			Addresses []struct {
				Family  string `json:"family"`
				Address string `json:"address"`
			} `json:"addresses"`
		} `json:"network"`
	}
	if err := m.lxcClient.Get(ctx, fmt.Sprintf("/1.0/instances/%s/state", name), &raw); err != nil {
		return nil, err
	}

	snap := &containerUsageSnapshot{Status: raw.Status}

	if v, ok := raw.CPU["usage"].(float64); ok {
		snap.CPUUsage = v
	}
	if v, ok := raw.Memory["usage"].(float64); ok {
		snap.MemUsage = v
	}
	if diskRoot, ok := raw.Disk["root"].(map[string]interface{}); ok {
		if v, ok := diskRoot["usage"].(float64); ok {
			snap.DiskUsage = v
		}
	}

	// 仅采集 eth0 网卡
	if eth0, ok := raw.Network["eth0"]; ok {
		snap.RxBytes = eth0.Counters.BytesReceived
		snap.TxBytes = eth0.Counters.BytesSent

		for _, addr := range eth0.Addresses {
			if addr.Family == "inet6" && addr.Address != "" && !strings.HasPrefix(addr.Address, "fe80:") {
				snap.IPv6s = append(snap.IPv6s, addr.Address)
			}
		}
	}

	return snap, nil
}

func (m *Monitor) collectContainerUsage(name string, snap *containerUsageSnapshot, baseCPU float64, sampledAt time.Time) {
	m.updateTraffic(name, &NetworkCounters{
		BytesReceived: snap.RxBytes,
		BytesSent:     snap.TxBytes,
	})

	wallSeconds := time.Since(sampledAt).Seconds()
	cpuPercent := 0.0
	if wallSeconds > 0 {
		cpuPercent = (snap.CPUUsage - baseCPU) / 1e9 / wallSeconds * 100
	}
	if cpuPercent < 0 {
		cpuPercent = 0
	}

	go func(cName string, cPercent float64) {
		_ = db.DB.Create(&models.CPUMetric{
			Name:     cName,
			CPUUsage: cPercent,
		}).Error
	}(name, cpuPercent)

	var trafficUsageRaw uint64
	var trafficUsage string
	var traffic models.Traffic
	if err := db.DB.Select("total_gb").Where("container_name = ?", name).First(&traffic).Error; err == nil {
		trafficUsageRaw = uint64(traffic.TotalGB)
		trafficUsage = fmt.Sprintf("%.2f", traffic.TotalGB)
	}

	db.DB.Model(&models.Container{}).Where("name = ?", name).Updates(map[string]interface{}{
		"memory_usage_raw":  uint64(snap.MemUsage),
		"memory_usage":      format.BytesUint64(uint64(snap.MemUsage)),
		"disk_usage_raw":    uint64(snap.DiskUsage),
		"disk_usage":        format.BytesUint64(uint64(snap.DiskUsage)),
		"traffic_usage_raw": trafficUsageRaw,
		"traffic_usage":     trafficUsage,
		"last_sync":         time.Now(),
	})
}

func (m *Monitor) updateTraffic(containerName string, stats *NetworkCounters) {
	var traffic models.Traffic
	err := db.DB.Where("container_name = ?", containerName).First(&traffic).Error

	if err != nil {
		traffic = models.Traffic{
			ContainerName: containerName,
			RxBytes:       int64(stats.BytesReceived),
			TxBytes:       int64(stats.BytesSent),
			TotalGB:       float64(stats.BytesReceived+stats.BytesSent) / (1024 * 1024 * 1024),
			LastUpdate:    time.Now(),
		}
		db.DB.Create(&traffic)
		return
	}

	newRx := stats.BytesReceived
	newTx := stats.BytesSent
	var deltaRx, deltaTx uint64

	if newRx >= uint64(traffic.RxBytes) {
		deltaRx = newRx - uint64(traffic.RxBytes)
	} else {
		deltaRx = newRx
	}

	if newTx >= uint64(traffic.TxBytes) {
		deltaTx = newTx - uint64(traffic.TxBytes)
	} else {
		deltaTx = newTx
	}

	if deltaRx > 0 || deltaTx > 0 {
		incrementGB := float64(deltaRx+deltaTx) / (1024 * 1024 * 1024)
		totalGB := traffic.TotalGB + incrementGB

		db.DB.Model(&traffic).Updates(map[string]interface{}{
			"rx_bytes":    int64(newRx),
			"tx_bytes":    int64(newTx),
			"total_gb":    totalGB,
			"last_update": time.Now(),
		})

		if traffic.LimitGB > 0 && totalGB >= float64(traffic.LimitGB) {
			logger.Warn("容器 %s 流量超限: %.2fGB / %dGB", containerName, totalGB, traffic.LimitGB)
			m.handleOverLimit(containerName, totalGB, traffic.LimitGB)
		}
	}
}

func (m *Monitor) handleOverLimit(containerName string, current float64, limit int) {
	db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Traffic{}).Where("container_name = ?", containerName).Update("locked", true).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Container{}).Where("name = ?", containerName).Update("status", "frozen").Error; err != nil {
			return err
		}
		return nil
	})

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

func (m *Monitor) isTrafficLocked(containerName string) bool {
	var traffic models.Traffic
	if err := db.DB.Select("locked", "limit_gb", "total_gb").Where("container_name = ?", containerName).First(&traffic).Error; err != nil {
		return false
	}

	if traffic.Locked {
		if traffic.LimitGB == 0 || traffic.TotalGB < float64(traffic.LimitGB) {
			db.DB.Model(&models.Traffic{}).Where("container_name = ?", containerName).Update("locked", false)
			logger.OK("容器 %s 流量已恢复正常，自动解锁 (使用: %.2fGB / 限制: %dGB)", containerName, traffic.TotalGB, traffic.LimitGB)
			return false
		}
	}
	return traffic.Locked
}

func (m *Monitor) stopLockedContainer(containerName string) {
	var container models.Container
	if err := db.DB.Select("status").Where("name = ?", containerName).First(&container).Error; err != nil {
		return
	}

	if container.Status == "stopped" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.lxcClient.StopContainer(ctx, containerName); err == nil {
		db.DB.Model(&models.Container{}).Where("name = ?", containerName).Update("status", "frozen")
		logger.Warn("流量锁定容器尝试运行已被自动停止: %s", containerName)
	}
}

func (m *Monitor) GetTraffic(containerName string) (*models.Traffic, error) {
	var traffic models.Traffic
	err := db.DB.Where("container_name = ?", containerName).First(&traffic).Error
	if err != nil {
		return nil, err
	}
	return &traffic, nil
}

func (m *Monitor) autoResetTraffic() {
	now := time.Now()
	firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var traffics []models.Traffic
	if err := db.DB.Where("last_reset < ? OR last_reset IS NULL", firstDayOfMonth).Find(&traffics).Error; err != nil {
		logger.Error("查询待重置流量记录失败: %v", err)
		return
	}

	for _, traffic := range traffics {
		var container models.Container
		if err := db.DB.Select("status").Where("name = ?", traffic.ContainerName).First(&container).Error; err != nil {
			continue
		}

		if container.Status == "frozen" && !traffic.Locked {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		state, err := m.getContainerState(ctx, traffic.ContainerName)
		cancel()

		if err != nil {
			logger.Error("重置容器流量时获取网卡状态失败 %s: %v", traffic.ContainerName, err)
			continue
		}

		db.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&traffic).Updates(map[string]interface{}{
				"rx_bytes":    int64(state.RxBytes),
				"tx_bytes":    int64(state.TxBytes),
				"total_gb":    0,
				"locked":      false,
				"last_update": now,
				"last_reset":  now,
			}).Error; err != nil {
				return err
			}

			if container.Status == "frozen" {
				tx.Model(&models.Container{}).Where("name = ?", traffic.ContainerName).Update("status", "stopped")
			}
			return nil
		})

		logger.OK("容器流量已按月自动重置并解除锁定: %s", traffic.ContainerName)
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

	err = db.DB.Model(&models.Traffic{}).Where("container_name = ?", containerName).Updates(map[string]interface{}{
		"rx_bytes":    int64(state.RxBytes),
		"tx_bytes":    int64(state.TxBytes),
		"total_gb":    0,
		"limit_gb":    container.TrafficLimit,
		"reset_day":   1,
		"locked":      false,
		"last_update": time.Now(),
		"last_reset":  time.Now(),
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