package service

import (
	"context"
	"fmt"
	"math/rand"

	"lxdapi/internal/db"
	"lxdapi/internal/ipv4"
	"lxdapi/internal/ipv6"
	"lxdapi/internal/lxc"
	"lxdapi/models"
	"lxdapi/pkg/logger"
)

type PortMappingService struct {
	lxcClient *lxc.Client
}

func NewPortMappingService() *PortMappingService {
	return &PortMappingService{
		lxcClient: lxc.DefaultClient(),
	}
}

func buildProtocols(protocol string) []string {
	protos := []string{}
	if protocol == "tcp" || protocol == "both" {
		protos = append(protos, "tcp", "both")
	}
	if protocol == "udp" || protocol == "both" {
		protos = append(protos, "udp", "both")
	}
	return protos
}

type portRange struct {
	start, end int
}

func checkPortConflict(publicPort, publicPortEnd int, existing []portRange) error {
	for _, e := range existing {
		if publicPort <= e.end && publicPortEnd >= e.start {
			if e.end == e.start {
				return fmt.Errorf("端口已被占用")
			}
			return fmt.Errorf("端口段 %d-%d 与已有端口段 %d-%d 冲突", publicPort, publicPortEnd, e.start, e.end)
		}
	}
	return nil
}

func findAvailablePortFromUsed(portStart, portEnd, portCount int, usedPorts map[int]bool) (int, error) {
	rangeSize := portEnd - portStart - portCount + 2
	if rangeSize <= 0 {
		return 0, fmt.Errorf("端口范围不足")
	}

	for attempt := 0; attempt < 100; attempt++ {
		startPort := portStart + rand.Intn(rangeSize)
		available := true
		for p := startPort; p < startPort+portCount; p++ {
			if usedPorts[p] {
				available = false
				break
			}
		}
		if available {
			return startPort, nil
		}
	}

	return 0, fmt.Errorf("无法找到可用的连续端口段")
}

func (s *PortMappingService) CheckV4PortRangeAvailable(publicIP string, publicPort, publicPortEnd int, protocol string) error {
	var existing []models.PortMappingV4
	if err := db.DB.Where("public_ip = ? AND protocol IN (?)", publicIP, buildProtocols(protocol)).Find(&existing).Error; err != nil {
		return fmt.Errorf("查询端口映射失败: %v", err)
	}
	ranges := make([]portRange, len(existing))
	for i, e := range existing {
		end := e.PublicPortEnd
		if end == 0 {
			end = e.PublicPort
		}
		ranges[i] = portRange{e.PublicPort, end}
	}
	if err := checkPortConflict(publicPort, publicPortEnd, ranges); err != nil {
		return fmt.Errorf("端口 %s:%d 已被占用", publicIP, publicPort)
	}
	return nil
}

func (s *PortMappingService) CheckV6PortRangeAvailable(publicIP string, publicPort, publicPortEnd int, protocol string) error {
	var existing []models.PortMappingV6
	if err := db.DB.Where("public_ip = ? AND protocol IN (?)", publicIP, buildProtocols(protocol)).Find(&existing).Error; err != nil {
		return fmt.Errorf("查询端口映射失败: %v", err)
	}
	ranges := make([]portRange, len(existing))
	for i, e := range existing {
		end := e.PublicPortEnd
		if end == 0 {
			end = e.PublicPort
		}
		ranges[i] = portRange{e.PublicPort, end}
	}
	if err := checkPortConflict(publicPort, publicPortEnd, ranges); err != nil {
		return fmt.Errorf("端口 %s:%d 已被占用", publicIP, publicPort)
	}
	return nil
}

func (s *PortMappingService) FindAvailableV4Port(portStart, portEnd, portCount int, protocol string) (int, error) {
	var existing []models.PortMappingV4
	if err := db.DB.Where("protocol IN (?)", buildProtocols(protocol)).Find(&existing).Error; err != nil {
		return 0, fmt.Errorf("查询端口映射失败: %v", err)
	}
	usedPorts := make(map[int]bool)
	for _, m := range existing {
		end := m.PublicPortEnd
		if end == 0 {
			end = m.PublicPort
		}
		for p := m.PublicPort; p <= end; p++ {
			usedPorts[p] = true
		}
	}
	return findAvailablePortFromUsed(portStart, portEnd, portCount, usedPorts)
}

func (s *PortMappingService) FindAvailableV6Port(portStart, portEnd, portCount int, protocol string) (int, error) {
	var existing []models.PortMappingV6
	if err := db.DB.Where("protocol IN (?)", buildProtocols(protocol)).Find(&existing).Error; err != nil {
		return 0, fmt.Errorf("查询端口映射失败: %v", err)
	}
	usedPorts := make(map[int]bool)
	for _, m := range existing {
		end := m.PublicPortEnd
		if end == 0 {
			end = m.PublicPort
		}
		for p := m.PublicPort; p <= end; p++ {
			usedPorts[p] = true
		}
	}
	return findAvailablePortFromUsed(portStart, portEnd, portCount, usedPorts)
}

func (s *PortMappingService) AllocateV4Mapping(ctx context.Context, containerName string, forwardIP string, displayIP string, publicPort, publicPortEnd, containerPort, containerPortEnd int, protocol, iface, description string) (*models.PortMappingV4, error) {
	if iface == "" {
		return nil, fmt.Errorf("必须指定网卡")
	}
	if forwardIP == "" {
		return nil, fmt.Errorf("必须指定转发IP")
	}
	if displayIP == "" {
		return nil, fmt.Errorf("必须指定公示IP")
	}

	var container models.Container
	if err := db.DB.Where("name = ?", containerName).First(&container).Error; err != nil {
		return nil, fmt.Errorf("容器不存在: %v", err)
	}
	containerIP := container.PrivateIP
	if containerIP == "" {
		return nil, fmt.Errorf("容器IP为空，请等待容器启动完成")
	}

	mapping := &models.PortMappingV4{
		ForwardIP: forwardIP, PublicIP: displayIP, PublicPort: publicPort,
		PublicPortEnd: publicPortEnd, ContainerName: containerName, ContainerIP: containerIP,
		ContainerPort: containerPort, ContainerPortEnd: containerPortEnd, Protocol: protocol,
		Status: "active", Interface: iface, Description: description,
	}
	if err := db.DB.Create(mapping).Error; err != nil {
		return nil, fmt.Errorf("保存端口映射失败: %v", err)
	}
	if ipv4.GlobalManager != nil {
		if err := ipv4.GlobalManager.AddPortMapping(forwardIP, publicPort, publicPortEnd, containerIP, containerPort, containerPortEnd, protocol, iface); err != nil {
			db.DB.Unscoped().Delete(mapping)
			return nil, fmt.Errorf("添加nftables规则失败: %v", err)
		}
	}
	logger.OK("IPv4端口映射创建成功: %s:%d -> %s:%d (%s)", displayIP, publicPort, containerIP, containerPort, protocol)
	return mapping, nil
}

func (s *PortMappingService) AllocateV6Mapping(ctx context.Context, containerName string, forwardIP string, displayIP string, publicPort, publicPortEnd, containerPort, containerPortEnd int, protocol, iface, description string) (*models.PortMappingV6, error) {
	if iface == "" {
		return nil, fmt.Errorf("必须指定网卡")
	}
	if forwardIP == "" {
		return nil, fmt.Errorf("必须指定转发IPv6")
	}
	if displayIP == "" {
		return nil, fmt.Errorf("必须指定公示IPv6")
	}

	var container models.Container
	if err := db.DB.Where("name = ?", containerName).First(&container).Error; err != nil {
		return nil, fmt.Errorf("容器不存在: %v", err)
	}
	containerIP := container.PrivateIPv6
	if containerIP == "" {
		return nil, fmt.Errorf("容器IPv6为空，请等待容器启动完成")
	}

	mapping := &models.PortMappingV6{
		ForwardIP: forwardIP, PublicIP: displayIP, PublicPort: publicPort,
		PublicPortEnd: publicPortEnd, ContainerName: containerName, ContainerIP: containerIP,
		ContainerPort: containerPort, ContainerPortEnd: containerPortEnd, Protocol: protocol,
		Status: "active", Interface: iface, Description: description,
	}
	if err := db.DB.Create(mapping).Error; err != nil {
		return nil, fmt.Errorf("保存IPv6端口映射失败: %v", err)
	}
	if ipv6.GlobalManager != nil {
		if err := ipv6.GlobalManager.AddPortMapping(forwardIP, publicPort, publicPortEnd, containerIP, containerPort, containerPortEnd, protocol, iface); err != nil {
			db.DB.Unscoped().Delete(mapping)
			return nil, fmt.Errorf("添加nftables规则失败: %v", err)
		}
	}
	logger.OK("IPv6端口映射创建成功: %s:%d -> %s:%d (%s)", displayIP, publicPort, containerIP, containerPort, protocol)
	return mapping, nil
}

func releaseV4NftMapping(mapping *models.PortMappingV4) error {
	if ipv4.GlobalManager == nil {
		return nil
	}
	forwardIP := mapping.ForwardIP
	if forwardIP == "" {
		forwardIP = mapping.PublicIP
	}
	publicPortEnd := mapping.PublicPortEnd
	if publicPortEnd == 0 {
		publicPortEnd = mapping.PublicPort
	}
	containerPortEnd := mapping.ContainerPortEnd
	if containerPortEnd == 0 {
		containerPortEnd = mapping.ContainerPort
	}
	return ipv4.GlobalManager.RemovePortMapping(forwardIP, mapping.PublicPort, publicPortEnd, mapping.ContainerIP, mapping.ContainerPort, containerPortEnd, mapping.Protocol, mapping.Interface)
}

func releaseV6NftMapping(mapping *models.PortMappingV6) error {
	if ipv6.GlobalManager == nil {
		return nil
	}
	forwardIP := mapping.ForwardIP
	if forwardIP == "" {
		forwardIP = mapping.PublicIP
	}
	publicPortEnd := mapping.PublicPortEnd
	if publicPortEnd == 0 {
		publicPortEnd = mapping.PublicPort
	}
	containerPortEnd := mapping.ContainerPortEnd
	if containerPortEnd == 0 {
		containerPortEnd = mapping.ContainerPort
	}
	return ipv6.GlobalManager.RemovePortMapping(forwardIP, mapping.PublicPort, publicPortEnd, mapping.ContainerIP, mapping.ContainerPort, containerPortEnd, mapping.Protocol, mapping.Interface)
}

func (s *PortMappingService) ReleaseV4Mapping(id uint) error {
	var mapping models.PortMappingV4
	if err := db.DB.First(&mapping, id).Error; err != nil {
		return fmt.Errorf("端口映射不存在: %v", err)
	}
	if err := releaseV4NftMapping(&mapping); err != nil {
		return fmt.Errorf("删除nftables规则失败: %v", err)
	}
	if err := db.DB.Unscoped().Delete(&mapping).Error; err != nil {
		return fmt.Errorf("删除端口映射记录失败: %v", err)
	}
	logger.OK("IPv4端口映射释放成功: %s:%d -> %s:%d", mapping.PublicIP, mapping.PublicPort, mapping.ContainerIP, mapping.ContainerPort)
	return nil
}

func (s *PortMappingService) ReleaseV6Mapping(id uint) error {
	var mapping models.PortMappingV6
	if err := db.DB.First(&mapping, id).Error; err != nil {
		return fmt.Errorf("端口映射不存在: %v", err)
	}
	if err := releaseV6NftMapping(&mapping); err != nil {
		return fmt.Errorf("删除nftables规则失败: %v", err)
	}
	if err := db.DB.Unscoped().Delete(&mapping).Error; err != nil {
		return fmt.Errorf("删除IPv6端口映射记录失败: %v", err)
	}
	logger.OK("IPv6端口映射释放成功: %s:%d -> %s:%d", mapping.PublicIP, mapping.PublicPort, mapping.ContainerIP, mapping.ContainerPort)
	return nil
}

func (s *PortMappingService) ListV4Mappings(containerName string) ([]models.PortMappingV4, error) {
	var mappings []models.PortMappingV4
	query := db.DB.Order("created_at DESC")
	if containerName != "" {
		query = query.Where("container_name = ?", containerName)
	}
	if err := query.Find(&mappings).Error; err != nil {
		return nil, err
	}
	return mappings, nil
}

func (s *PortMappingService) ListV6Mappings(containerName string) ([]models.PortMappingV6, error) {
	var mappings []models.PortMappingV6
	query := db.DB.Order("created_at DESC")
	if containerName != "" {
		query = query.Where("container_name = ?", containerName)
	}
	if err := query.Find(&mappings).Error; err != nil {
		return nil, err
	}
	return mappings, nil
}
