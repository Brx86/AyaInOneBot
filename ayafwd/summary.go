package ayafwd

import (
	"context"
	"fmt"
	"strings"
	"sync"

	llm "github.com/sashabaranov/go-openai"

	"ayainonebot/internal/tgutil"
)

const (
	GEMINI_BASE  = "https://generativelanguage.googleapis.com/v1beta/openai/"
	GEMINI_MODEL = "gemini-flash-latest"
	MAX_BUFFER   = 1000
	DEFAULT_N    = 200
)

// MessageEntry 记录单条群聊消息。
type MessageEntry struct {
	Text      string
	Sender    string
	MessageID int
	ChatID    int64
}

// RingBuffer 线程安全的环形缓冲区。
type RingBuffer struct {
	mu   sync.RWMutex
	buf  []MessageEntry
	head int
	size int
	max  int
}

func NewRingBuffer(max int) *RingBuffer {
	return &RingBuffer{buf: make([]MessageEntry, max), max: max}
}

func (rb *RingBuffer) Push(e MessageEntry) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.buf[rb.head] = e
	rb.head = (rb.head + 1) % rb.max
	if rb.size < rb.max {
		rb.size++
	}
}

// Last 返回最近 n 条消息（按时间升序，旧→新）。
func (rb *RingBuffer) Last(n int) []MessageEntry {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	if n > rb.size {
		n = rb.size
	}
	if n <= 0 {
		return nil
	}
	result := make([]MessageEntry, n)
	start := (rb.head - n + rb.max) % rb.max
	for i := 0; i < n; i++ {
		result[i] = rb.buf[(start+i)%rb.max]
	}
	return result
}

var (
	buffers   = make(map[int64]*RingBuffer)
	buffersMu sync.RWMutex
)

func getBuffer(chatID int64) *RingBuffer {
	buffersMu.Lock()
	defer buffersMu.Unlock()
	if b, ok := buffers[chatID]; ok {
		return b
	}
	b := NewRingBuffer(MAX_BUFFER)
	buffers[chatID] = b
	return b
}

// BuildMessageLink 构造 tg 消息链接。
// 群组 chat_id 格式为 -100XXXXXXXXXX，链接中需去掉 -100 前缀。
func BuildMessageLink(chatID int64, msgID int) string {
	s := fmt.Sprintf("%d", chatID)
	s = strings.TrimPrefix(s, "-100")
	return fmt.Sprintf("https://t.me/c/%s/%d", s, msgID)
}

// NewSummaryClient 创建 Gemini 客户端。
func NewSummaryClient(apiKey string) *llm.Client {
	cfg := llm.DefaultConfig(apiKey)
	cfg.BaseURL = GEMINI_BASE
	return llm.NewClientWithConfig(cfg)
}

const summaryPrompt = `你是一个群聊总结助手。根据以下近期的群聊消息，用纯文本中文生成一份简洁的总结，而不是 markdown。

要求：
1. 总结讨论的核心话题（按话题分组，每个话题一句话）
2. 不要逐条复述，而是提炼
3. 如果消息中有链接、决定、待办事项等重要信息，单独列出
4. 每条消息末尾的 [原文] 是可点击的原文引用，总结中引用具体消息时使用 markdown 链接指向原文，例如 "某人[提到](原文链接)..."
5. 控制在 500 字以内`

// SummarizeMessages 调用 Gemini 总结消息列表，返回中文总结文本。
func SummarizeMessages(client *llm.Client, messages []MessageEntry, chatID int64) (string, error) {
	var sb strings.Builder
	for i, m := range messages {
		link := BuildMessageLink(chatID, m.MessageID)
		sb.WriteString(fmt.Sprintf("[%d] %s: %s [原文](%s)\n", i+1, m.Sender, m.Text, link))
	}

	resp, err := client.CreateChatCompletion(context.Background(), llm.ChatCompletionRequest{
		Model: GEMINI_MODEL,
		Messages: []llm.ChatCompletionMessage{
			{Role: "system", Content: summaryPrompt},
			{Role: "user", Content: sb.String()},
		},
	})
	if err != nil {
		return "", fmt.Errorf("Gemini 调用失败: %w", err)
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("Gemini 返回空响应（可能被安全过滤）")
	}
	return tgutil.FoldQuote(resp.Choices[0].Message.Content, 0), nil
}
