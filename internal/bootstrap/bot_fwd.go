//go:build !nofwd

package bootstrap

import "ayainonebot/ayafwd"

func init() {
	Register("FWD", ayafwd.Start)
}
