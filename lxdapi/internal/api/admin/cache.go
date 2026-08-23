package admin

import (
	"github.com/gin-gonic/gin"
	"lxdapi/internal/cache"
	"lxdapi/pkg/response"
)

// GetContainersCache 获取容器缓存
// @Summary 获取容器缓存
// @Description 从缓存获取所有容器信息
// @Tags Admin API - 缓存管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "获取成功"
// @Security SessionAuth
// @Router /api/admin/cache/containers [get]
func GetContainersCache(c *gin.Context) {
	containers := cache.GetAllContainersCache()
	
	response.Success(c, gin.H{
		"data":       containers,
		"count":      len(containers),
		"from_cache": true,
	})
}

