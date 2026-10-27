package admin

import (
	"github.com/gin-gonic/gin"

	"lxcapi/internal/db"
	"lxcapi/internal/ipv6"
	"lxcapi/internal/monitor"
	"lxcapi/models"
	"lxcapi/pkg/response"
)

func GetIPv6NeighborConfig(c *gin.Context) {
	var config models.IPv6NeighborConfig

	if err := db.DB.First(&config).Error; err != nil {
		config = models.IPv6NeighborConfig{
			Enabled:  false,
			Iface:   "",
			Prefix:  "",
			Gateway: "",
		}
		db.DB.Create(&config)
	}

	response.Success(c, config)
}

func SaveIPv6NeighborConfig(c *gin.Context) {
	var req struct {
		Enabled  bool   `json:"enabled"`
		Iface   string `json:"iface"`
		Prefix  string `json:"prefix"`
		Gateway string `json:"gateway"`
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
	oldIface := config.Iface
	oldPrefix := config.Prefix

	config.Enabled = req.Enabled
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

	if wasEnabled && (!req.Enabled || oldIface != req.Iface || oldPrefix != req.Prefix) {
		ipv6.FlushProxyNDP(oldIface, oldPrefix)
	}
	if req.Enabled {
		monitor.RefreshProxyNDP()
	}

	response.Success(c, config)
}
