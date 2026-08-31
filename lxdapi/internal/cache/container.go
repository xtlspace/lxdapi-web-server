package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"lxdapi/internal/db"
	"lxdapi/internal/lxc"
	"lxdapi/models"
	"lxdapi/pkg/logger"
)

type ContainerCacheJSON struct {
	Name            string                 `json:"name"`
	Status          string                 `json:"status"`
	Image           string                 `json:"image"`
	CPU             int                    `json:"cpu"`
	Memory          string                 `json:"memory"`
	Disk            string                 `json:"disk"`
	MemoryUsageRaw  uint64                 `json:"memory_usage_raw"`
	DiskUsageRaw    uint64                 `json:"disk_usage_raw"`
	TrafficUsage    float64                `json:"traffic_usage"`
	TrafficLimit    int                    `json:"traffic_limit"`
	RxBytes         int64                  `json:"rx_bytes"`
	TxBytes         int64                  `json:"tx_bytes"`
	Locked          bool                   `json:"locked"`
	Config          map[string]interface{} `json:"config"`
	CreatedAt       *time.Time             `json:"created_at"`
	LastSync        *time.Time             `json:"last_sync"`
	Remark          string                 `json:"remark"`
}

// GetAllContainersFromDB 直接从数据库读取全部容器并映射为缓存结构，保证实时一致
func GetAllContainersFromDB() []ContainerCacheJSON {
	var containers []models.Container
	db.DB.Find(&containers)

	result := make([]ContainerCacheJSON, 0, len(containers))
	for i := range containers {
		result = append(result, *buildCacheEntry(containers[i]))
	}
	return result
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

	if info.State != nil && info.Status == "Running" {
		if memUsage, ok := info.State.Memory["usage"].(float64); ok {
			container.MemoryUsageRaw = uint64(memUsage)
		}

		if diskUsage, ok := info.State.Disk["root"].(map[string]interface{}); ok {
			if usage, ok := diskUsage["usage"].(float64); ok {
				container.DiskUsageRaw = uint64(usage)
			}
		}
	}

	if info.Config != nil {
		configJSON, _ := json.Marshal(info.Config)
		container.ConfigJSON = string(configJSON)
	}

	if err := db.DB.Save(&container).Error; err != nil {
		return fmt.Errorf("更新容器状态失败: %v", err)
	}

	logger.Info("刷新容器 %s 状态成功", name)
	return nil
}

func buildCacheEntry(c models.Container) *ContainerCacheJSON {
	cacheData := &ContainerCacheJSON{
		Name:           c.Name,
		Status:         c.Status,
		Image:          c.Image,
		CPU:            c.CPU,
		Memory:         formatMBToString(c.Memory),
		Disk:           formatMBToString(c.Disk),
		MemoryUsageRaw: c.MemoryUsageRaw,
		DiskUsageRaw:   c.DiskUsageRaw,
		TrafficUsage:   c.TrafficUsage,
		TrafficLimit:   c.TrafficLimit,
		RxBytes:        c.RxBytes,
		TxBytes:        c.TxBytes,
		Locked:         c.Locked,
		CreatedAt:      c.CreatedAtLXD,
		LastSync:       c.LastSync,
		Remark:         c.Remark,
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
