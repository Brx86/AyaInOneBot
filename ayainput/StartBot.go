package ayainput

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	tele "gopkg.in/telebot.v4"

	"ayainonebot/internal/config"
	"ayainonebot/internal/tgutil"
)

const (
	SchemeQuanpin = "zh-t-i0-pinyin"
	SchemeFlypy   = "zh-t-i0-pinyin-x0-shuangpin-flypy"
	SchemeZiranma = "zh-t-i0-pinyin-x0-shuangpin-ziranma"
)

var userSettings sync.Map

func Start() {
	bot, err := tele.NewBot(tele.Settings{
		Token:  config.Must("INPUT_BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 30 * time.Second},
	})
	if err != nil {
		panic(" URF│ " + err.Error())
	}
	log.Println(" URF│", "Connected:", bot.Me.FirstName)

	// --- 1. 定义按钮菜单 ---
	menu := &tele.ReplyMarkup{}
	// Data 参数： 按钮文字, 唯一ID, 自定义负载数据(即方案名)
	btnQuanpin := menu.Data("全拼", "set_scheme", SchemeQuanpin)
	btnFlypy := menu.Data("小鹤双拼", "set_scheme", SchemeFlypy)
	btnZiranma := menu.Data("自然码", "set_scheme", SchemeZiranma)
	bot.Handle(&btnQuanpin, changeSchemeHandler)
	bot.Handle(&btnFlypy, changeSchemeHandler)
	bot.Handle(&btnZiranma, changeSchemeHandler)

	// 排版：第一行全拼，第二行两个双拼
	menu.Inline(
		menu.Row(btnQuanpin),
		menu.Row(btnFlypy),
		menu.Row(btnZiranma),
	)
	// --- 2. 处理 /start 命令 ---
	bot.Handle("/start", func(c tele.Context) error {
		return c.Send("请选择您的输入方案：", menu)
	})
	// 处理 inline 搜索
	bot.Handle(tele.OnQuery, func(c tele.Context) error {
		u := c.Sender()
		inputText := c.Query().Text
		if inputText == "" {
			return nil
		}
		log.Println(" URF│", tgutil.UserName(u), "input:", inputText)
		// 搜索结果
		method := SchemeQuanpin
		// 从 map 中查找用户设置
		if val, ok := userSettings.Load(u.ID); ok {
			method = val.(string)
		}
		inputResults, _, err := GoogleInput(method, inputText)
		if err != nil {
			if !errors.Is(err, ErrUnexpectedJSON) {
				log.Println(" URF│", "Error:", err.Error())
			}
			return nil
		}
		// 创建 inline 列表
		results := make(tele.Results, len(inputResults))
		for i, text := range inputResults {
			results[i] = &tele.ArticleResult{
				Title: text,
				Text:  text,
			}
			results[i].SetResultID(fmt.Sprint(i))
		}
		log.Println(" URF│", inputResults[0])
		// 回复 inline 消息
		return c.Answer(&tele.QueryResponse{
			Results:   results,
			CacheTime: 10,
		})
	})
	// 启动 bot
	bot.Start()
}

// changeSchemeHandler 处理按钮点击逻辑，c.Data() 为按钮携带的方案名。
func changeSchemeHandler(c tele.Context) error {
	scheme := c.Data()
	u := c.Sender()
	userSettings.Store(u.ID, scheme)

	// 获取方案的可读名称用于回复
	schemeName := "未知"
	switch scheme {
	case SchemeQuanpin:
		schemeName = "全拼"
	case SchemeFlypy:
		schemeName = "小鹤双拼"
	case SchemeZiranma:
		schemeName = "自然码"
	}

	log.Printf(" URF│ User %s set scheme to: %s", tgutil.UserName(u), schemeName)

	// 应答 callback 并提示设置成功
	return c.RespondText("设置成功！当前方案：" + schemeName)
}
