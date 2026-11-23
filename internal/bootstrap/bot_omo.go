//go:build !noomo

package bootstrap

import "ayainonebot/ayaomo"

func init() {
	Register("OMO", ayaomo.Start)
}
