package system

import (
	"github.com/gin-gonic/gin"
	"lxdapi/internal/monitor"
	"lxdapi/pkg/response"
)

// GetTraffic 获取流量信息
func GetTraffic(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		response.Error(c, 400, "缺少容器名称")
		return
	}
	
	if monitor.GlobalMonitor == nil {
		response.Error(c, 500, "流量监控未启用")
		return
	}
	
	t, err := monitor.GlobalMonitor.GetTraffic(name)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	
	response.Success(c, t)
}

// ResetTraffic 重置流量
func ResetTraffic(c *gin.Context) {
	name := c.Query("name")

	if name == "" {
		response.Error(c, 400, "缺少容器名称")
		return
	}

	if monitor.GlobalMonitor == nil {
		response.Error(c, 500, "流量监控未启用")
		return
	}

	if err := monitor.GlobalMonitor.ResetTraffic(name); err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, "流量已重置")
}

