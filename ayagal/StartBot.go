package ayagal

import (
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v4"

	"ayainonebot/internal/config"
	"ayainonebot/internal/tgutil"
)

func Start() {
	bot, err := tele.NewBot(tele.Settings{
		URL:    TG_API,
		Token:  config.Must("GAL_BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 30 * time.Second},
	})
	if err != nil {
		panic(" GAL│ " + err.Error())
	}
	log.Println(" GAL│", "Connected:", bot.Me.FirstName)
	// 新建 GalChat
	gal := NewGalChat(config.Must("DEEPSEEK_API_KEY"), baseUrl, model)
	// 处理 inline 搜索
	ALLOW_USERS := []int64{1587739157, 5006810550} // 文，羽
	bot.Handle(tele.OnQuery, func(c tele.Context) error {
		// 仅白名单用户可用
		u := c.Sender()
		if !slices.Contains(ALLOW_USERS, u.ID) {
			return nil
		}
		inputText := c.Query().Text
		// 查询结果
		var inputResults []string
		var err error
		if inputText == "" {
			return nil
		} else if strings.HasPrefix(inputText, "-") {
			inputResults, err = StringToList(inputText[1:])
		} else {
			inputResults, err = gal(inputText)
		}
		log.Println(" GAL│", tgutil.UserName(u), "Q:", inputText)
		if err != nil {
			log.Println(" GAL│", tgutil.UserName(u), "error:", err.Error())
			return nil
		}
		// 创建 inline 列表
		results := make(tele.Results, len(inputResults))
		for i, text := range inputResults {
			results[i] = &tele.ArticleResult{
				Title: text,
				Text:  text,
			}
			results[i].SetResultID(strconv.Itoa(i))
		}
		// 回复 inline 消息
		return c.Answer(&tele.QueryResponse{
			Results:   results,
			CacheTime: 10,
		})
	})
	// 启动 bot
	bot.Start()
}
