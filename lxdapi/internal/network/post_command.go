package network

import (
	"lxdapi/internal/core"
	"lxdapi/pkg/logger"
	"os/exec"
	"sync"
	"sync/atomic"
)

// 统计 ipv4/ipv6 两条重建链的完成次数
var rebuildDone atomic.Int32

// 保证 start_post_command 只执行一次
var postOnce sync.Once

// NotifyRebuildDone 由 ipv4/ipv6 各自的重建链在结束时调用。
// 当两条重建链都完成后，由最后一个完成者触发 start_post_command。
func NotifyRebuildDone() {
	if rebuildDone.Add(1) >= 2 {
		postOnce.Do(runStartPostCommand)
	}
}

func runStartPostCommand() {
	cmd := core.GlobalConfig.Network.StartPostCommand
	if cmd == "" {
		return
	}
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		logger.Error("[network] start_post_command 执行失败: %v, 输出: %s", err, string(out))
		return
	}
	logger.OK("[network] start_post_command 执行成功: %s", string(out))
}
