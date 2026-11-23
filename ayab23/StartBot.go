package ayab23

import (
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	tele "gopkg.in/telebot.v4"

	"ayainonebot/internal/config"
	"ayainonebot/internal/tgutil"
)

var patterns = []*regexp.Regexp{
	regexp.MustCompile(`https://www.bilibili.com/video/(BV1\w{9})`), // 匹配 BV 号
	regexp.MustCompile(`(BV1\w{9})`),                                // 匹配 BV 号
	regexp.MustCompile(`b23.tv/(\w+)`),                              // 匹配短链接
	regexp.MustCompile(`(【.+?】)`),                                   // 匹配标题
	regexp.MustCompile(`xiaoheihe.cn.+link(_id=|/)(\w+)`),           // 匹配小黑盒
	regexp.MustCompile(`mp.weixin.qq.com/(.+)`),                     // 匹配公众号文章
	regexp.MustCompile(`(:?http[s]*://xhslink\.c\w+/o/\w+)|(:?https://www\.xiaohongshu\.com/explore/.+?xsec_token=[^& ]+?)[$|&|\s]`), // 匹配小红书链接
}

func getLocation(url string) string {
	c := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := c.Head(url)
	if err != nil {
		return ""
	}
	return resp.Header.Get("Location")
}

func handleUrl(text string) string {
	var header string
	var match []string
	if strings.HasPrefix(text, "+") { // +加标题
		header = patterns[3].FindString(text) + "\n"
	}
	for _, p := range patterns[:2] { // 匹配 BV 号
		match = p.FindStringSubmatch(text)
		if len(match) > 1 {
			return header + `https://www.bilibili.com/video/` + match[1]
		}
	}
	match = patterns[2].FindStringSubmatch(text) // 匹配b23短链接
	if len(match) > 1 {
		text = getLocation(`https://b23.tv/` + match[1])
		return header + handleUrl(text)
	}
	match = patterns[4].FindStringSubmatch(text) // 匹配小黑盒
	if len(match) > 2 {
		fx_url := `https://fxhey.aya1.de/` + match[2]
		hb_url := `https://www.xiaoheihe.cn/app/bbs/link/` + match[2]
		return `[` + hb_url + `](` + fx_url + `)`
	}
	match = patterns[5].FindStringSubmatch(text) // 匹配微信
	if len(match) > 1 {
		fx_url := `https://fxwx.aya1.de/` + match[1]
		wx_url := `https://mp.weixin.qq.com/s/` + match[1]
		if len(wx_url) > 52 {
			return `[` + wx_url[:52] + `...](` + fx_url + `)`
		}
		return `[` + wx_url + `](` + fx_url + `)`
	}
	match = patterns[6].FindStringSubmatch(text) // 匹配小红书
	if len(match) == 3 {
		return FetchAndUpload(match[0])
	}
	return "未找到符合URL"
}

func Start() {
	telegraphToken = config.Must("TELEGRAPH_TOKEN")
	bot, err := tele.NewBot(tele.Settings{
		Token:  config.Must("B23_BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 30 * time.Second},
	})
	if err != nil {
		panic(" B23│ " + err.Error())
	}
	log.Println(" B23│", "Connected:", bot.Me.FirstName)
	bot.Handle(tele.OnText, func(c tele.Context) error {
		text := c.Text()
		log.Println(" B23│", tgutil.UserName(c.Sender())+":", text)
		return c.Reply(handleUrl(text), tele.ModeMarkdown)
	})
	bot.Handle(tele.OnQuery, func(c tele.Context) error {
		qText := c.Query().Text
		if qText == "" {
			return nil
		}
		log.Println(" B23│", tgutil.UserName(c.Sender())+"@:", qText)
		// 创建 inline 列表
		url := handleUrl(qText)
		title := url
		if strings.Contains(title, "]") {
			title = strings.Split(title, "]")[0][1:]
		}
		results := tele.Results{&tele.ArticleResult{Title: title, Text: url}}
		results[0].SetParseMode(tele.ModeMarkdown)
		// 回复 inline 消息
		return c.Answer(&tele.QueryResponse{
			Results:   results,
			CacheTime: 5,
		})
	})
	// 启动 bot
	bot.Start()
}
