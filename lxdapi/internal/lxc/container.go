package lxc

import (
	"context"
	"encoding/base64"
	"fmt"
	"lxdapi/pkg/logger"
	"strings"
	"time"
)

func (c *Client) GetContainerMAC(ctx context.Context, name string) (string, error) {
	info, err := c.GetContainerInfo(ctx, name)
	if err != nil {
		return "", err
	}

	if info.State == nil || info.State.Network == nil {
		return "", fmt.Errorf("容器网络信息不可用")
	}

	eth0Data, exists := info.State.Network["eth0"]
	if !exists {
		return "", fmt.Errorf("未找到 eth0 网卡")
	}

	if iface, ok := eth0Data.(map[string]interface{}); ok {
		if hwaddr, ok := iface["hwaddr"].(string); ok {
			return hwaddr, nil
		}
	}

	return "", fmt.Errorf("未找到MAC地址")
}

func (c *Client) CreateContainerWithConfig(ctx context.Context, name, image, storagePool string, cpu int, memory, disk, ingress, egress string, allowNesting, memorySwap, privileged bool, macAddress, cpuAllowance, ioRead, ioWrite string, processesLimit int) error {
	logger.Info("创建容器: %s, 镜像: %s, 存储池: %s", name, image, storagePool)

	config := map[string]interface{}{}
	if cpu > 0 {
		config["limits.cpu"] = fmt.Sprintf("%d", cpu)
	}
	if memory != "" {
		config["limits.memory"] = memory
	}
	config["security.nesting"] = fmt.Sprintf("%t", allowNesting)
	config["limits.memory.swap"] = fmt.Sprintf("%t", memorySwap)
	config["security.privileged"] = fmt.Sprintf("%t", privileged)
	if cpuAllowance != "" {
		config["limits.cpu.allowance"] = cpuAllowance
	}
	if processesLimit > 0 {
		config["limits.processes"] = fmt.Sprintf("%d", processesLimit)
	}

	devices := map[string]interface{}{
		"root": map[string]interface{}{
			"type": "disk",
			"path": "/",
			"pool": storagePool,
		},
	}
	if disk != "" {
		devices["root"].(map[string]interface{})["size"] = disk
	}
	if ioRead != "" {
		devices["root"].(map[string]interface{})["limits.read"] = ioRead
	}
	if ioWrite != "" {
		devices["root"].(map[string]interface{})["limits.write"] = ioWrite
	}
	if macAddress != "" || ingress != "" || egress != "" {
		nic := map[string]interface{}{
			"type":    "nic",
			"nictype": "bridged",
			"parent":  "lxdbr0",
		}
		if macAddress != "" {
			nic["hwaddr"] = macAddress
		}
		if ingress != "" {
			nic["limits.ingress"] = ingress
		}
		if egress != "" {
			nic["limits.egress"] = egress
		}
		devices["eth0"] = nic
	}

	body := map[string]interface{}{
		"name":   name,
		"source": map[string]interface{}{"type": "image", "alias": image},
		"config": config,
		"devices": devices,
	}

	data, err := c.post(ctx, "/1.0/instances", body)
	if err != nil {
		return fmt.Errorf("创建容器失败: %v", err)
	}
	opPath, err := parseOperationID(data)
	if err != nil {
		return fmt.Errorf("解析创建操作失败: %v", err)
	}
	if err := c.waitForOperation(ctx, opPath, 180); err != nil {
		return fmt.Errorf("等待创建完成失败: %v", err)
	}
	logger.OK("容器创建成功: %s", name)

	return nil
}

func (c *Client) CreateContainer(ctx context.Context, name, image string) error {
	return c.CreateContainerWithConfig(ctx, name, image, "default", 0, "", "", "", "", false, false, false, "", "", "", "", 0)
}

func (c *Client) StartContainer(ctx context.Context, name string) error {
	logger.Info("启动容器: %s", name)
	data, err := c.put(ctx, fmt.Sprintf("/1.0/instances/%s/state", name), map[string]interface{}{"action": "start"})
	if err != nil {
		return fmt.Errorf("启动容器失败: %v", err)
	}
	opPath, err := parseOperationID(data)
	if err != nil {
		return fmt.Errorf("解析启动操作失败: %v", err)
	}
	if err := c.waitForOperation(ctx, opPath, 60); err != nil {
		return fmt.Errorf("等待容器启动失败: %v", err)
	}
	for i := 0; i < 10; i++ {
		status, err := c.GetContainerStatus(ctx, name)
		if err == nil && status == "Running" {
			logger.OK("容器启动成功: %s", name)
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("容器 %s 启动后未达到 Running 状态", name)
}

func (c *Client) StopContainer(ctx context.Context, name string) error {
	logger.Info("停止容器: %s", name)
	data, err := c.put(ctx, fmt.Sprintf("/1.0/instances/%s/state", name), map[string]interface{}{"action": "stop", "force": true})
	if err != nil {
		return fmt.Errorf("停止容器失败: %v", err)
	}
	opPath, err := parseOperationID(data)
	if err != nil {
		return fmt.Errorf("解析停止操作失败: %v", err)
	}
	if err := c.waitForOperation(ctx, opPath, 60); err != nil {
		return fmt.Errorf("等待容器停止失败: %v", err)
	}
	logger.OK("容器停止成功: %s", name)
	return nil
}

func (c *Client) RestartContainer(ctx context.Context, name string) error {
	logger.Info("重启容器: %s", name)
	if err := c.StopContainer(ctx, name); err != nil {
		return fmt.Errorf("重启-停止失败: %v", err)
	}
	if err := c.StartContainer(ctx, name); err != nil {
		return fmt.Errorf("重启-启动失败: %v", err)
	}
	logger.OK("容器重启成功: %s", name)
	return nil
}

func (c *Client) PauseContainer(ctx context.Context, name string) error {
	logger.Info("暂停容器: %s", name)
	data, err := c.put(ctx, fmt.Sprintf("/1.0/instances/%s/state", name), map[string]interface{}{"action": "freeze"})
	if err != nil {
		return fmt.Errorf("暂停容器失败: %v", err)
	}
	opPath, err := parseOperationID(data)
	if err != nil {
		return fmt.Errorf("解析暂停操作失败: %v", err)
	}
	if err := c.waitForOperation(ctx, opPath, 60); err != nil {
		return fmt.Errorf("等待容器暂停失败: %v", err)
	}
	logger.OK("容器暂停成功: %s", name)
	return nil
}

func (c *Client) ResumeContainer(ctx context.Context, name string) error {
	logger.Info("恢复容器: %s", name)
	data, err := c.put(ctx, fmt.Sprintf("/1.0/instances/%s/state", name), map[string]interface{}{"action": "unfreeze"})
	if err != nil {
		return fmt.Errorf("恢复容器失败: %v", err)
	}
	opPath, err := parseOperationID(data)
	if err != nil {
		return fmt.Errorf("解析恢复操作失败: %v", err)
	}
	if err := c.waitForOperation(ctx, opPath, 60); err != nil {
		return fmt.Errorf("等待容器恢复失败: %v", err)
	}
	logger.OK("容器恢复成功: %s", name)
	return nil
}

func (c *Client) DeleteContainer(ctx context.Context, name string) error {
	logger.Info("删除容器: %s", name)
	if err := c.StopContainer(ctx, name); err != nil {
		return fmt.Errorf("停止容器失败: %v", err)
	}
	data, err := c.deleteReq(ctx, fmt.Sprintf("/1.0/instances/%s", name))
	if err != nil {
		return fmt.Errorf("删除容器失败: %v", err)
	}
	opPath, err := parseOperationID(data)
	if err != nil {
		return fmt.Errorf("解析删除操作失败: %v", err)
	}
	if err := c.waitForOperation(ctx, opPath, 60); err != nil {
		return fmt.Errorf("等待容器删除失败: %v", err)
	}
	logger.OK("容器删除成功: %s", name)
	return nil
}

func (c *Client) RebuildContainer(ctx context.Context, name, image string) error {
	logger.Info("重装容器: %s, 新镜像: %s", name, image)
	data, err := c.post(ctx, fmt.Sprintf("/1.0/instances/%s/rebuild", name), map[string]interface{}{
		"source": map[string]interface{}{
			"type":  "image",
			"alias": image,
		},
	})
	if err != nil {
		return fmt.Errorf("重装容器失败: %v", err)
	}
	opPath, err := parseOperationID(data)
	if err != nil {
		return fmt.Errorf("解析重装操作失败: %v", err)
	}
	if err := c.waitForOperation(ctx, opPath, 180); err != nil {
		return fmt.Errorf("等待重装完成失败: %v", err)
	}
	logger.OK("容器重装成功: %s", name)
	return nil
}

func (c *Client) GetContainerStatus(ctx context.Context, name string) (string, error) {
	var info struct {
		Status string `json:"status"`
	}
	err := c.get(ctx, fmt.Sprintf("/1.0/instances/%s", name), &info)
	if err != nil {
		return "", fmt.Errorf("获取容器状态失败: %v", err)
	}
	return info.Status, nil
}

func (c *Client) ContainerExists(ctx context.Context, name string) bool {
	var info map[string]interface{}
	err := c.get(ctx, fmt.Sprintf("/1.0/instances/%s", name), &info)
	return err == nil
}

func (c *Client) SetContainerConfig(ctx context.Context, name, key, value string) error {
	return c.patch(ctx, fmt.Sprintf("/1.0/instances/%s", name), map[string]interface{}{
		"config": map[string]interface{}{key: value},
	})
}

func (c *Client) SetRootPassword(ctx context.Context, name, password string) error {
	logger.Info("设置容器 %s 的root密码", name)

	encoded := base64.StdEncoding.EncodeToString([]byte("root:" + password))
	cmd := []string{"exec", name, "--", "sh", "-c", fmt.Sprintf("echo %s | base64 -d | chpasswd", encoded)}
	_, err := c.execIncus(ctx, cmd...)
	if err != nil {
		return fmt.Errorf("设置root密码失败: %v", err)
	}

	logger.OK("容器 %s 的root密码设置成功", name)
	return nil
}

func (c *Client) GetContainerIP(ctx context.Context, name string) (string, error) {
	info, err := c.GetContainerInfo(ctx, name)
	if err != nil {
		return "", err
	}

	if info.State == nil || info.State.Network == nil {
		return "", fmt.Errorf("容器网络信息不可用")
	}

	eth0Data, exists := info.State.Network["eth0"]
	if !exists {
		return "", fmt.Errorf("未找到 eth0 网卡")
	}

	if iface, ok := eth0Data.(map[string]interface{}); ok {
		if addresses, ok := iface["addresses"].([]interface{}); ok {
			for _, addr := range addresses {
				if addrMap, ok := addr.(map[string]interface{}); ok {
					if family, ok := addrMap["family"].(string); ok && family == "inet" {
						if address, ok := addrMap["address"].(string); ok {
							if !isInvalidIPv4(address) {
								logger.Info("从容器 %s eth0 网卡获取到IPv4: %s", name, address)
								return address, nil
							}
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("eth0 网卡未找到有效的IPv4地址")
}

func isInvalidIPv4(ip string) bool {
	if ip == "127.0.0.1" {
		return true
	}
	if strings.HasPrefix(ip, "169.254.") {
		return true
	}
	return false
}

func (c *Client) UpdateContainerConfig(ctx context.Context, name string, cpu int, memory, disk, ingress, egress, cpuAllowance, ioRead, ioWrite string, processesLimit int, privileged, memorySwap, allowNesting *bool) error {
	logger.Info("更新容器配置: %s", name)

	patch := map[string]interface{}{}

	config := map[string]interface{}{}
	if cpu > 0 {
		config["limits.cpu"] = fmt.Sprintf("%d", cpu)
	}
	if memory != "" {
		config["limits.memory"] = memory
	}
	if cpuAllowance != "" {
		config["limits.cpu.allowance"] = cpuAllowance
	}
	if processesLimit > 0 {
		config["limits.processes"] = fmt.Sprintf("%d", processesLimit)
	}
	if privileged != nil {
		config["security.privileged"] = fmt.Sprintf("%t", *privileged)
	}
	if memorySwap != nil {
		config["limits.memory.swap"] = fmt.Sprintf("%t", *memorySwap)
	}
	if allowNesting != nil {
		config["security.nesting"] = fmt.Sprintf("%t", *allowNesting)
	}
	if len(config) > 0 {
		patch["config"] = config
	}

	devices := map[string]interface{}{}
	needDevices := ingress != "" || egress != "" || disk != "" || ioRead != "" || ioWrite != ""
	var info *ContainerInfo
	if needDevices {
		var err error
		info, err = c.GetContainerInfo(ctx, name)
		if err != nil {
			return fmt.Errorf("获取容器配置失败: %v", err)
		}
	}
	if ingress != "" || egress != "" {
		eth0 := map[string]interface{}{}
		if devConfig, ok := info.Devices["eth0"].(map[string]interface{}); ok {
			if t, ok := devConfig["type"].(string); ok {
				eth0["type"] = t
			}
			if nt, ok := devConfig["nictype"].(string); ok {
				eth0["nictype"] = nt
			}
			if p, ok := devConfig["parent"].(string); ok {
				eth0["parent"] = p
			}
		}
		if ingress != "" {
			eth0["limits.ingress"] = ingress
		}
		if egress != "" {
			eth0["limits.egress"] = egress
		}
		devices["eth0"] = eth0
	}
	if disk != "" || ioRead != "" || ioWrite != "" {
		root := map[string]interface{}{}
		if devConfig, ok := info.Devices["root"].(map[string]interface{}); ok {
			if t, ok := devConfig["type"].(string); ok {
				root["type"] = t
			}
			if path, ok := devConfig["path"].(string); ok {
				root["path"] = path
			}
			if pool, ok := devConfig["pool"].(string); ok {
				root["pool"] = pool
			}
		}
		if disk != "" {
			root["size"] = disk
		}
		if ioRead != "" {
			root["limits.read"] = ioRead
		}
		if ioWrite != "" {
			root["limits.write"] = ioWrite
		}
		devices["root"] = root
	}
	if len(devices) > 0 {
		patch["devices"] = devices
	}

	if len(patch) == 0 {
		logger.Info("无配置需要更新")
		return nil
	}

	err := c.patch(ctx, fmt.Sprintf("/1.0/instances/%s", name), patch)
	if err != nil {
		return fmt.Errorf("更新容器配置失败: %v", err)
	}

	logger.OK("容器配置更新成功: %s", name)
	return nil
}

func (c *Client) GetContainerIPv6(ctx context.Context, name string) (string, error) {
	info, err := c.GetContainerInfo(ctx, name)
	if err != nil {
		return "", err
	}

	if info.State == nil || info.State.Network == nil {
		return "", fmt.Errorf("容器网络信息不可用")
	}

	eth0Data, exists := info.State.Network["eth0"]
	if !exists {
		return "", fmt.Errorf("未找到 eth0 网卡")
	}

	if iface, ok := eth0Data.(map[string]interface{}); ok {
		if addresses, ok := iface["addresses"].([]interface{}); ok {
			for _, addr := range addresses {
				if addrMap, ok := addr.(map[string]interface{}); ok {
					if family, ok := addrMap["family"].(string); ok && family == "inet6" {
						if address, ok := addrMap["address"].(string); ok {
							if !strings.HasPrefix(address, "fe80:") {
								logger.Info("从容器 %s eth0 网卡获取到IPv6: %s", name, address)
								return address, nil
							}
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("eth0 网卡未找到有效的IPv6地址")
}

func (c *Client) SetContainerIPv4Address(ctx context.Context, name, ip string) error {
	err := c.patch(ctx, fmt.Sprintf("/1.0/instances/%s", name), map[string]interface{}{
		"devices": map[string]interface{}{"eth0": map[string]interface{}{
			"ipv4.address": ip,
		}},
	})
	if err != nil {
		return fmt.Errorf("固定容器IPv4地址失败: %v", err)
	}
	logger.OK("容器 %s 的IPv4地址已固定: %s", name, ip)
	return nil
}
