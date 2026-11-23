// Package filebot 是 ayagun / ayalex / ayaslay 共用的 inline 文件搜索引擎：
// 从内嵌的文件名列表中搜索或随机抽取，组装成 inline 结果返回。
package filebot

import (
	"log"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v4"

	"ayainonebot/internal/tgutil"
)

// Options 描述一个基于文件列表的 inline 搜索 bot。
type Options struct {
	Tag   string // 日志前缀，如 "GUN│"
	Token string

	Files []string // 搜索范围（不含扩展名的文件名列表）
	Cards []string // 随机抽取池，为空时用 Files（ayaslay 只随机抽卡片）

	RandomN   int  // 空查询时返回的随机条数
	SearchN   int  // 搜索结果上限，<= 0 不限制（Telegram 单次最多 50 条）
	Cache     int  // inline 结果缓存秒数
	StripDash bool // true 时 "-" 前缀表示结果不带说明（ayaslay），false 时按普通关键字搜索

	// Build 把文件名组装成一条 inline 结果；hasCaption 为 false 表示
	// 用户以 "-" 前缀要求不带说明。
	Build func(name string, hasCaption bool) tele.Result
}

// Search 返回文件名中包含 keyword 的子集。
func Search(files []string, keyword string) []string {
	return slices.Collect(func(yield func(string) bool) {
		for _, name := range files {
			if strings.Contains(name, keyword) {
				if !yield(name) {
					return
				}
			}
		}
	})
}

// Random 从池中随机取 n 个不重复文件名；n 超过池大小时返回全池乱序。
func Random(pool []string, n int) []string {
	perm := rand.Perm(len(pool))
	if n > len(pool) {
		n = len(pool)
	}
	out := make([]string, n)
	for i, p := range perm[:n] {
		out[i] = pool[p]
	}
	return out
}

// Start 启动一个基于 Options 的 inline 搜索 bot。
func Start(o Options) {
	bot, err := tele.NewBot(tele.Settings{
		Token:  o.Token,
		Poller: &tele.LongPoller{Timeout: 30 * time.Second},
	})
	if err != nil {
		panic(o.Tag + " " + err.Error())
	}
	log.Println(o.Tag, "Connected:", bot.Me.FirstName)

	bot.Handle(tele.OnQuery, func(c tele.Context) error {
		q := c.Query().Text
		log.Println(o.Tag, tgutil.UserName(c.Sender()), "search:", q)

		var files []string
		hasCaption := true
		switch {
		case q == "" || q == "+" || (o.StripDash && q == "-"):
			pool := o.Files
			if len(o.Cards) > 0 {
				pool = o.Cards
			}
			files = Random(pool, o.RandomN)
		case o.StripDash && strings.HasPrefix(q, "-"):
			hasCaption = false
			files = Search(o.Files, q[1:])
		default:
			files = Search(o.Files, q)
		}
		if o.SearchN > 0 && len(files) > o.SearchN {
			files = files[:o.SearchN]
		}

		results := make(tele.Results, len(files))
		for i, name := range files {
			results[i] = o.Build(name, hasCaption)
			results[i].SetResultID(strconv.Itoa(i))
		}
		return c.Answer(&tele.QueryResponse{
			Results:   results,
			CacheTime: o.Cache,
		})
	})

	bot.Start()
}
