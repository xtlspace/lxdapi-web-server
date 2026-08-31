package admin

import (
	"github.com/gin-gonic/gin"
	"lxdapi/internal/cache"
	"lxdapi/pkg/response"
)

// GetContainersCache 获取容器列表
func GetContainersCache(c *gin.Context) {
	containers := cache.GetAllContainersFromDB()
	
	response.Success(c, gin.H{
		"data":       containers,
		"count":      len(containers),
		"from_cache": false,
	})
}

