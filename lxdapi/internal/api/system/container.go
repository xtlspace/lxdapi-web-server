package system

import (
	"context"
	"github.com/gin-gonic/gin"
	"lxdapi/internal/executor"
	"lxdapi/internal/service"
	"lxdapi/models"
	"lxdapi/pkg/response"
)

var containerService *service.ContainerService

func InitContainerService(svc *service.ContainerService) {
	containerService = svc
}

// CreateContainer 创建容器
func CreateContainer(c *gin.Context) {
	var req models.CreateContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	params := map[string]interface{}{
		"name":               req.Name,
		"image":              req.Image,
		"cpu":                req.CPU,
		"memory":             req.Memory,
		"disk":               req.Disk,
		"ingress":            req.Ingress,
		"egress":             req.Egress,
		"traffic_limit":      req.TrafficLimit,
		"allow_nesting":      req.AllowNesting,
		"memory_swap":        req.MemorySwap,
		"privileged":         req.Privileged,
		"ipv4_pool_limit":    req.IPv4PoolLimit,
		"ipv4_mapping_limit": req.IPv4MappingLimit,
		"ipv6_mapping_limit": req.IPv6MappingLimit,
		"password":           req.Password,
		"cpu_allowance":      req.CPUAllowance,
		"io_read":            req.IORead,
		"io_write":           req.IOWrite,
		"processes_limit":    req.ProcessesLimit,
	}

	task, err := executor.CreateTask(req.Name, "create", params, func(ctx context.Context) error {
		return containerService.Create(ctx, &req)
	})

	if err != nil {
		response.Error(c, 500, "创建任务失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"name":    req.Name,
		"task_id": task.ID,
		"message": "容器创建任务已提交，正在后台执行",
	})
}

// ListContainers 获取容器列表
func ListContainers(c *gin.Context) {
	containers, err := containerService.List()
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, containers)
}

// GetContainer 获取容器详情（包含状态）
func GetContainer(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.Error(c, 400, "缺少容器名称")
		return
	}

	container, err := containerService.Get(name)
	if err != nil {
		response.Error(c, 404, "容器不存在")
		return
	}

	ctx := c.Request.Context()
	status, _ := containerService.GetStatus(ctx, name)

	response.Success(c, gin.H{
		"container": container,
		"status":    status,
	})
}

// DeleteContainer 删除容器
func DeleteContainer(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.Error(c, 400, "缺少容器名称")
		return
	}

	container, err := containerService.Get(name)
	if err != nil {
		response.Error(c, 404, "容器不存在")
		return
	}

	params := map[string]interface{}{
		"container_name": name,
		"image":          container.Image,
		"status":         container.Status,
	}

	task, err := executor.CreateTask(name, "delete", params, func(ctx context.Context) error {
		return containerService.Delete(ctx, name)
	})

	if err != nil {
		response.Error(c, 500, "创建任务失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"task_id": task.ID,
		"message": "容器删除任务已提交，正在后台执行",
	})
}

// ContainerAction 容器操作统一接口
func ContainerAction(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.Error(c, 400, "缺少容器名称")
		return
	}

	action := c.Query("action")
	if action == "" {
		response.Error(c, 400, "缺少操作类型")
		return
	}

	container, err := containerService.Get(name)
	if err != nil {
		response.Error(c, 404, "容器不存在")
		return
	}

	switch action {
	case "start":
		err := executor.CreateSyncTask(name, "start", func(ctx context.Context) error {
			return containerService.Start(ctx, name)
		})
		if err != nil {
			response.Error(c, 500, err.Error())
			return
		}
		response.Success(c, "容器启动成功")

	case "stop":
		err := executor.CreateSyncTask(name, "stop", func(ctx context.Context) error {
			return containerService.Stop(ctx, name)
		})
		if err != nil {
			response.Error(c, 500, err.Error())
			return
		}
		response.Success(c, "容器停止成功")

	case "restart":
		err := executor.CreateSyncTask(name, "restart", func(ctx context.Context) error {
			return containerService.Restart(ctx, name)
		})
		if err != nil {
			response.Error(c, 500, err.Error())
			return
		}
		response.Success(c, "容器重启成功")

	case "pause":
		err := executor.CreateSyncTask(name, "pause", func(ctx context.Context) error {
			return containerService.Pause(ctx, name)
		})
		if err != nil {
			response.Error(c, 500, err.Error())
			return
		}
		response.Success(c, "容器暂停成功")

	case "resume":
		err := executor.CreateSyncTask(name, "resume", func(ctx context.Context) error {
			return containerService.Resume(ctx, name)
		})
		if err != nil {
			response.Error(c, 500, err.Error())
			return
		}
		response.Success(c, "容器恢复成功")

	case "reinstall":
		var req struct {
			Image    string `json:"image"`
			Password string `json:"password"`
		}
		c.ShouldBindJSON(&req)

		finalImage := req.Image
		if finalImage == "" {
			finalImage = container.Image
		}

		params := map[string]interface{}{
			"container_name": name,
			"original_image": container.Image,
			"new_image":      finalImage,
			"password_set":   req.Password != "",
			"image":          finalImage,
			"password":       req.Password,
		}

		task, err := executor.CreateTask(name, "reinstall", params, func(ctx context.Context) error {
			return containerService.Reinstall(ctx, name, req.Image, req.Password)
		})
		if err != nil {
			response.Error(c, 500, "创建任务失败: "+err.Error())
			return
		}
		response.Success(c, gin.H{
			"task_id": task.ID,
			"message": "重装任务已创建，正在后台执行",
		})

	case "reset-password":
		var req struct {
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, 400, "参数错误: 缺少password")
			return
		}

		err := executor.CreateSyncTask(name, "reset_password", func(ctx context.Context) error {
			return containerService.ResetPassword(ctx, name, req.Password)
		})
		if err != nil {
			response.Error(c, 500, err.Error())
			return
		}
		response.Success(c, "密码重置成功")

	default:
		response.Error(c, 400, "不支持的操作类型: "+action)
	}
}

// UpdateContainerConfig 更新容器配置
func UpdateContainerConfig(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.Error(c, 400, "缺少容器名称")
		return
	}

	if _, err := containerService.Get(name); err != nil {
		response.Error(c, 404, "容器不存在")
		return
	}

	var req models.UpdateContainerConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	params := map[string]interface{}{
		"container_name": name,
		"config":         req,
	}

	task, err := executor.CreateTask(name, "update_config", params, func(ctx context.Context) error {
		_, err := containerService.UpdateConfig(ctx, name, &req)
		return err
	})
	if err != nil {
		response.Error(c, 500, "创建任务失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"name":    name,
		"task_id": task.ID,
		"message": "配置更新任务已提交",
	})
}
