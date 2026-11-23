//go:build secrets

package config

import "testing"

// TestEmbeddedLoaded 在 -tags=secrets 构建下确认密钥已生成内嵌。
// 若失败，先运行 go run ./tools/gensecrets。
func TestEmbeddedLoaded(t *testing.T) {
	if len(embedded) == 0 {
		t.Fatal("构建标签 secrets 下 embedded 为空：先运行 go run ./tools/gensecrets")
	}
	if v := Must("GEMINI_API_KEY"); v == "" {
		t.Error("Must(GEMINI_API_KEY) 不应返回空")
	}
}
