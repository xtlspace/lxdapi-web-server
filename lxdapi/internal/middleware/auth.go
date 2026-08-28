package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"lxdapi/internal/core"
	"lxdapi/internal/db"
	"lxdapi/internal/service"
	"lxdapi/models"
	"lxdapi/pkg/auth"
	"lxdapi/pkg/response"
)

func SystemAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiHash := c.GetHeader("X-API-Hash")
		if apiHash != core.GlobalConfig.System.Security.APIHash {
			response.Error(c, 401, "系统级认证失败")
			c.Abort()
			return
		}
		c.Set("auth_level", auth.LevelSystem)
		c.Next()
	}
}

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		loggedIn := session.Get("admin_logged_in")
		
		if loggedIn == nil || loggedIn != true {
			response.Error(c, 401, "未登录或登录已过期，请重新登录")
			c.Abort()
			return
		}
		
		username := session.Get("admin_username")
		c.Set("auth_level", auth.LevelAdmin)
		c.Set("admin_username", username)
		c.Next()
	}
}

func AdminPageAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Query("token") != "" {
			c.Next()
			return
		}
		
		session := sessions.Default(c)
		loggedIn := session.Get("admin_logged_in")
		
		if loggedIn == nil || loggedIn != true {
			c.Redirect(302, "/admin/login")
			c.Abort()
			return
		}
		
		c.Next()
	}
}

func ContainerAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		containerHash := c.GetHeader("X-Container-Hash")
		if containerHash == "" {
			response.Error(c, 401, "缺少容器凭证")
			c.Abort()
			return
		}
		
		cred, err := service.GetContainerByHash(containerHash)
		if err != nil {
			response.Error(c, 401, "容器级认证失败")
			c.Abort()
			return
		}
		
		var container models.Container
		if err := db.DB.Where("name = ?", cred.ContainerName).First(&container).Error; err == nil &&
			container.Status == "frozen" && c.Request.Method != "GET" {
			response.Error(c, 401, "容器已暂停，无法进行任何操作")
			c.Abort()
			return
		}
		
		c.Set("auth_level", auth.LevelContainer)
		c.Set("container_name", cred.ContainerName)
		c.Next()
	}
}

