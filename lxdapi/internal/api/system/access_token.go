package system

import (
	"github.com/gin-gonic/gin"
	"lxdapi/internal/service"
	"lxdapi/pkg/response"
	"time"
)

func GetAdminAccessToken(c *gin.Context) {
	token, err := service.CreateAccessToken("admin", "admin", 30*time.Minute)
	if err != nil {
		response.Error(c, 500, "生成令牌失败")
		return
	}

	response.Success(c, gin.H{
		"token":      token.Token,
		"jump_url":   "/admin/dashboard?token=" + token.Token,
		"expires_at": token.ExpiresAt,
	})
}
