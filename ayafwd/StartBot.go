package ayafwd

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
		Token:  config.Must("FWD_BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 30 * time.Second},
	})
	if err != nil {
		panic(" FWD│ " + err.Error())
	}
	log.Println(" FWD│", "Connected:", bot.Me.FirstName)
	summarizer := NewSummaryClient(config.Must("GEMINI_API_KEY"))

	// 中间件：拦截来自 SOURCE_CHAT_IDS 的新消息与编辑消息，转发并写入环形缓冲区
	bot.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			msg := c.Message()
			if msg == nil {
				msg = c.Update().EditedMessage
			}
			if msg != nil && slices.Contains(SOURCE_CHAT_IDS, c.Chat().ID) {
				text := msg.Text
				isCommand := strings.HasPrefix(text, "/tldr")
				buf := getBuffer(c.Chat().ID)
				sender := tgutil.UserName(msg.Sender)
				// 缓冲文本消息；命令本身不进缓冲，媒体消息记占位符
				switch {
				case text != "":
					if !isCommand {
						buf.Push(MessageEntry{
							Text:      text,
							Sender:    sender,
							MessageID: msg.ID,
							ChatID:    c.Chat().ID,
						})
					}
				default:
					if label := mediaLabel(msg); label != "" {
						buf.Push(MessageEntry{
							Text:      label,
							Sender:    sender,
							MessageID: msg.ID,
							ChatID:    c.Chat().ID,
						})
					}
				}
				// 不转发 /tldr 命令
				if !isCommand {
					if _, err := c.Bot().Forward(tele.ChatID(TARGET_CHAT_ID), msg); err != nil {
						log.Println(" FWD│", "Forward error:", err)
					}
				}
			}
			return next(c)
		}
	})

	// /tldr — 群聊总结（私发用户）
	bot.Handle("/tldr", func(c tele.Context) error {
		if !slices.Contains(SOURCE_CHAT_IDS, c.Chat().ID) {
			return nil
		}

		// 解析参数 N
		n := DEFAULT_N
		if payload := strings.TrimSpace(c.Message().Payload); payload != "" {
			if parsed, err := strconv.Atoi(payload); err == nil {
				n = parsed
			}
		}
		if n <= 0 {
			n = DEFAULT_N
		}
		if n > MAX_BUFFER {
			n = MAX_BUFFER
		}

		// 获取缓冲区消息
		buf := getBuffer(c.Chat().ID)
		msgs := buf.Last(n)
		if len(msgs) == 0 {
			replyAndAutoDelete(c, "还没有缓存的消息，无法总结。")
			return nil
		}

		c.Notify(tele.Typing)

		// 调用 Gemini 总结
		summary, err := SummarizeMessages(summarizer, msgs, c.Chat().ID)
		if err != nil {
			log.Println(" FWD│", "Summarize error:", err)
			replyAndAutoDelete(c, "总结失败："+err.Error())
			return nil
		}

		// 私发给用户
		userID := c.Sender().ID
		_, err = c.Bot().Send(tele.ChatID(userID), summary, tele.ModeHTML)
		if err != nil {
			log.Println(" FWD│", "Send DM error:", err)
			// 失败提示：群内回复，10 秒后自动删除
			replyAndAutoDelete(c, "私发失败，请先私聊我发送 /start。")
			return nil
		}
		log.Println("·FWD│", "Summary for @"+tgutil.UserName(c.Sender()))
		// 成功提示：群内回复，10 秒后自动删除
		replyAndAutoDelete(c, "已私发 ✓")
		return nil
	})

	// telebot.v4 没有 OnMessage 通配 endpoint，需要为每种类型注册空 handler 来激活处理链
	noop := func(c tele.Context) error { return nil }
	bot.Handle(tele.OnText, noop)
	bot.Handle(tele.OnPhoto, noop)
	bot.Handle(tele.OnSticker, noop)
	bot.Handle(tele.OnVideo, noop)
	bot.Handle(tele.OnDocument, noop)
	bot.Handle(tele.OnAudio, noop)
	bot.Handle(tele.OnVoice, noop)
	bot.Handle(tele.OnAnimation, noop)
	bot.Handle(tele.OnVideoNote, noop)
	bot.Handle(tele.OnContact, noop)
	bot.Handle(tele.OnLocation, noop)
	bot.Handle(tele.OnVenue, noop)
	bot.Handle(tele.OnPoll, noop)
	bot.Handle(tele.OnDice, noop)
	bot.Handle(tele.OnEdited, noop)

	bot.Start()
}

// mediaLabel 返回媒体消息在总结缓冲区中的占位文本。
func mediaLabel(m *tele.Message) string {
	switch {
	case m.Photo != nil:
		return "[图片]"
	case m.Animation != nil:
		return "[动图]"
	case m.Sticker != nil:
		return "[贴纸]"
	case m.Video != nil:
		return "[视频]"
	case m.VideoNote != nil:
		return "[视频留言]"
	case m.Voice != nil:
		return "[语音]"
	case m.Audio != nil:
		return "[音频]"
	case m.Document != nil:
		return "[文件]"
	case m.Location != nil:
		return "[位置]"
	case m.Contact != nil:
		return "[联系人]"
	case m.Poll != nil:
		return "[投票]"
	case m.Dice != nil:
		return "[骰子]"
	}
	return ""
}

// replyAndAutoDelete 发送群组回复并在指定时间后自动删除
func replyAndAutoDelete(c tele.Context, text string) {
	tip, err := c.Bot().Reply(c.Message(), text)
	if err != nil {
		log.Println(" FWD│", "Reply error:", err)
		return
	}

	// 启动协程异步等待并删除
	go func() {
		time.Sleep(10 * time.Second)
		if err := c.Bot().Delete(tip); err != nil {
			log.Println(" FWD│", "Delete tip error:", err)
		}
	}()
}
