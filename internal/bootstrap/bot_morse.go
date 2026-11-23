//go:build !nomorse

package bootstrap

import "ayainonebot/ayamorse"

func init() {
	Register("MORSE", ayamorse.Start)
}
