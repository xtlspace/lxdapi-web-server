package admin

import (
	"github.com/gin-gonic/gin"
	"lxdapi/internal/db"
	"lxdapi/models"
	"lxdapi/pkg/response"
	"lxdapi/pkg/system"
)

// GetDashboard 获取仪表盘数据
func GetDashboard(c *gin.Context) {
	var containerCount int64
	db.DB.Model(&models.Container{}).Count(&containerCount)

	var runningCount int64
	db.DB.Model(&models.Container{}).Where("status = ?", "running").Count(&runningCount)

	var stoppedCount int64
	db.DB.Model(&models.Container{}).Where("status = ?", "stopped").Count(&stoppedCount)

	var frozenCount int64
	db.DB.Model(&models.Container{}).Where("status = ?", "frozen").Count(&frozenCount)

	var portMappingV4Count int64
	db.DB.Model(&models.PortMappingV4{}).Count(&portMappingV4Count)
	
	var portMappingV6Count int64
	db.DB.Model(&models.PortMappingV6{}).Count(&portMappingV6Count)

	sysInfo := system.GetSystemInfo()
	
	response.Success(c, gin.H{
		"containers": gin.H{
			"total":   containerCount,
			"running": runningCount,
			"frozen":  frozenCount,
			"stopped": stoppedCount,
		},
		"port_mappings": gin.H{
			"ipv4": portMappingV4Count,
			"ipv6": portMappingV6Count,
		},
		"system": gin.H{
			"os":           sysInfo.OS,
			"arch":         sysInfo.Arch,
			"kernel":       sysInfo.Kernel,
			"distribution": sysInfo.Distribution,
			"lxd_version":  sysInfo.LXDVersion,
		},
	})
}

// GetHostStats 获取主机状态
func GetHostStats(c *gin.Context) {
	stats, err := system.GetHostStats()
	if err != nil {
		response.Error(c, 500, "获取主机状态失败: "+err.Error())
		return
	}
	
	response.Success(c, stats)
}

