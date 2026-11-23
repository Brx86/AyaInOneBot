//go:build !nogun

package bootstrap

import "ayainonebot/ayagun"

func init() {
	Register("GUN", ayagun.Start)
}
