package ayalex

import (
	tele "gopkg.in/telebot.v4"

	"ayainonebot/internal/config"
	"ayainonebot/internal/filebot"
)

func Start() {
	filebot.Start(filebot.Options{
		Tag:     " LEX│",
		Token:   config.Must("LEX_BOT_TOKEN"),
		Files:   Data.FileList,
		RandomN: 20,
		SearchN: 50, // Telegram 单次 inline 最多 50 条
		Cache:   5,
		Build: func(name string, _ bool) tele.Result {
			return &tele.VoiceResult{
				Title:   Data.Prefix + name,
				URL:     Data.BaseURL + name + ".mp3",
				Caption: Data.Prefix + name,
			}
		},
	})
}
