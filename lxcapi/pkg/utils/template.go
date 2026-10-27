package utils

import (
	"github.com/gin-gonic/gin"
	"lxcapi/internal/core"
	"lxcapi/pkg/version"
)

// MergeTemplateData 注入通用模板数据。
// 品牌信息已改为在静态页面中写死，不再从数据库读取。
func MergeTemplateData(c *gin.Context, data gin.H) gin.H {
	data["Version"] = version.Version
	data["AdminPath"] = core.AdminBasePath()
	return data
}
