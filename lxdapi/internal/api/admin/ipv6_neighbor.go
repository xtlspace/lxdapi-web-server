package admin

import (
	"github.com/gin-gonic/gin"

	"lxdapi/internal/db"
	"lxdapi/internal/ipv6"
	"lxdapi/internal/monitor"
	"lxdapi/models"
	"lxdapi/pkg/response"
)

func GetIPv6NeighborConfig(c *gin.Context) {
	var config models.IPv6NeighborConfig

	if err := db.DB.First(&config).Error; err != nil {
		config = models.IPv6NeighborConfig{
			Enabled:  false,
			ProxyNDP: false,
			Iface:    "",
			Prefix:   "",
			Gateway:  "",
		}
		db.DB.Create(&config)
	}

	response.Success(c, config)
}

func SaveIPv6NeighborConfig(c *gin.Context) {
	var req struct {
		Enabled  bool   `json:"enabled"`
		ProxyNDP bool   `json:"proxy_ndp"`
		Iface    string `json:"iface"`
		Prefix   string `json:"prefix"`
		Gateway  string `json:"gateway"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	if req.Enabled {
		if req.Iface == "" {
			response.Error(c, 400, "网卡接口不能为空")
			return
		}
		if req.Prefix == "" {
			response.Error(c, 400, "IPv6前缀不能为空")
			return
		}
		if req.Gateway == "" {
			response.Error(c, 400, "网关地址不能为空")
			return
		}
	}

	var config models.IPv6NeighborConfig
	if err := db.DB.First(&config).Error; err != nil {
		config = models.IPv6NeighborConfig{}
	}

	wasEnabled := config.Enabled
	wasProxyNDP := config.ProxyNDP
	oldIface := config.Iface
	oldPrefix := config.Prefix

	config.Enabled = req.Enabled
	config.ProxyNDP = req.ProxyNDP
	config.Iface = req.Iface
	config.Prefix = req.Prefix
	config.Gateway = req.Gateway

	if err := db.DB.Save(&config).Error; err != nil {
		response.Error(c, 500, "保存失败: "+err.Error())
		return
	}

	if req.Enabled && !wasEnabled {
		ipv6.ApplySysctl()
	}

	monitor.ReloadNeighborConfig()

	// 代理NDP状态迁移处理
	proxyWasActive := wasEnabled && wasProxyNDP
	proxyNowActive := req.Enabled && req.ProxyNDP
	if proxyWasActive && !proxyNowActive {
		// 关闭代理NDP（或关闭主开关）：清理旧配置下的代理NDP条目
		ipv6.FlushProxyNDP(oldIface, oldPrefix)
	} else if proxyWasActive && proxyNowActive && (oldIface != req.Iface || oldPrefix != req.Prefix) {
		// 网卡/前缀变更：清理旧配置下的条目
		ipv6.FlushProxyNDP(oldIface, oldPrefix)
	}

	if proxyNowActive {
		monitor.RefreshProxyNDP()
	}

	response.Success(c, config)
}
