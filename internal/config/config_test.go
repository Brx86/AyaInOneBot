package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "" +
		"# 注释行\n" +
		"\n" +
		"PLAIN=value\n" +
		"export EXPORTED=v2\n" +
		"QUOTED=\"has # not comment\"\n" +
		"SQUOTED='v3'\n" +
		"SPACED = v4 \n" +
		"NO_EQUALS_INVALID\n" +
		"EMPTY=\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	got := ParseFile(path)
	want := map[string]string{
		"PLAIN":    "value",
		"EXPORTED": "v2",
		"QUOTED":   "has # not comment",
		"SQUOTED":  "v3",
		"SPACED":   "v4",
		"EMPTY":    "",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("ParseFile()[%q] = %q, want %q", k, got[k], v)
		}
	}
	if _, ok := got["NO_EQUALS_INVALID"]; ok {
		t.Error("无法解析的行不应出现在结果中")
	}
	if got := ParseFile(filepath.Join(dir, "不存在")); got != nil {
		t.Errorf("文件不存在时应返回 nil，得到 %v", got)
	}
}

func TestMustFallsBackToEmbeddedThenPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("缺失配置项时应 panic")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "环境变量 / .env / 构建内嵌均未提供") {
			t.Fatalf("panic 信息不符: %v", r)
		}
	}()
	Must("DEFINITELY_MISSING_KEY_XYZ")
}
