package system

import (
	"github.com/gin-gonic/gin"
	"lxdapi/internal/monitor"
	"lxdapi/pkg/response"
)

// GetTraffic 获取流量信息
// @Summary 获取容器流量
// @Description 获取指定容器的流量使用情况
// @Tags System API - 流量管理
// @Accept json
// @Produce json
// @Param name query string true "容器名称"
// @Success 200 {object} response.Response "获取成功"
// @Failure 400 {object} response.Response "缺少容器名称"
// @Failure 500 {object} response.Response "获取失败"
// @Security ApiKeyAuth
// @Router /api/system/traffic [get]
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
// @Summary 重置容器流量
// @Description 重置指定容器的流量统计
// @Tags System API - 流量管理
// @Accept json
// @Produce json
// @Param name query string true "容器名称"
// @Success 200 {object} response.Response "重置成功"
// @Failure 400 {object} response.Response "缺少容器名称"
// @Failure 500 {object} response.Response "重置失败"
// @Security ApiKeyAuth
// @Router /api/system/traffic/reset [post]
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

