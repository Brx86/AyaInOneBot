package ayaomo

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	tele "gopkg.in/telebot.v4"

	"ayainonebot/internal/config"
	"ayainonebot/internal/tgutil"
)

var (
	baseUrl = `https://www.omofun.link/`
	p1      = regexp.MustCompile(`href="(/vod/play/id/\d+?/sid/\d/nid/\d+?\.html)"`)
	p2      = regexp.MustCompile(`"url":"(https:.+?\.(m3u8|mp4).*?)"`)
	client  = resty.New().
		SetHeader("referer", baseUrl).
		SetHeader("User-Agent", "Mozilla/5.0 Firefox/141.0")
)

func detail_to_play(detailUrl string) (playUrl string, err error) {
	resp, err := client.R().Get(detailUrl)
	body := resp.String()
	match := p1.FindStringSubmatch(body)
	if len(match) > 0 {
		playUrl = baseUrl + match[1]
	}
	return
}

func play_to_m3u8(playUrl string) (m3u8Url string, err error) {
	resp, err := client.R().Get(playUrl)
	body := resp.String()
	match := p2.FindStringSubmatch(body)
	if len(match) > 0 {
		m3u8Url = "`" + strings.Replace(match[1], `\/`, `/`, -1) + "`"
	}
	return
}

func handleUrl(url string) (m3u8Url string) {
	var err error
	if strings.Contains(url, "/vod/detail/id/") {
		url, err = detail_to_play(url)
	}
	if strings.Contains(url, "/vod/play/id/") && err == nil {
		m3u8Url, err = play_to_m3u8(url)
	}
	if err != nil {
		m3u8Url = "未发现链接！"
	}
	return
}

func Start() {
	apiKey = config.Must("QWEATHER_API_KEY")
	bot, err := tele.NewBot(tele.Settings{
		Token:  config.Must("OMO_BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 30 * time.Second},
	})
	if err != nil {
		panic(" OMO│ " + err.Error())
	}
	log.Println(" OMO│", "Connected:", bot.Me.FirstName)
	bot.Handle(tele.OnText, func(c tele.Context) error {
		text := c.Text()
		log.Println(" OMO│", tgutil.UserName(c.Sender())+":", text)
		if text == "/start" {
			return c.Reply("OMO!")
		}
		if text == "/t" {
			return c.Send("`"+getTem()+"`", tele.ModeMarkdownV2)
		}
		return c.Reply(handleUrl(text), tele.ModeMarkdownV2)
	})
	// 启动 bot
	bot.Start()
}
