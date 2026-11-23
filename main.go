package main

import (
	"log"

	"ayainonebot/internal/bootstrap"
	"ayainonebot/internal/config"
)

func main() {
	config.Load()
	log.Printf(" MAIN│ 共注册 %d 个 bot", len(bootstrap.Bots()))
	bootstrap.Run()
	select {}
}
