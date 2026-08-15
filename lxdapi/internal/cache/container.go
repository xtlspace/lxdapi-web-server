package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
	CPUUsage        float64                `json:"cpu_usage"`
	MemoryUsage     string                 `json:"memory_usage"`
	MemoryUsageRaw  uint64                 `json:"memory_usage_raw"`
	DiskUsage       string                 `json:"disk_usage"`
	DiskUsageRaw    uint64                 `json:"disk_usage_raw"`
	TrafficUsage    string                 `json:"traffic_usage"`
	TrafficUsageRaw uint64                 `json:"traffic_usage_raw"`
	TrafficLimit    int                    `json:"traffic_limit"`
	Config          map[string]interface{} `json:"config"`
	CreatedAt       string                 `json:"created_at"`
	LastSync        string                 `json:"last_sync"`
	Remark          string                 `json:"remark"`
}

func GetAllContainersCache() []ContainerCacheJSON {
	var containers []models.Container
	db.DB.Find(&containers)

	result := make([]ContainerCacheJSON, 0, len(containers))
	for _, c := range containers {
		cacheData := ContainerCacheJSON{
			Name:            c.Name,
			Status:          c.Status,
			Image:           c.Image,
			CPU:             c.CPU,
			Memory:          formatMBToString(c.Memory),
			Disk:            formatMBToString(c.Disk),
			CPUUsage:        c.CPUUsage,
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

		result = append(result, cacheData)
	}

	return result
}

func GetContainerCache(name string) (*ContainerCacheJSON, bool) {
	var container models.Container
	if err := db.DB.Where("name = ?", name).First(&container).Error; err != nil {
		return nil, false
	}

	cacheData := &ContainerCacheJSON{
		Name:            container.Name,
		Status:          container.Status,
		Image:           container.Image,
		CPU:             container.CPU,
		Memory:          formatMBToString(container.Memory),
		Disk:            formatMBToString(container.Disk),
		CPUUsage:        container.CPUUsage,
		MemoryUsage:     container.MemoryUsage,
		MemoryUsageRaw:  container.MemoryUsageRaw,
		DiskUsage:       container.DiskUsage,
		DiskUsageRaw:    container.DiskUsageRaw,
		TrafficUsage:    container.TrafficUsage,
		TrafficUsageRaw: container.TrafficUsageRaw,
		TrafficLimit:    container.TrafficLimit,
		CreatedAt:       container.CreatedAtLXD,
		LastSync:        container.LastSync,
		Remark:          container.Remark,
	}

	if container.ConfigJSON != "" {
		json.Unmarshal([]byte(container.ConfigJSON), &cacheData.Config)
	}

	return cacheData, true
}

func DeleteContainerCache(name string) {
}

func RefreshAllContainersCache(ctx context.Context) (int, int, error) {
	lxcClient := lxc.NewClient()
	
	containerNames, err := lxcClient.ListAllContainers(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("获取容器列表失败: %v", err)
	}

	total := len(containerNames)
	success := 0
	failed := 0

	for _, name := range containerNames {
		if err := RefreshContainerCache(ctx, name); err != nil {
			logger.Error("刷新容器 %s 缓存失败: %v", name, err)
			failed++
		} else {
			success++
		}
	}

	return total, success, nil
}

func RefreshContainerCache(ctx context.Context, name string) error {
	lxcClient := lxc.NewClient()

	info, err := lxcClient.GetContainerInfo(ctx, name)
	if err != nil {
		return err
	}

	var container models.Container
	if err := db.DB.Where("name = ?", name).First(&container).Error; err != nil {
		return fmt.Errorf("容器不存在于数据库: %s", name)
	}

	container.Status = strings.ToLower(info.Status)
	container.LastSync = time.Now().Format(time.RFC3339)
	container.CreatedAtLXD = info.Created

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
		cpuUsage := getContainerCPUUsagePercent(ctx, name)
		container.CPUUsage = cpuUsage

		if memUsage, ok := info.State.Memory["usage"].(float64); ok {
			container.MemoryUsageRaw = uint64(memUsage)
			container.MemoryUsage = formatBytes(uint64(memUsage))
		}

		if diskUsage, ok := info.State.Disk["root"].(map[string]interface{}); ok {
			if usage, ok := diskUsage["usage"].(float64); ok {
				container.DiskUsageRaw = uint64(usage)
				container.DiskUsage = formatBytes(uint64(usage))
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

	logger.Info("刷新容器 %s 缓存成功", name)
	return nil
}

func parseCPULimit(limit string) int {
	var cpus int
	fmt.Sscanf(limit, "%d", &cpus)
	if cpus == 0 {
		cpus = 1
	}
	return cpus
}

func getContainerCPUUsagePercent(ctx context.Context, containerName string) float64 {
	lxcClient := lxc.NewClient()
	
	output, err := lxcClient.ExecInContainer(ctx, containerName, []string{"sh", "-c", "vmstat 1 2 | tail -1 | awk '{print $15}'"})
	if err != nil {
		logger.Warn("获取容器 %s CPU使用率失败: %v", containerName, err)
		return 0
	}
	
	idleStr := strings.TrimSpace(output)
	idlePercent, err := strconv.ParseFloat(idleStr, 64)
	if err != nil {
		logger.Warn("解析容器 %s CPU idle值失败: %v, 输出: %s", containerName, err, idleStr)
		return 0
	}
	
	cpuUsage := 100 - idlePercent
	if cpuUsage < 0 {
		cpuUsage = 0
	}
	if cpuUsage > 100 {
		cpuUsage = 100
	}
	
	return cpuUsage
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
