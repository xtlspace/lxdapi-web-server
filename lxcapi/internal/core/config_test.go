package core

import "testing"

func TestAdminPath(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", "/admin"},
		{"   ", "/admin"},
		{"/", "/admin"},
		{"admin", "/admin"},
		{"/admin", "/admin"},
		{"/test", "/test"},
		{"/test/", "/test"},
		{"/test///", "/test"},
		{"  /a/b/  ", "/a/b"},
		{"test/admin", "/test/admin"},
	}

	for _, c := range cases {
		cfg := AdminConfig{Path: c.in}
		if got := cfg.AdminPath(); got != c.want {
			t.Errorf("AdminConfig{Path: %q}.AdminPath() = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestAdminBasePath(t *testing.T) {
	old := GlobalConfig
	defer func() { GlobalConfig = old }()

	GlobalConfig = nil
	if got := AdminBasePath(); got != "/admin" {
		t.Errorf("配置未加载时 AdminBasePath() = %q, want %q", got, "/admin")
	}

	GlobalConfig = &Config{Admin: AdminConfig{Path: "/panel"}}
	if got := AdminBasePath(); got != "/panel" {
		t.Errorf("AdminBasePath() = %q, want %q", got, "/panel")
	}
}
