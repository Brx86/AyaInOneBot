//go:build !nolex

package bootstrap

import "ayainonebot/ayalex"

func init() {
	Register("LEX", ayalex.Start)
}
