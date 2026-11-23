//go:build !noinput

package bootstrap

import "ayainonebot/ayainput"

func init() {
	Register("INPUT", ayainput.Start)
}
