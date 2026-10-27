package handlers

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"lxcapi/internal/console"
	"lxcapi/internal/core"
	"lxcapi/internal/db"
	"lxcapi/internal/service"
	"lxcapi/models"
	"lxcapi/pkg/utils"
	"net/http"
)

func AdminLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/admin_login.html", utils.MergeTemplateData(c, gin.H{}))
}

func AdminLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, core.AdminBasePath()+"/login")
}

func AdminDashboard(c *gin.Context) {
	token := c.Query("token")
	if token != "" {
		accessToken, err := service.ValidateAccessToken(token)
		if err == nil && accessToken.Type == "admin" {
			session := sessions.Default(c)
			session.Set("admin_logged_in", true)
			session.Set("admin_username", accessToken.Target)
			session.Save()
			c.Redirect(http.StatusFound, core.AdminBasePath()+"/dashboard")
			return
		}
	}

	c.HTML(http.StatusOK, "admin/admin_dashboard.html", utils.MergeTemplateData(c, gin.H{
		"username": "admin",
	}))
}

func AdminContainers(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/admin_containers.html", utils.MergeTemplateData(c, gin.H{
		"username": "admin",
	}))
}

func AdminContainerDetail(c *gin.Context) {
	containerName := c.Param("name")
	c.HTML(http.StatusOK, "admin/admin_container_detail.html", utils.MergeTemplateData(c, gin.H{
		"username":       "admin",
		"container_name": containerName,
	}))
}

func AdminTasks(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/admin_tasks.html", utils.MergeTemplateData(c, gin.H{
		"username": "admin",
	}))
}

func AdminIPPoolV4(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/admin_ip_pool_v4.html", utils.MergeTemplateData(c, gin.H{
		"username": "admin",
	}))
}

func AdminPortMappingV4(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/admin_port_mapping_v4.html", utils.MergeTemplateData(c, gin.H{
		"username": "admin",
	}))
}

func AdminPortMappingV6(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/admin_port_mapping_v6.html", utils.MergeTemplateData(c, gin.H{
		"username": "admin",
	}))
}

func AdminIPv6Neighbor(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/admin_ipv6_neighbor.html", utils.MergeTemplateData(c, gin.H{
		"username": "admin",
	}))
}

func AdminTemplates(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/admin_templates.html", utils.MergeTemplateData(c, gin.H{
		"username": "admin",
	}))
}

func AdminNginx(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/admin_nginx.html", utils.MergeTemplateData(c, gin.H{
		"username": "admin",
	}))
}

func ConsolePage(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "缺少访问令牌",
		})
		return
	}

	containerName, valid, errMsg := console.ValidateToken(token)
	if !valid {
		c.HTML(http.StatusUnauthorized, "error.html", gin.H{
			"error": errMsg,
		})
		return
	}

	var container models.Container
	image := "未知"
	if err := db.DB.Where("name = ?", containerName).First(&container).Error; err == nil {
		if container.Image != "" {
			image = container.Image
		}
	}

	c.HTML(http.StatusOK, "console.html", gin.H{
		"hostname": containerName,
		"image":    image,
	})
}

