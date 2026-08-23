package traffic

import (
	"context"
	"encoding/json"
	"fmt"
	"lxdapi/internal/core"
	"lxdapi/internal/db"
	"lxdapi/internal/lxc"
	"lxdapi/models"
	"lxdapi/pkg/logger"
	"lxdapi/pkg/plugin"
	"os/exec"
	"sort"
	"strings"
	"time"
)

type Monitor struct {
	lxcClient *lxc.Client
	interval  time.Duration
	pending   map[string]*sampleState
	listAt    time.Time
}

const (
	listRefreshInterval = 5 * time.Second
	maxIdleWait         = time.Second
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
	if !cfg.Enabled {
		logger.Info("流量监控未启用")
		return nil
	}

	GlobalMonitor = &Monitor{
		lxcClient: lxc.NewClient(),
		interval:  time.Duration(cfg.Interval) * time.Second,
		pending:   make(map[string]*sampleState),
	}

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

		if now.Sub(m.listAt) >= listRefreshInterval {
			m.refreshSchedule(now)
			m.listAt = now
		}

		if periodicAt.IsZero() || now.Sub(periodicAt) >= m.interval {
			m.checkAutoReset()
			m.checkUserTrafficLimits()
			periodicAt = now
		}

		due, next := m.nextDue(now)
		if due != nil {
			m.sampleContainer(due)
			continue
		}

		wait := maxIdleWait
		if !next.IsZero() {
			if d := next.Sub(now); d < wait {
				wait = d
			}
		}
		if d := listRefreshInterval - now.Sub(m.listAt); d < wait {
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
	return due, next
}

func (m *Monitor) refreshSchedule(now time.Time) {
	var containers []models.Container
	if err := db.DB.Select("name", "status").Where("status = ?", "running").Find(&containers).Error; err != nil {
		logger.Error("获取容器列表失败: %v", err)
		return
	}

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
	if m.isUserTrafficLocked(s.name) {
		m.stopLockedContainer(s.name)
		delete(m.pending, s.name)
		return
	}

	if m.isTrafficLocked(s.name) {
		m.stopLockedContainer(s.name)
		delete(m.pending, s.name)
		return
	}

	state, err := m.getContainerState(s.name)
	if err != nil || state.Status != "Running" {
		delete(m.pending, s.name)
		return
	}

	now := time.Now()

	if s.hasBase {
		m.collectContainerUsage(s.name, state, s.baseCPU, s.baseAt)
	}

	s.baseCPU = state.CPUUsage
	s.baseAt = now
	s.hasBase = true
	s.nextDue = now.Add(m.interval)
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
}

func (m *Monitor) getContainerState(name string) (*containerUsageSnapshot, error) {
	cmd := exec.Command("lxc", "query", fmt.Sprintf("/1.0/instances/%s/state", name))
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var raw struct {
		Status  string                 `json:"status"`
		CPU     map[string]interface{} `json:"cpu"`
		Memory  map[string]interface{} `json:"memory"`
		Disk    map[string]interface{} `json:"disk"`
		Network map[string]struct {
			Counters NetworkCounters `json:"counters"`
		} `json:"network"`
	}
	if err := json.Unmarshal(output, &raw); err != nil {
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
	if eth0, ok := raw.Network["eth0"]; ok {
		snap.RxBytes = eth0.Counters.BytesReceived
		snap.TxBytes = eth0.Counters.BytesSent
	}

	return snap, nil
}

func (m *Monitor) collectContainerUsage(name string, snap *containerUsageSnapshot, baseCPU float64, sampledAt time.Time) {
	m.updateTraffic(name, &NetworkCounters{
		BytesReceived: snap.RxBytes,
		BytesSent:     snap.TxBytes,
	})

	var container models.Container
	if err := db.DB.Where("name = ?", name).First(&container).Error; err != nil {
		return
	}

	wallSeconds := time.Since(sampledAt).Seconds()
	cpuPercent := 0.0
	if wallSeconds > 0 {
		cpuPercent = (snap.CPUUsage - baseCPU) / 1e9 / wallSeconds * 100
	}
	if cpuPercent < 0 {
		cpuPercent = 0
	}
	if cpuPercent > 100 {
		cpuPercent = 100
	}

	db.DB.Create(&models.CPUMetric{
		Name:     name,
		CPUUsage: cpuPercent,
	})

	container.MemoryUsageRaw = uint64(snap.MemUsage)
	container.MemoryUsage = formatBytes(uint64(snap.MemUsage))
	container.DiskUsageRaw = uint64(snap.DiskUsage)
	container.DiskUsage = formatBytes(uint64(snap.DiskUsage))

	var traffic models.Traffic
	if err := db.DB.Where("container_name = ?", name).First(&traffic).Error; err == nil {
		container.TrafficUsageRaw = uint64(traffic.TotalGB)
		container.TrafficUsage = fmt.Sprintf("%.2f", traffic.TotalGB)
	}

	container.LastSync = time.Now().Format(time.RFC3339)

	db.DB.Save(&container)
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
		
		traffic.RxBytes = int64(newRx)
		traffic.TxBytes = int64(newTx)
		traffic.TotalGB = totalGB
		traffic.LastUpdate = time.Now()
		
		db.DB.Save(&traffic)
		
		if traffic.LimitGB > 0 && totalGB >= float64(traffic.LimitGB) {
			logger.Warn("容器 %s 流量超限: %.2fGB / %dGB", containerName, totalGB, traffic.LimitGB)
			m.handleOverLimit(containerName, totalGB, traffic.LimitGB)
		}
	}
}

func (m *Monitor) handleOverLimit(containerName string, current float64, limit int) {
	db.DB.Model(&models.Traffic{}).Where("container_name = ?", containerName).Update("locked", true)
	
	ctx := context.Background()
	
	if err := m.lxcClient.StopContainer(ctx, containerName); err != nil {
		logger.Error("自动停止超限容器失败 %s: %v", containerName, err)
		return
	}
	
	db.DB.Model(&models.Container{}).Where("name = ?", containerName).Update("status", "frozen")
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
	if err := db.DB.Where("container_name = ?", containerName).First(&traffic).Error; err != nil {
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
	if err := db.DB.Where("name = ?", containerName).First(&container).Error; err != nil {
		return
	}
	
	if container.Status == "stopped" {
		return
	}
	
	ctx := context.Background()
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



func (m *Monitor) checkAutoReset() {
	now := time.Now()
	if now.Day() != 1 {
		return
	}
	
	var traffics []models.Traffic
	firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	if err := db.DB.Where("last_reset < ? OR last_reset IS NULL", firstDayOfMonth).Find(&traffics).Error; err != nil {
		return
	}
	
	for _, traffic := range traffics {
		m.autoResetTraffic(traffic.ContainerName)
	}
}

func (m *Monitor) autoResetTraffic(containerName string) {
	state, err := m.getContainerState(containerName)
	if err != nil {
		return
	}
	
	var traffic models.Traffic
	if err := db.DB.Where("container_name = ?", containerName).First(&traffic).Error; err != nil {
		return
	}
	
	traffic.RxBytes = int64(state.RxBytes)
	traffic.TxBytes = int64(state.TxBytes)
	traffic.TotalGB = 0
	traffic.Locked = false
	traffic.LastUpdate = time.Now()
	traffic.LastReset = time.Now()
	
	db.DB.Save(&traffic)
	
	db.DB.Model(&models.Container{}).Where("name = ?", containerName).Update("status", "stopped")
	
	logger.OK("流量已自动重置并解锁: %s", containerName)
}

func (m *Monitor) ResetTraffic(containerName string) error {
	state, err := m.getContainerState(containerName)
	if err != nil {
		return fmt.Errorf("获取容器流量数据失败")
	}
	
	var container models.Container
	if err := db.DB.Where("name = ?", containerName).First(&container).Error; err != nil {
		return fmt.Errorf("容器不存在: %v", err)
	}
	
	if err := db.DB.Unscoped().Where("container_name = ?", containerName).Delete(&models.Traffic{}).Error; err != nil {
		return fmt.Errorf("重置流量统计失败: %v", err)
	}
	
	resetDay := container.CreatedAt.Day()
	
	traffic := models.Traffic{
		ContainerName: containerName,
		RxBytes:       int64(state.RxBytes),
		TxBytes:       int64(state.TxBytes),
		TotalGB:       0,
		LimitGB:       container.TrafficLimit,
		ResetDay:      1, // 固定每月1号
		Locked:        false,
		LastUpdate:    time.Now(),
		LastReset:     time.Now(),
	}
	
	if err := db.DB.Create(&traffic).Error; err != nil {
		return fmt.Errorf("创建重置基准记录失败: %v", err)
	}

	ctx := context.Background()
	if err := m.lxcClient.StartContainer(ctx, containerName); err != nil {
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

	logger.OK("流量统计已重置: %s (重置日期: 每月%d号)", containerName, resetDay)
	return nil
}

func (m *Monitor) checkUserTrafficLimits() {
	var users []models.User
	if err := db.DB.Where("traffic_limit > 0 AND traffic_locked = ?", false).Find(&users).Error; err != nil {
		return
	}
	
	for _, user := range users {
		var containers []models.Container
		if err := db.DB.Where("user_id = ?", user.Username).Find(&containers).Error; err != nil {
			continue
		}
		
		totalUsage := user.TrafficUsed
		for _, c := range containers {
			var traffic models.Traffic
			if err := db.DB.Where("container_name = ?", c.Name).First(&traffic).Error; err == nil {
				totalUsage += traffic.TotalGB
			}
		}
		
		if totalUsage >= float64(user.TrafficLimit) {
			m.handleUserOverLimit(&user, containers, totalUsage)
		}
	}
}

func (m *Monitor) handleUserOverLimit(user *models.User, containers []models.Container, currentUsage float64) {
	db.DB.Model(user).Update("traffic_locked", true)
	
	ctx := context.Background()
	for _, c := range containers {
		if c.Status == "stopped" || c.Status == "frozen" {
			continue
		}
		if err := m.lxcClient.StopContainer(ctx, c.Name); err != nil {
			logger.Error("停止用户超限容器失败 %s: %v", c.Name, err)
			continue
		}
		db.DB.Model(&models.Container{}).Where("name = ?", c.Name).Update("status", "frozen")
	}
	
	logger.Warn("用户 %s 流量超限: %.2fGB / %dGB，已停止所有容器", user.Username, currentUsage, user.TrafficLimit)
	
	if mgr := plugin.GetManager(); mgr != nil {
		mgr.GetHookManager().TriggerAsync(plugin.HookTrafficOverLimit, ctx, map[string]interface{}{
			"type":     "user",
			"username": user.Username,
			"current":  uint64(currentUsage * 1024 * 1024 * 1024),
			"limit":    uint64(user.TrafficLimit) * 1024 * 1024 * 1024,
		})
	}
}

func (m *Monitor) isUserTrafficLocked(containerName string) bool {
	var container models.Container
	if err := db.DB.Where("name = ?", containerName).First(&container).Error; err != nil {
		return false
	}
	
	var user models.User
	if err := db.DB.Where("username = ?", container.UserID).First(&user).Error; err != nil {
		return false
	}
	
	if user.TrafficLocked {
		if user.TrafficLimit == 0 {
			db.DB.Model(&user).Update("traffic_locked", false)
			logger.OK("用户 %s 无流量限制，自动解锁", user.Username)
			return false
		}
		
		var containers []models.Container
		db.DB.Where("user_id = ?", user.Username).Find(&containers)
		
		totalUsage := user.TrafficUsed
		for _, c := range containers {
			var traffic models.Traffic
			if err := db.DB.Where("container_name = ?", c.Name).First(&traffic).Error; err == nil {
				totalUsage += traffic.TotalGB
			}
		}
		
		if totalUsage < float64(user.TrafficLimit) {
			db.DB.Model(&user).Update("traffic_locked", false)
			logger.OK("用户 %s 流量已恢复正常，自动解锁 (使用: %.2fGB / 限制: %dGB)", user.Username, totalUsage, user.TrafficLimit)
			return false
		}
	}
	
	return user.TrafficLocked
}

func (m *Monitor) ResetUserTraffic(username string) error {
	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return fmt.Errorf("用户不存在: %v", err)
	}
	
	var containers []models.Container
	db.DB.Where("user_id = ?", username).Find(&containers)
	
	for _, c := range containers {
		m.ResetTraffic(c.Name)
	}
	
	db.DB.Model(&user).Update("traffic_locked", false)
	logger.OK("用户流量已重置: %s", username)
	return nil
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
