package ipv6

import (
	"os/exec"

	"lxdapi/pkg/logger"
)

// kernel sysctl parameters required for IPv6 neighbor request (NDP proxy) to work.
var neighborSysctls = []string{
	"net.ipv6.ip_nonlocal_bind=1",
	"net.ipv6.conf.all.proxy_ndp=1",
}

// ApplySysctl sets the kernel parameters required by the IPv6 neighbor request
// feature. Errors are logged but not returned as fatal (best-effort).
func ApplySysctl() {
	for _, param := range neighborSysctls {
		if err := runSysctl(param); err != nil {
			logger.Warn("设置内核参数失败 %s: %v", param, err)
			continue
		}
		logger.OK("内核参数已设置: %s", param)
	}
}

func runSysctl(param string) error {
	cmd := exec.Command("sysctl", "-w", param)
	_, err := cmd.CombinedOutput()
	return err
}
