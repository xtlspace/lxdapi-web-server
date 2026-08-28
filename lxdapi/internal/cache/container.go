package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"lxdapi/internal/db"
	"lxdapi/internal/lxc"
	"lxdapi/models"
	"lxdapi/pkg/format"
	"lxdapi/pkg/logger"
)

type ContainerCacheJSON struct {
	Name            string                 `json:"name"`
	Status          string                 `json:"status"`
	Image           string                 `json:"image"`
	CPU             int                    `json:"cpu"`
	Memory          string                 `json:"memory"`
	Disk            string                 `json:"disk"`
	CPUUsage        float64                `json:"cpu_usage"`
	MemoryUsage     string                 `json:"memory_usage"`
	MemoryUsageRaw  uint64                 `json:"memory_usage_raw"`
	DiskUsage       string                 `json:"disk_usage"`
	DiskUsageRaw    uint64                 `json:"disk_usage_raw"`
	TrafficUsage    string                 `json:"traffic_usage"`
	TrafficUsageRaw uint64                 `json:"traffic_usage_raw"`
	TrafficLimit    int                    `json:"traffic_limit"`
	Config          map[string]interface{} `json:"config"`
	CreatedAt       *time.Time             `json:"created_at"`
	LastSync        *time.Time             `json:"last_sync"`
	Remark          string                 `json:"remark"`
}

var (
	containerCache   = make(map[string]*ContainerCacheJSON)
	containerCacheMu sync.RWMutex
)

func GetAllContainersCache() []ContainerCacheJSON {
	containerCacheMu.RLock()
	result := make([]ContainerCacheJSON, 0, len(containerCache))
	for _, v := range containerCache {
		result = append(result, *v)
	}
	containerCacheMu.RUnlock()
	if len(result) > 0 {
		return result
	}

	RefreshAllCache()
	containerCacheMu.RLock()
	result = make([]ContainerCacheJSON, 0, len(containerCache))
	for _, v := range containerCache {
		result = append(result, *v)
	}
	containerCacheMu.RUnlock()
	return result
}

func GetContainerCache(name string) (*ContainerCacheJSON, bool) {
	containerCacheMu.RLock()
	v, ok := containerCache[name]
	containerCacheMu.RUnlock()
	if ok {
		return v, true
	}

	var container models.Container
	if err := db.DB.Where("name = ?", name).First(&container).Error; err != nil {
		return nil, false
	}

	cacheData := buildCacheEntry(container, GetLatestCPUUsage(container.Name))

	containerCacheMu.Lock()
	containerCache[name] = cacheData
	containerCacheMu.Unlock()

	return cacheData, true
}

func DeleteContainerCache(name string) {
	containerCacheMu.Lock()
	delete(containerCache, name)
	containerCacheMu.Unlock()
}

func RefreshAllCache() {
	var containers []models.Container
	db.DB.Find(&containers)

	cpuUsages := getLatestCPUUsages()

	newCache := make(map[string]*ContainerCacheJSON, len(containers))
	for _, c := range containers {
		entry := buildCacheEntry(c, cpuUsages[c.Name])
		newCache[c.Name] = entry
	}

	containerCacheMu.Lock()
	containerCache = newCache
	containerCacheMu.Unlock()
}

func RefreshContainerCache(ctx context.Context, name string) error {
	lxcClient := lxc.DefaultClient()

	var container models.Container
	if err := db.DB.Where("name = ?", name).First(&container).Error; err != nil {
		return fmt.Errorf("容器不存在于数据库: %s", name)
	}

	if container.Status == "frozen" {
		logger.Info("容器 %s 处于暂停状态，跳过缓存刷新", name)
		return nil
	}

	info, err := lxcClient.GetContainerInfo(ctx, name)
	if err != nil {
		return err
	}

	container.Status = strings.ToLower(info.Status)
	now := time.Now()
	container.LastSync = &now
	if t, err := time.Parse(time.RFC3339, info.Created); err == nil {
		container.CreatedAtLXD = &t
	}

	if image, ok := info.Config["image.description"].(string); ok {
		container.Image = image
	}

	if cpuLimit, ok := info.Config["limits.cpu"].(string); ok {
		container.CPU = parseCPULimit(cpuLimit)
	}

	if memLimit, ok := info.Config["limits.memory"].(string); ok {
		container.Memory = parseMemoryStringToMB(memLimit)
	}

	for _, devConfig := range info.Devices {
		if devMap, ok := devConfig.(map[string]interface{}); ok {
			if devType, ok := devMap["type"].(string); ok && devType == "disk" {
				if path, ok := devMap["path"].(string); ok && path == "/" {
					if size, ok := devMap["size"].(string); ok {
						container.Disk = parseMemoryStringToMB(size)
					}
				}
			}
		}
	}

	var traffic models.Traffic
	if err := db.DB.Where("container_name = ?", name).First(&traffic).Error; err == nil {
		container.TrafficUsageRaw = uint64(traffic.TotalGB)
		container.TrafficUsage = fmt.Sprintf("%.2f", traffic.TotalGB)
	}

	if info.State != nil && info.Status == "Running" {
		if memUsage, ok := info.State.Memory["usage"].(float64); ok {
			container.MemoryUsageRaw = uint64(memUsage)
			container.MemoryUsage = format.BytesUint64(uint64(memUsage))
		}

		if diskUsage, ok := info.State.Disk["root"].(map[string]interface{}); ok {
			if usage, ok := diskUsage["usage"].(float64); ok {
				container.DiskUsageRaw = uint64(usage)
				container.DiskUsage = format.BytesUint64(uint64(usage))
			}
		}
	}

	if info.Config != nil {
		configJSON, _ := json.Marshal(info.Config)
		container.ConfigJSON = string(configJSON)
	}

	if err := db.DB.Save(&container).Error; err != nil {
		return fmt.Errorf("更新容器缓存失败: %v", err)
	}

	entry := buildCacheEntry(container, GetLatestCPUUsage(container.Name))
	containerCacheMu.Lock()
	containerCache[name] = entry
	containerCacheMu.Unlock()

	logger.Info("刷新容器 %s 缓存成功", name)
	return nil
}

func buildCacheEntry(c models.Container, cpuUsage float64) *ContainerCacheJSON {
	cacheData := &ContainerCacheJSON{
		Name:            c.Name,
		Status:          c.Status,
		Image:           c.Image,
		CPU:             c.CPU,
		Memory:          formatMBToString(c.Memory),
		Disk:            formatMBToString(c.Disk),
		CPUUsage:        cpuUsage,
		MemoryUsage:     c.MemoryUsage,
		MemoryUsageRaw:  c.MemoryUsageRaw,
		DiskUsage:       c.DiskUsage,
		DiskUsageRaw:    c.DiskUsageRaw,
		TrafficUsage:    c.TrafficUsage,
		TrafficUsageRaw: c.TrafficUsageRaw,
		TrafficLimit:    c.TrafficLimit,
		CreatedAt:       c.CreatedAtLXD,
		LastSync:        c.LastSync,
		Remark:          c.Remark,
	}

	if c.ConfigJSON != "" {
		json.Unmarshal([]byte(c.ConfigJSON), &cacheData.Config)
	}

	return cacheData
}

func parseCPULimit(limit string) int {
	var cpus int
	fmt.Sscanf(limit, "%d", &cpus)
	if cpus == 0 {
		cpus = 1
	}
	return cpus
}

func GetLatestCPUUsage(name string) float64 {
	var metric models.CPUMetric
	if err := db.DB.Where("name = ?", name).Order("created_at DESC").First(&metric).Error; err != nil {
		return 0
	}
	return metric.CPUUsage
}

func getLatestCPUUsages() map[string]float64 {
	var metrics []models.CPUMetric
	db.DB.Raw(`SELECT cm.* FROM cpu_metrics cm
		JOIN (SELECT name, MAX(id) AS id FROM cpu_metrics GROUP BY name) t ON cm.id = t.id`).Scan(&metrics)

	result := make(map[string]float64, len(metrics))
	for _, v := range metrics {
		result[v.Name] = v.CPUUsage
	}
	return result
}

func GetRecentCPUMetrics(name string, since time.Time) []models.CPUMetric {
	var metrics []models.CPUMetric
	db.DB.Where("name = ? AND created_at >= ?", name, since).Order("created_at ASC").Find(&metrics)
	return metrics
}

func parseMemoryStringToMB(memStr string) int {
	memStr = strings.ToUpper(strings.TrimSpace(memStr))
	if memStr == "" {
		return 0
	}

	var value float64
	var unit string
	fmt.Sscanf(memStr, "%f%s", &value, &unit)

	switch unit {
	case "KB":
		return int(value / 1024)
	case "MB":
		return int(value)
	case "GB":
		return int(value * 1024)
	case "TB":
		return int(value * 1024 * 1024)
	default:
		return int(value / 1024 / 1024)
	}
}

func formatMBToString(mb int) string {
	if mb == 0 {
		return ""
	}
	if mb < 1024 {
		return fmt.Sprintf("%dMB", mb)
	}
	return fmt.Sprintf("%.1fGB", float64(mb)/1024)
}
