package admin

import (
	"github.com/gin-gonic/gin"
	"lxcapi/internal/cache"
	"lxcapi/pkg/response"
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

