package admin

import (
	"github.com/gin-gonic/gin"
	"lxdapi/internal/service"
	"lxdapi/models"
	"lxdapi/pkg/logger"
	"lxdapi/pkg/response"
)

// GetTemplateList 获取模板列表
// @Summary 获取模板列表
// @Description 获取所有容器模板
// @Tags Admin API - 模板管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "获取成功"
// @Failure 500 {object} response.Response "获取失败"
// @Security SessionAuth
// @Router /api/admin/templates [get]
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
// @Summary 同步模板
// @Description 从LXD同步模板到数据库
// @Tags Admin API - 模板管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "同步成功"
// @Failure 500 {object} response.Response "同步失败"
// @Security SessionAuth
// @Router /api/admin/templates/sync [post]
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
// @Summary 删除模板
// @Description 删除指定模板，支持单个和批量删除
// @Tags Admin API - 模板管理
// @Accept json
// @Produce json
// @Param fingerprint path string false "模板指纹（单个删除）"
// @Param fingerprints query string false "模板指纹列表，逗号分隔（批量删除）"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "缺少参数"
// @Failure 500 {object} response.Response "删除失败"
// @Security SessionAuth
// @Router /api/admin/templates/:fingerprint [delete]
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
// @Summary 批量删除模板
// @Description 批量删除多个模板
// @Tags Admin API - 模板管理
// @Accept json
// @Produce json
// @Param request body models.BatchDeleteTemplateRequest true "模板指纹列表"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "删除失败"
// @Security SessionAuth
// @Router /api/admin/templates/batch-delete [post]
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
