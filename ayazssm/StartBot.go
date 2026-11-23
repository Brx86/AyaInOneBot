package ayazssm

import (
	"io"
	"log"
	"strings"
	"time"

	tele "gopkg.in/telebot.v4"

	"ayainonebot/internal/config"
	"ayainonebot/internal/tgutil"
)

func Start() {
	bot, err := tele.NewBot(tele.Settings{
		Token:  config.Must("ZSSM_BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 30 * time.Second},
	})
	if err != nil {
		panic("ZSSM│ " + err.Error())
	}
	log.Println("ZSSM│", "Connected:", bot.Me.FirstName)

	client := NewMemeClient(config.Must("GEMINI_API_KEY"))

	// --- /start ---
	bot.Handle("/start", func(c tele.Context) error {
		return c.Reply(
			"👋 你好！把看不懂的梗图或文字发给我，我会帮你解释。\n支持私聊发送图片和文字。\n群组中回复某条消息并发送 /zssm 即可让我解释它。",
		)
	})

	// --- /zssm — 回复消息解释（群组/私聊均可，仅回复时触发） ---
	bot.Handle("/zssm", func(c tele.Context) error {
		if !c.Message().IsReply() {
			return nil
		}
		replied := c.Message().ReplyTo
		name := tgutil.UserName(c.Sender())

		c.Notify(tele.Typing)

		var result string
		var err error

		switch {
		case replied.Photo != nil:
			log.Println("ZSSM│", name, "/zssm P:", replied.Caption)
			imageData, fileErr := downloadFile(bot, &replied.Photo.File)
			if fileErr != nil {
				log.Println("ZSSM│", name, "download err:", fileErr)
				return c.Reply("抱歉，下载图片失败：" + fileErr.Error())
			}
			result, err = ExplainImage(client, imageData, replied.Caption)

		case replied.Document != nil && isImageDoc(replied.Document.MIME):
			log.Println("ZSSM│", name, "/zssm D:", replied.Document.FileName, replied.Caption)
			imageData, fileErr := downloadFile(bot, &replied.Document.File)
			if fileErr != nil {
				log.Println("ZSSM│", name, "download err:", fileErr)
				return c.Reply("抱歉，下载文件失败：" + fileErr.Error())
			}
			result, err = ExplainImage(client, imageData, replied.Caption)

		case replied.Text != "":
			log.Println("ZSSM│", name, "/zssm T:", replied.Text)
			result, err = ExplainText(client, replied.Text)

		default:
			return c.Reply("该消息无可解释的内容（暂不支持贴纸、音频等类型）。")
		}

		if err != nil {
			log.Println("ZSSM│", name, "/zssm error:", err)
			return c.Reply("抱歉，出错了：" + err.Error())
		}

		_, err = c.Bot().Reply(replied, tgutil.FoldQuote(result, 3800), &tele.SendOptions{ParseMode: tele.ModeHTML})
		return err
	})

	// --- 文字消息（仅私聊、非命令） ---
	bot.Handle(tele.OnText, func(c tele.Context) error {
		if c.Chat().Type != tele.ChatPrivate {
			return nil
		}
		if strings.HasPrefix(c.Text(), "/") {
			return nil
		}
		name := tgutil.UserName(c.Sender())
		text := c.Text()
		log.Println("ZSSM│", name, "T:", text)

		c.Notify(tele.Typing)
		result, err := ExplainText(client, text)
		if err != nil {
			log.Println("ZSSM│", name, "error:", err)
			return c.Reply("抱歉，出错了：" + err.Error())
		}
		return c.Reply(tgutil.Truncate(result, 4000))
	})

	// --- 图片消息（仅私聊） ---
	bot.Handle(tele.OnPhoto, func(c tele.Context) error {
		if c.Chat().Type != tele.ChatPrivate {
			return nil
		}
		name := tgutil.UserName(c.Sender())
		log.Println("ZSSM│", name, "P:", c.Message().Caption)

		c.Notify(tele.Typing)
		imageData, err := downloadFile(bot, &c.Message().Photo.File)
		if err != nil {
			log.Println("ZSSM│", name, "download err:", err)
			return c.Reply("抱歉，下载图片失败：" + err.Error())
		}
		result, err := ExplainImage(client, imageData, c.Message().Caption)
		if err != nil {
			log.Println("ZSSM│", name, "error:", err)
			return c.Reply("抱歉，出错了：" + err.Error())
		}
		return c.Reply(tgutil.Truncate(result, 4000))
	})

	// --- 图片文件消息（仅私聊，jpg/png/webp） ---
	bot.Handle(tele.OnDocument, func(c tele.Context) error {
		if c.Chat().Type != tele.ChatPrivate {
			return nil
		}
		doc := c.Message().Document
		if !isImageDoc(doc.MIME) {
			return nil
		}
		name := tgutil.UserName(c.Sender())
		log.Println("ZSSM│", name, "D:", doc.FileName)

		c.Notify(tele.Typing)
		imageData, err := downloadFile(bot, &doc.File)
		if err != nil {
			log.Println("ZSSM│", name, "download err:", err)
			return c.Reply("抱歉，下载文件失败：" + tgutil.Truncate(err.Error(), 70))
		}
		result, err := ExplainImage(client, imageData, c.Message().Caption)
		if err != nil {
			log.Println("ZSSM│", name, "error:", err)
			return c.Reply("抱歉，出错了：" + tgutil.Truncate(err.Error(), 70))
		}
		return c.Reply(tgutil.Truncate(result, 4000))
	})

	bot.Start()
}

// downloadFile 从 Telegram 服务器下载文件并返回全部字节。
func downloadFile(bot *tele.Bot, file *tele.File) ([]byte, error) {
	reader, err := bot.File(file)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}
