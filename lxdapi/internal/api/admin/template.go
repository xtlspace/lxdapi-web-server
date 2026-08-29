package admin

import (
	"github.com/gin-gonic/gin"
	"lxdapi/internal/service"
	"lxdapi/models"
	"lxdapi/pkg/logger"
	"lxdapi/pkg/response"
)

// GetTemplateList 获取模板列表
func GetTemplateList(c *gin.Context) {
	svc := service.NewTemplateService()
	
	templates, err := svc.List()
	if err != nil {
		logger.Error("获取模板列表失败: %v", err)
		response.Error(c, 500, "获取模板列表失败: "+err.Error())
		return
	}
	
	response.Success(c, gin.H{
		"templates": templates,
		"count":     len(templates),
	})
}

// SyncTemplates 同步模板
func SyncTemplates(c *gin.Context) {
	svc := service.NewTemplateService()
	ctx := c.Request.Context()
	
	logger.Info("开始同步模板")
	
	added, updated, deleted, err := svc.SyncFromLXD(ctx)
	if err != nil {
		logger.Error("同步模板失败: %v", err)
		response.Error(c, 500, "同步模板失败: "+err.Error())
		return
	}
	
	logger.OK("模板同步完成")
	response.Success(c, gin.H{
		"added":   added,
		"updated": updated,
		"deleted": deleted,
		"message": "模板同步完成",
	})
}

// DeleteTemplate 删除模板（统一接口）
func DeleteTemplate(c *gin.Context) {
	fingerprint := c.Param("fingerprint")
	
	if fingerprint != "" {
		svc := service.NewTemplateService()
		ctx := c.Request.Context()
		
		if err := svc.Delete(ctx, fingerprint); err != nil {
			logger.Error("删除模板失败: %v", err)
			response.Error(c, 500, "删除模板失败: "+err.Error())
			return
		}
		
		logger.OK("模板删除成功: %s", fingerprint)
		response.Success(c, "模板删除成功")
		return
	}

	fp := c.Query("fingerprint")
	if fp != "" {
		svc := service.NewTemplateService()
		ctx := c.Request.Context()
		
		if err := svc.Delete(ctx, fp); err != nil {
			logger.Error("删除模板失败: %v", err)
			response.Error(c, 500, "删除模板失败: "+err.Error())
			return
		}
		
		logger.OK("模板删除成功: %s", fp)
		response.Success(c, "模板删除成功")
		return
	}

	response.Error(c, 400, "缺少fingerprint参数")
}

// BatchDeleteTemplates 批量删除模板
func BatchDeleteTemplates(c *gin.Context) {
	var req models.BatchDeleteTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误: "+err.Error())
		return
	}
	
	if len(req.Fingerprints) == 0 {
		response.Error(c, 400, "请选择要删除的模板")
		return
	}
	
	svc := service.NewTemplateService()
	ctx := c.Request.Context()
	
	logger.Info("批量删除模板，数量: %d", len(req.Fingerprints))
	
	success, failed, err := svc.BatchDelete(ctx, req.Fingerprints)
	if err != nil {
		logger.Error("批量删除模板失败: %v", err)
		response.Error(c, 500, "批量删除模板失败: "+err.Error())
		return
	}
	
	logger.OK("批量删除完成: 成功 %d, 失败 %d", success, failed)
	response.Success(c, gin.H{
		"deleted_count": success,
		"failed":        failed,
		"message":       "批量删除完成",
	})
}
