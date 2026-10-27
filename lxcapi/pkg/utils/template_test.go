package utils

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/admin/dashboard", nil)
	return c
}

func TestMergeTemplateData_InjectsVersion(t *testing.T) {
	data := MergeTemplateData(newTestContext(), gin.H{})
	if data["Version"] == "" {
		t.Error("Version 应始终被注入")
	}
}

func TestMergeTemplateData_KeepsPresetData(t *testing.T) {
	data := MergeTemplateData(newTestContext(), gin.H{"username": "admin"})
	if data["username"] != "admin" {
		t.Errorf("预设数据不应被改写, got %v", data["username"])
	}
}
