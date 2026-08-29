package models

import (
	"fmt"
	"regexp"
	"time"

	"gorm.io/gorm"
)

var containerNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

func ValidateContainerName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("容器名不能为空")
	}
	if len(name) > 63 {
		return fmt.Errorf("容器名长度不能超过63个字符")
	}
	if !containerNameRegex.MatchString(name) {
		return fmt.Errorf("容器名只允许字母、数字、下划线和连字符，且必须以字母或数字开头")
	}
	return nil
}

type Container struct {
	gorm.Model
	Name         string `gorm:"uniqueIndex;size:255"`
	Image        string `gorm:"size:255"`
	Password     string `gorm:"size:255"`
	Status       string `gorm:"size:50"`
	CPU          int
	Memory       int
	Disk         int
	Ingress      int
	Egress       int
	TrafficLimit int
	PrivateIP    string `gorm:"size:50"`
	PrivateIPv6  string `gorm:"size:100"`
	MacAddress   string `gorm:"size:50"`
	Remark       string `gorm:"size:20"`
	
	AllowNesting bool
	MemorySwap   bool
	Privileged   bool
	
	CPUAllowance    int
	IORead          int
	IOWrite         int
	ProcessesLimit  int
	
	MemoryUsageRaw uint64
	DiskUsageRaw   uint64
	TrafficUsage   string  `gorm:"size:50"` // 已用流量，单位GB，4位小数文本
	RxBytes        int64   `gorm:"default:0"`
	TxBytes        int64   `gorm:"default:0"`
	Locked         bool    `gorm:"default:false"`
	LastReset      time.Time
	IPv4PoolLimit   int
	IPv4MappingLimit int
	IPv6MappingLimit int
	ReverseProxyLimit int
	NetworkMode     string  `gorm:"size:50"`
	ConfigJSON      string  `gorm:"type:text"`
	CreatedAtLXD    *time.Time
	LastSync        *time.Time
}

type CreateContainerRequest struct {
	Name            string `json:"name" binding:"required"`
	Password        string `json:"password"`
	Image           string `json:"image" binding:"required"`
	CPU             int    `json:"cpu"`
	Memory          int    `json:"memory"`
	Disk            int    `json:"disk"`
	Ingress         int    `json:"ingress"`
	Egress          int    `json:"egress"`
	TrafficLimit    int    `json:"traffic_limit"`
	AllowNesting    bool   `json:"allow_nesting"`
	MemorySwap      bool   `json:"memory_swap"`
	Privileged      bool   `json:"privileged"`
	Remark          string `json:"remark"`
	IPv4PoolLimit    int    `json:"ipv4_pool_limit"`
	IPv4MappingLimit int    `json:"ipv4_mapping_limit"`
	IPv6MappingLimit int    `json:"ipv6_mapping_limit"`
	ReverseProxyLimit int   `json:"reverse_proxy_limit"`
	CPUAllowance     int    `json:"cpu_allowance"`
	IORead          int    `json:"io_read"`
	IOWrite         int    `json:"io_write"`
	ProcessesLimit  int    `json:"processes_limit"`
}

type UpdateContainerConfigRequest struct {
	CPU               *int    `json:"cpu"`
	Memory            *int    `json:"memory"`
	Disk              *int    `json:"disk"`
	Ingress           *int    `json:"ingress"`
	Egress            *int    `json:"egress"`
	CPUAllowance      *int    `json:"cpu_allowance"`
	IORead            *int    `json:"io_read"`
	IOWrite           *int    `json:"io_write"`
	ProcessesLimit    *int    `json:"processes_limit"`
	TrafficLimit      *int    `json:"traffic_limit"`
	IPv4PoolLimit     *int    `json:"ipv4_pool_limit"`
	IPv4MappingLimit  *int    `json:"ipv4_mapping_limit"`
	IPv6MappingLimit  *int    `json:"ipv6_mapping_limit"`
	ReverseProxyLimit *int    `json:"reverse_proxy_limit"`
	Remark            *string `json:"remark"`
	Privileged        *bool   `json:"privileged"`
	MemorySwap        *bool   `json:"memory_swap"`
	AllowNesting      *bool   `json:"allow_nesting"`
}

