package system

import (
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"lxdapi/internal/core"
	"lxdapi/pkg/version"
)

type SystemInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Description  string `json:"description"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	LXDVersion   string `json:"lxd_version,omitempty"`
	Distribution string `json:"distribution,omitempty"`
	Kernel       string `json:"kernel,omitempty"`
}

func GetSystemInfo() *SystemInfo {
	info := &SystemInfo{
		Name:        "lxdapi",
		Version:     version.Version,
		Description: "Incus容器管理后端API服务",
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
	}

	if lxdVersion := getLXDVersion(); lxdVersion != "" {
		info.LXDVersion = lxdVersion
	}

	if kernel := getKernelVersion(); kernel != "" {
		info.Kernel = kernel
	}

	if distro := getDistribution(); distro != "" {
		info.Distribution = distro
	}

	return info
}

func getLXDVersion() string {
	socket := core.GlobalConfig.LXC.Socket
	if socket == "" {
		socket = "/var/lib/incus/unix.socket"
	}

	client := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.DialTimeout("unix", socket, 3*time.Second)
			},
		},
	}

	resp, err := client.Get("http://localhost/1.0")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	buf := make([]byte, 4096)
	n, _ := resp.Body.Read(buf)
	body := string(buf[:n])

	for _, part := range strings.Split(body, "\n") {
		if strings.Contains(part, "version") && strings.Contains(part, ":") {
			idx := strings.Index(part, ":")
			if idx > 0 {
				v := strings.TrimSpace(part[idx+1:])
				v = strings.Trim(v, "\"")
				return v
			}
		}
	}

	return ""
}

func getKernelVersion() string {
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func getDistribution() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			name := strings.TrimPrefix(line, "PRETTY_NAME=")
			name = strings.Trim(name, "\"")
			return name
		}
	}

	return ""
}
