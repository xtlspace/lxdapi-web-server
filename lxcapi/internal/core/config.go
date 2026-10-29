package core

import (
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

type Config struct {
	System   SystemConfig   `yaml:"system"`
	LXC      LXCConfig      `yaml:"lxc"`
	Network  NetworkConfig  `yaml:"network"`
	Traffic  TrafficConfig  `yaml:"traffic"`
	Task     TaskConfig     `yaml:"task"`
	Admin    AdminConfig    `yaml:"admin"`
	Plugins  PluginsConfig  `yaml:"plugins"`
}

type SystemConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Security SecurityConfig `yaml:"security"`
	Logger   LoggerConfig   `yaml:"logger"`
	Database DatabaseConfig `yaml:"database"`
}

type ServerConfig struct {
	Host string    `yaml:"host"`
	Port int       `yaml:"port"`
	TLS  TLSConfig `yaml:"tls"`
}

type TLSConfig struct {
	Enabled      bool   `yaml:"enabled"`
	CertFile     string `yaml:"cert_file"`
	KeyFile      string `yaml:"key_file"`
	AutoGenerate bool   `yaml:"auto_generate"`
}

type SecurityConfig struct {
	APIHash string `yaml:"api_hash"`
}

type LoggerConfig struct {
	Level    string `yaml:"level"`
	Colorful bool   `yaml:"colorful"`
}

// IsDebug 统一判定是否启用 debug 级别（忽略大小写与首尾空格）
func (c *LoggerConfig) IsDebug() bool {
	return strings.EqualFold(strings.TrimSpace(c.Level), "debug")
}

type LXCConfig struct {
	Socket  string `yaml:"socket"`
	Timeout int    `yaml:"timeout"`
	Storage string `yaml:"storage"`
	Bridge  string `yaml:"bridge"`
}

const defaultLXCSocket = "/var/lib/incus/unix.socket"

// SocketPath 返回 Incus socket 路径，未配置或为空白时回退到默认路径。
func (c *LXCConfig) SocketPath() string {
	if s := strings.TrimSpace(c.Socket); s != "" {
		return s
	}
	return defaultLXCSocket
}

// BridgeName 返回默认网桥名称，为空时回退到 incusbr0。
func (c *LXCConfig) BridgeName() string {
	if c.Bridge != "" {
		return c.Bridge
	}
	return "incusbr0"
}

type NetworkConfig struct {
	StartPostCommand string `yaml:"start_post_command"`
}

type TrafficConfig struct {
	Interval int `yaml:"interval"`
}

type TaskConfig struct {
	Workers         int `yaml:"workers"`
	QueueSize       int `yaml:"queue_size"`
	Timeout         int `yaml:"timeout"`
	AutoCleanupDays int `yaml:"auto_cleanup_days"`
}

type DatabaseConfig struct {
	Type     string `yaml:"type"`
	SQLite   SQLiteConfig   `yaml:"sqlite"`
	MySQL    MySQLConfig    `yaml:"mysql"`
	Postgres PostgresConfig `yaml:"postgres"`
}

type SQLiteConfig struct {
	Path string `yaml:"path"`
}

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"sslmode"`
}

type PluginsConfig struct {
	Nginx PluginConfig `yaml:"nginx"`
}

type PluginConfig struct {
	Enabled bool `yaml:"enabled"`
}

type AdminConfig struct {
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
	SessionSecret string `yaml:"session_secret"`
	Path          string `yaml:"path"`
}

// AdminPath 返回管理面板路径，未配置或配置为根路径时回退到 /admin。
// 返回值保证以 / 开头且不带结尾斜杠。
func (c *AdminConfig) AdminPath() string {
	p := strings.TrimSpace(c.Path)
	if p == "" || p == "/" {
		return "/admin"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	p = strings.TrimRight(p, "/")
	if p == "" {
		return "/admin"
	}
	return p
}

// AdminBasePath 返回当前生效的管理面板路径，配置未加载时回退到 /admin。
func AdminBasePath() string {
	if GlobalConfig == nil {
		return "/admin"
	}
	return GlobalConfig.Admin.AdminPath()
}

var GlobalConfig *Config

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	GlobalConfig = &Config{}
	return yaml.Unmarshal(data, GlobalConfig)
}

