package admin

import (
	"github.com/gin-gonic/gin"
	"lxdapi/internal/service"
	"lxdapi/models"
	"lxdapi/pkg/logger"
	"lxdapi/pkg/response"
)

var brandService *service.BrandService

func InitBrandService() {
	brandService = service.NewBrandService()
}

// GetBrandSettings 获取品牌设置
func GetBrandSettings(c *gin.Context) {
	settings, err := brandService.GetSettings()
	if err != nil {
		logger.Error("获取品牌设置失败: %v", err)
		response.Error(c, 500, "获取品牌设置失败")
		return
	}
	
	response.Success(c, settings)
}

// UpdateBrandSettings 更新品牌设置（统一接口）
func UpdateBrandSettings(c *gin.Context) {
	reset := c.Query("reset")
	if reset == "true" {
		if err := brandService.ResetToDefault(); err != nil {
			response.Error(c, 500, "重置失败")
			return
		}
		
		logger.OK("品牌设置已重置")
		response.Success(c, "重置成功")
		return
	}

	var req models.BrandSettings
	
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误: "+err.Error())
		return
	}
	
	if len(req.AdminSystemName) == 0 || len(req.AdminSystemName) > 100 {
		response.Error(c, 400, "管理后台系统名称长度必须在1-100字符之间")
		return
	}
	if len(req.AdminSystemTitle) == 0 || len(req.AdminSystemTitle) > 100 {
		response.Error(c, 400, "管理后台浏览器标题长度必须在1-100字符之间")
		return
	}
	if len(req.AdminLoginTitle) == 0 || len(req.AdminLoginTitle) > 100 {
		response.Error(c, 400, "管理后台登录标题长度必须在1-100字符之间")
		return
	}

	if len(req.ContainerSystemName) == 0 || len(req.ContainerSystemName) > 100 {
		response.Error(c, 400, "容器面板系统名称长度必须在1-100字符之间")
		return
	}
	if len(req.ContainerSystemTitle) == 0 || len(req.ContainerSystemTitle) > 100 {
		response.Error(c, 400, "容器面板浏览器标题长度必须在1-100字符之间")
		return
	}
	if len(req.ContainerLoginTitle) == 0 || len(req.ContainerLoginTitle) > 100 {
		response.Error(c, 400, "容器面板登录标题长度必须在1-100字符之间")
		return
	}
	
	if err := brandService.UpdateSettings(&req); err != nil {
		response.Error(c, 500, "更新品牌设置失败")
		return
	}
	
	logger.OK("品牌设置已更新")
	response.Success(c, "更新成功")
}