package ipv6

import (
	"lxdapi/internal/db"
	"lxdapi/internal/network"
	"lxdapi/internal/nftutil"
	"lxdapi/models"
	"lxdapi/pkg/logger"
	"sync"
)

type Manager struct {
	mu sync.Mutex
}

var GlobalManager *Manager

func InitManager() error {
	GlobalManager = &Manager{}

	if err := nftutil.InitNftTable(); err != nil {
		logger.Warn("初始化nftables表失败: %v", err)
	}

	go GlobalManager.restorePortMappings()
	logger.OK("IPv6管理器初始化成功，已启动端口映射恢复")

	neighborInit()
	return nil
}

// neighborInit seeds the IPv6 neighbor request config table and, if enabled,
// applies the required kernel parameters at startup.
func neighborInit() {
	var cfg models.IPv6NeighborConfig
	if err := db.DB.First(&cfg).Error; err != nil {
		cfg = models.IPv6NeighborConfig{}
		db.DB.Create(&cfg)
		return
	}

	if cfg.Enabled {
		logger.OK("IPv6邻居请求已启用: iface=%s prefix=%s gateway=%s", cfg.Iface, cfg.Prefix, cfg.Gateway)
		ApplySysctl()
	}
}

// GetNeighborConfig returns the single-row IPv6 neighbor request configuration.
func GetNeighborConfig() (models.IPv6NeighborConfig, error) {
	var cfg models.IPv6NeighborConfig
	err := db.DB.First(&cfg).Error
	if err != nil {
		return models.IPv6NeighborConfig{}, err
	}
	return cfg, nil
}


func (m *Manager) AddPortMapping(publicIP string, publicPort, publicPortEnd int, containerIP string, containerPort, containerPortEnd int, protocol, iface string) error {
	return m.addPortDNAT(publicIP, publicPort, publicPortEnd, containerIP, containerPort, containerPortEnd, protocol, iface)
}

func (m *Manager) RemovePortMapping(publicIP string, publicPort, publicPortEnd int, containerIP string, containerPort, containerPortEnd int, protocol, iface string) error {
	return m.removePortDNAT(publicIP, publicPort, publicPortEnd, containerIP, containerPort, containerPortEnd, protocol, iface)
}

func (m *Manager) restorePortMappings() {
	defer network.NotifyRebuildDone()
	var mappings []models.PortMappingV6
	if err := db.DB.Where("status = ?", "active").Find(&mappings).Error; err != nil {
		logger.Warn("查询IPv6端口映射失败: %v", err)
		return
	}

	if len(mappings) == 0 {
		return
	}

	logger.Info("重建 %d 个IPv6端口映射...", len(mappings))
	restored := 0

	for _, mapping := range mappings {
		publicPortEnd := mapping.PublicPortEnd
		if publicPortEnd == 0 {
			publicPortEnd = mapping.PublicPort
		}
		containerPortEnd := mapping.ContainerPortEnd
		if containerPortEnd == 0 {
			containerPortEnd = mapping.ContainerPort
		}
		if err := m.addPortDNAT(mapping.ForwardIP, mapping.PublicPort, publicPortEnd, mapping.ContainerIP, mapping.ContainerPort, containerPortEnd, mapping.Protocol, mapping.Interface); err != nil {
			logger.Warn("重建IPv6端口映射失败 %s:%d: %v", mapping.ForwardIP, mapping.PublicPort, err)
			continue
		}
		restored++
	}

	if restored > 0 {
		logger.OK("IPv6端口映射重建: %d 条", restored)
	}
}
