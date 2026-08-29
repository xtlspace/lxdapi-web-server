package system

import (
	"github.com/gin-gonic/gin"
	"lxdapi/internal/ipv4"
	"lxdapi/internal/service"
	"lxdapi/pkg/response"
)

var ipv4Service *service.IPv4Service

func InitIPv4Service(svc *service.IPv4Service) {
	ipv4Service = svc
}

// GetContainerIP 获取容器IP地址（统一接口）
func GetContainerIP(c *gin.Context) {
	name := c.Query("container")
	if name == "" {
		response.Error(c, 400, "缺少容器名称")
		return
	}

	version := c.DefaultQuery("version", "all")

	if version == "v6" {
		response.Error(c, 400, "IPv6地址池功能已移除")
		return
	}

	result := gin.H{"container": name}

	if version == "v4" || version == "all" {
		if ipv4.GlobalManager == nil {
			result["ipv4"] = []string{}
			result["ipv4_count"] = 0
		} else {
			ips, err := ipv4.GlobalManager.GetContainerIPs(name)
			if err != nil {
				result["ipv4"] = []string{}
				result["ipv4_count"] = 0
			} else {
				result["ipv4"] = ips
				result["ipv4_count"] = len(ips)
			}
		}
	}

	response.Success(c, result)
}

// AllocateIP 分配IP地址（统一接口）
func AllocateIP(c *gin.Context) {
	version := c.Query("version")
	if version != "v4" {
		if version == "v6" {
			response.Error(c, 400, "IPv6地址池功能已移除")
			return
		}
		response.Error(c, 400, "version参数必须是v4")
		return
	}

	var req struct {
		Name  string `json:"name" binding:"required"`
		Count int    `json:"count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	if req.Count <= 0 {
		req.Count = 1
	}

	ctx := c.Request.Context()

	ips, err := ipv4Service.AllocateIPv4(ctx, req.Name, req.Count)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, gin.H{
		"container": req.Name,
		"ipv4":      ips,
		"count":     len(ips),
	})
}

// ReleaseIP 释放IP地址（统一接口）
func ReleaseIP(c *gin.Context) {
	version := c.Query("version")
	if version != "v4" {
		if version == "v6" {
			response.Error(c, 400, "IPv6地址池功能已移除")
			return
		}
		response.Error(c, 400, "version参数必须是v4")
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
		IP   string `json:"ip" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	if err := ipv4Service.ReleaseIPv4(req.Name, req.IP); err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, "IPv4释放成功")
}
