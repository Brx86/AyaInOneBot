// Package bootstrap 集中注册并守护所有 bot。
// 每个 bot 由同目录的 bot_<名>.go 通过 init() 注册，文件带 //go:build !no<名>
// 构建标签（Go 的标签粒度是文件，条件导入需一 bot 一文件），
// 构建时传 -tags=no<名> 即可把该 bot 从二进制中排除。
package bootstrap

import (
	"log"
	"time"
)

// Bot 描述一个待启动的 bot。
type Bot struct {
	Name  string
	Start func()
}

var bots []Bot

// Register 由各 bot_*.go 的 init() 调用。
func Register(name string, start func()) {
	bots = append(bots, Bot{name, start})
}

// Bots 返回已注册的 bot 列表。
func Bots() []Bot { return bots }

// Run 为所有 bot 启动守护协程：panic 或退出后 5s 重启，互不影响。
func Run() {
	for _, b := range bots {
		go supervise(b)
	}
}

func supervise(b Bot) {
	for {
		runBot(b)
		log.Printf(" MAIN│ %s 退出，5s 后重启", b.Name)
		time.Sleep(5 * time.Second)
	}
}

func runBot(b Bot) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf(" MAIN│ %s panic: %v", b.Name, r)
		}
	}()
	b.Start()
}
