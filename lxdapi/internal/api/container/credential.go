package container

import (
	"github.com/gin-gonic/gin"
	"lxdapi/internal/service"
	"lxdapi/pkg/logger"
	"lxdapi/pkg/response"
)

// RegenerateHash 重置当前容器的访问码
func RegenerateHash(c *gin.Context) {
	containerName, _ := c.Get("container_name")
	name := containerName.(string)

	credential, err := service.RegenerateContainerHash(name)
	if err != nil {
		logger.Error("重置容器访问码失败: %v", err)
		response.Error(c, 500, err.Error())
		return
	}

	logger.OK("重置容器访问码成功: %s", name)
	response.Success(c, gin.H{
		"container_name": credential.ContainerName,
		"hash":           credential.Hash,
	})
}
