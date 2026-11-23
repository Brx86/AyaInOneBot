package ayamorse

import (
	"bytes"
	"log"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v4"

	"ayainonebot/internal/config"
	"ayainonebot/internal/tgutil"
)

func Start() {
	bot, err := tele.NewBot(tele.Settings{
		Token:  config.Must("MORSE_BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 30 * time.Second},
	})
	if err != nil {
		panic("MORSE│ " + err.Error())
	}
	log.Println("MORSE│", "Connected:", bot.Me.FirstName)

	// /start
	bot.Handle("/start", func(c tele.Context) error {
		return c.Reply("Morse code bot. Send text to convert.\nPrefix =N (5-30) for audio speed, e.g. =20 Hello")
	})

	// Text messages: convert and send audio
	bot.Handle(tele.OnText, func(c tele.Context) error {
		text := c.Text()
		log.Println("MORSE│", tgutil.UserName(c.Sender())+":", text)

		wpm, cleanText := parseWPM(text)
		if cleanText == "" {
			return c.Reply("Please send some text to convert.")
		}
		cleanText = normalizeChinesePunctuation(cleanText)

		morseText := textToMorse(cleanText)
		if morseText == "" {
			return c.Reply("No recognizable characters in your message.")
		}

		wav := generateWAV(cleanText, wpm)

		// Send audio
		sizeStr := strconv.Itoa(len(wav)/1024) + "KB"
		c.Notify(tele.RecordingAudio)

		duration := len(wav) / SampleRate // rough estimate in seconds
		_, err := c.Bot().Send(c.Chat(), &tele.Voice{
			File:     tele.FromReader(bytes.NewReader(wav)),
			Duration: duration,
			Caption:  "Morse @ " + strconv.Itoa(wpm) + "wpm | " + sizeStr,
		})
		if err != nil {
			log.Println("MORSE│ Voice send error:", err)
			// Fallback: send text
			return c.Reply(morseText)
		}

		// Also send the text version
		return c.Reply(morseText)
	})

	// Inline query
	bot.Handle(tele.OnQuery, func(c tele.Context) error {
		qText := c.Query().Text
		if qText == "" {
			return nil
		}
		log.Println("MORSE│", tgutil.UserName(c.Sender())+"@:", qText)

		wpm, cleanText := parseWPM(qText)
		if cleanText == "" {
			return nil
		}
		cleanText = normalizeChinesePunctuation(cleanText)

		morseText := textToMorse(cleanText)
		if morseText == "" {
			return nil
		}

		// Telegram inline VoiceResult requires a URL; since we generate
		// WAV on-the-fly we can only return text results here.
		desc := morseText
		if wpm != 20 {
			desc = strconv.Itoa(wpm) + " wpm\n" + morseText
		}

		results := tele.Results{
			&tele.ArticleResult{
				Title:       "Morse Code",
				Text:        morseText,
				Description: desc,
			},
		}
		results[0].SetResultID("morse")

		return c.Answer(&tele.QueryResponse{
			Results:   results,
			CacheTime: 10,
		})
	})

	bot.Start()
}
