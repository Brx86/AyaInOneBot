//go:build !nogal

package bootstrap

import "ayainonebot/ayagal"

func init() {
	Register("GAL", ayagal.Start)
}
