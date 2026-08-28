package admin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"lxdapi/internal/lxc"
	"lxdapi/pkg/response"
)

func GetNetworkNATStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := lxc.DefaultClient()
	result := gin.H{}

	var v4Config map[string]interface{}
	err := client.Get(ctx, "/1.0/networks/lxdbr0", &v4Config)
	if err != nil {
		result["ipv4_nat"] = false
		result["ipv6_nat"] = false
	} else {
		config, ok := v4Config["config"].(map[string]interface{})
		if !ok {
			config = make(map[string]interface{})
		}
		result["ipv4_nat"] = config["ipv4.nat"] == "true"
		result["ipv6_nat"] = config["ipv6.nat"] == "true"
	}

	response.Success(c, result)
}

func SetNetworkNATStatus(c *gin.Context) {
	var req struct {
		IPv4NAT *bool `json:"ipv4_nat"`
		IPv6NAT *bool `json:"ipv6_nat"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := lxc.DefaultClient()

	if req.IPv4NAT != nil {
		value := "false"
		if *req.IPv4NAT {
			value = "true"
		}
		err := client.Patch(ctx, "/1.0/networks/lxdbr0", map[string]interface{}{
			"config": map[string]interface{}{"ipv4.nat": value},
		})
		if err != nil {
			response.Error(c, 500, fmt.Sprintf("设置IPv4 NAT失败: %v", err))
			return
		}
	}

	if req.IPv6NAT != nil {
		value := "false"
		if *req.IPv6NAT {
			value = "true"
		}
		err := client.Patch(ctx, "/1.0/networks/lxdbr0", map[string]interface{}{
			"config": map[string]interface{}{"ipv6.nat": value},
		})
		if err != nil {
			response.Error(c, 500, fmt.Sprintf("设置IPv6 NAT失败: %v", err))
			return
		}
	}

	response.Success(c, "设置成功")
}

func FormatNetworkStatus(v4nat, v6nat bool) string {
	var parts []string
	if v4nat {
		parts = append(parts, "IPv4 NAT: ON")
	} else {
		parts = append(parts, "IPv4 NAT: OFF")
	}
	if v6nat {
		parts = append(parts, "IPv6 NAT: ON")
	} else {
		parts = append(parts, "IPv6 NAT: OFF")
	}
	return strings.Join(parts, " | ")
}
