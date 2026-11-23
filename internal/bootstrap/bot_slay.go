//go:build !noslay

package bootstrap

import "ayainonebot/ayaslay"

func init() {
	Register("SLAY", ayaslay.Start)
}
