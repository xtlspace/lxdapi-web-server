package lxc

import (
	"context"
	"fmt"
)

type ContainerInfo struct {
	Name         string                 `json:"name"`
	Status       string                 `json:"status"`
	Type         string                 `json:"type"`
	Architecture string                 `json:"architecture"`
	PID          int                    `json:"pid"`
	Created      string                 `json:"created_at"`
	LastUsed     string                 `json:"last_used_at"`
	Config       map[string]interface{} `json:"config"`
	Devices      map[string]interface{} `json:"devices"`
	State        *ContainerState        `json:"state"`
}

type ContainerState struct {
	Status     string                 `json:"status"`
	StatusCode int                    `json:"status_code"`
	Disk       map[string]interface{} `json:"disk"`
	Memory     map[string]interface{} `json:"memory"`
	Network    map[string]interface{} `json:"network"`
	Pid        int                    `json:"pid"`
	Processes  int                    `json:"processes"`
	CPU        map[string]interface{} `json:"cpu"`
}

func (c *Client) GetContainerInfo(ctx context.Context, name string) (*ContainerInfo, error) {
	var info ContainerInfo
	err := c.get(ctx, fmt.Sprintf("/1.0/instances/%s?recursion=2", name), &info)
	if err != nil {
		return nil, fmt.Errorf("获取容器信息失败: %v", err)
	}

	var state ContainerState
	if err := c.get(ctx, fmt.Sprintf("/1.0/instances/%s/state", name), &state); err == nil {
		info.State = &state
	}

	return &info, nil
}
