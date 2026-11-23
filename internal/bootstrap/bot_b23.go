//go:build !nob23

package bootstrap

import "ayainonebot/ayab23"

func init() {
	Register("B23", ayab23.Start)
}
