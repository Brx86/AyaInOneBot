//go:build !nozssm

package bootstrap

import "ayainonebot/ayazssm"

func init() {
	Register("ZSSM", ayazssm.Start)
}
