package lxc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"lxdapi/internal/core"
	"lxdapi/pkg/logger"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Client struct {
	socket     string
	timeout    time.Duration
	httpClient *http.Client
	execPath   string
}

var (
	defaultClientOnce sync.Once
	defaultClient     *Client
)

func DefaultClient() *Client {
	defaultClientOnce.Do(func() {
		defaultClient = NewClient()
	})
	return defaultClient
}

func NewClient() *Client {
	cfg := core.GlobalConfig.LXC
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return net.DialTimeout("unix", cfg.Socket, time.Duration(cfg.Timeout)*time.Second)
		},
	}
	execPath, _ := exec.LookPath("incus")
	return &Client{
		socket:  cfg.Socket,
		timeout: time.Duration(cfg.Timeout) * time.Second,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   time.Duration(cfg.Timeout) * time.Second,
		},
		execPath: execPath,
	}
}

func (c *Client) SocketPath() string { return c.socket }

func (c *Client) doRequest(ctx context.Context, method, apiPath string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, "http://localhost"+apiPath, bodyReader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	logger.Info("Incus API: %s %s", method, apiPath)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error     string `json:"error"`
			ErrorCode int    `json:"error_code"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("Incus API 错误 %d: %s", resp.StatusCode, errResp.Error)
		}
		return nil, fmt.Errorf("Incus API 错误 %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *Client) get(ctx context.Context, apiPath string, result interface{}) error {
	data, err := c.doRequest(ctx, "GET", apiPath, nil)
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}

	var wrapper struct {
		Type     string          `json:"type"`
		Metadata json.RawMessage `json:"metadata"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return json.Unmarshal(data, result)
	}
	if wrapper.Type == "sync" && wrapper.Metadata != nil {
		return json.Unmarshal(wrapper.Metadata, result)
	}
	return json.Unmarshal(data, result)
}

func parseOperationID(data []byte) (string, error) {
	var resp struct {
		Type      string `json:"type"`
		Operation string `json:"operation"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("解析操作响应失败: %v", err)
	}
	if resp.Type != "async" || resp.Operation == "" {
		return "", fmt.Errorf("非异步操作响应: type=%s operation=%s", resp.Type, resp.Operation)
	}
	return resp.Operation, nil
}

func (c *Client) waitForOperation(ctx context.Context, operationPath string, timeoutSec int) error {
	apiPath := fmt.Sprintf("%s/wait?timeout=%d", operationPath, timeoutSec)
	data, err := c.doRequest(ctx, "GET", apiPath, nil)
	if err != nil {
		return err
	}
	var resp struct {
		Metadata struct {
			Status     string                 `json:"status"`
			StatusCode int                    `json:"status_code"`
			Err        string                 `json:"err"`
			Metadata   map[string]interface{} `json:"metadata"`
		} `json:"metadata"`
	}
	if json.Unmarshal(data, &resp) == nil && resp.Metadata.Status != "" && resp.Metadata.Status != "Success" {
		logger.Error("Incus 操作原始响应: %s", string(data))
		return fmt.Errorf("Incus 操作失败 (status=%s, code=%d): %s",
			resp.Metadata.Status, resp.Metadata.StatusCode, resp.Metadata.Err)
	}
	return nil
}

func (c *Client) put(ctx context.Context, apiPath string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, "PUT", apiPath, body)
}

func (c *Client) patch(ctx context.Context, apiPath string, body interface{}) error {
	_, err := c.doRequest(ctx, "PATCH", apiPath, body)
	return err
}

func (c *Client) post(ctx context.Context, apiPath string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, "POST", apiPath, body)
}

func (c *Client) deleteReq(ctx context.Context, apiPath string) ([]byte, error) {
	return c.doRequest(ctx, "DELETE", apiPath, nil)
}

func (c *Client) Get(ctx context.Context, apiPath string, result interface{}) error {
	return c.get(ctx, apiPath, result)
}

// GetRaw 返回原始响应字节，供 gjson 按需提取，避免完整反序列化。
func (c *Client) GetRaw(ctx context.Context, apiPath string) ([]byte, error) {
	return c.doRequest(ctx, "GET", apiPath, nil)
}

func (c *Client) Patch(ctx context.Context, apiPath string, body interface{}) error {
	return c.patch(ctx, apiPath, body)
}

func (c *Client) execIncus(ctx context.Context, args ...string) (string, error) {
	if c.execPath == "" {
		return "", fmt.Errorf("incus CLI 未找到")
	}
	cmd := exec.CommandContext(ctx, c.execPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	logger.Info("执行Incus命令: %s %s", c.execPath, strings.Join(args, " "))

	err := cmd.Run()
	if err != nil {
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = err.Error()
		}
		logger.Error("Incus命令执行失败: %s", errMsg)
		return "", fmt.Errorf("%s", errMsg)
	}

	return stdout.String(), nil
}
