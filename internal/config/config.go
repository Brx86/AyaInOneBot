// Package config 提供运行时配置。查找顺序：
//  1. 系统环境变量（含启动时从 .env 加载的部分）
//  2. 构建时内嵌的值（tools/gensecrets 由 .env 生成 embedded_gen.go，
//     构建标签 -tags=secrets，随 garble -literals 混淆进二进制）
//
// 开发机用 .env 即可；服务器只需二进制，无需明文配置文件。
package config

import (
	"os"
	"strings"
)

// ParseFile 解析 .env 格式文件，返回键值对；空行、# 注释与无法解析的行跳过。
func ParseFile(path string) map[string]string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	values := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
			value = value[1 : len(value)-1]
		}
		if key == "" {
			continue
		}
		values[key] = value
	}
	return values
}

// Load 读取 .env（存在时）并写入进程环境变量。
// 系统环境变量优先，不被 .env 覆盖；文件不存在时静默跳过。
func Load() {
	for key, value := range ParseFile(".env") {
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}
}

// Must 返回配置值：先查环境变量，再查构建内嵌值；均未提供时 panic，
// 由 main 的 supervisor 捕获记录后重启该 bot。
func Must(key string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	if v := strings.TrimSpace(embedded[key]); v != "" {
		return v
	}
	panic("未设置配置项 " + key + "（环境变量 / .env / 构建内嵌均未提供）")
}
