package utils

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestContext(keys map[string]interface{}) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/container/dashboard", nil)
	for k, v := range keys {
		c.Set(k, v)
	}
	return c
}

func TestMergeTemplateData_InjectsVersion(t *testing.T) {
	data := MergeTemplateData(newTestContext(nil), gin.H{"title": "preset"})
	if data["Version"] == "" {
		t.Error("Version 应始终被注入")
	}
}

func TestMergeTemplateData_ContextKeysOverride(t *testing.T) {
	c := newTestContext(map[string]interface{}{
		"SystemName":     "测试系统",
		"BgImage":        "/bg.png",
		"BgOpacity":      75,
		"ContentOpacity": 85,
	})
	data := MergeTemplateData(c, gin.H{"title": "preset"})

	if data["SystemName"] != "测试系统" {
		t.Errorf("SystemName = %v, want 测试系统", data["SystemName"])
	}
	if data["BgImage"] != "/bg.png" {
		t.Errorf("BgImage = %v, want /bg.png", data["BgImage"])
	}
	if data["BgOpacity"] != 75 {
		t.Errorf("BgOpacity = %v, want 75", data["BgOpacity"])
	}
	if data["ContentOpacity"] != 85 {
		t.Errorf("ContentOpacity = %v, want 85", data["ContentOpacity"])
	}
}

func TestMergeTemplateData_PresetTitleAvoidsDB(t *testing.T) {
	data := MergeTemplateData(newTestContext(nil), gin.H{"title": "预设标题"})
	if data["title"] != "预设标题" {
		t.Errorf("已存在的 title 不应被覆盖, got %v", data["title"])
	}
}

func TestMergeTemplateData_MissingKeysNotInjected(t *testing.T) {
	data := MergeTemplateData(newTestContext(nil), gin.H{"title": "x"})
	for _, k := range []string{"SystemName", "BgImage", "ContainerNoticeOpacity"} {
		if _, ok := data[k]; ok {
			t.Errorf("上下文未设置时不应注入 %s", k)
		}
	}
}
