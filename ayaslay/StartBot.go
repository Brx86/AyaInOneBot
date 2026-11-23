package ayaslay

import (
	tele "gopkg.in/telebot.v4"

	"ayainonebot/internal/config"
	"ayainonebot/internal/filebot"
)

func Start() {
	filebot.Start(filebot.Options{
		Tag:       "SLAY│",
		Token:     config.Must("SLAY_BOT_TOKEN"),
		Files:     Data.FileList,
		Cards:     Data.CardList, // 随机抽取只出卡片
		RandomN:   10,
		SearchN:   20,
		Cache:     120,
		StripDash: true, // "-" 前缀的搜索结果不带文件名说明
		Build: func(name string, hasCaption bool) tele.Result {
			caption := ""
			if hasCaption {
				caption = name
			}
			url := Data.BaseURL + name + ".jpg"
			return &tele.PhotoResult{
				Title:       name,
				URL:         url,
				ThumbURL:    url,
				Height:      40,
				Description: name,
				Caption:     caption,
			}
		},
	})
}
