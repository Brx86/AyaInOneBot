//go:build !secrets

package config

// embedded 的默认空实现。真实值由 tools/gensecrets 从 .env 生成到
// internal/config/embedded_gen.go（构建标签 secrets），随 garble 混淆进二进制。
var embedded = map[string]string{}
