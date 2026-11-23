package ayazssm

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	llm "github.com/sashabaranov/go-openai"
)

const (
	GEMINI_BASE  = "https://generativelanguage.googleapis.com/v1beta/openai/"
	GEMINI_MODEL = "gemini-flash-latest"
)

// imageMIMEs 是当作图片处理的文件 MIME 类型。
var imageMIMEs = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// MemePrompt 是发给 Gemini 的系统提示词，用于解释迷因和梗。
const MemePrompt = `你是一个互联网迷因（meme）解密助手。用户会向你发送看不懂的图片或文字，请你用中文通俗易懂地解释其中的内容：

1. 如果是图片：解释这是什么梗或迷因、来源和背景。
2. 如果是纯文字：解释这句话的含义、出自哪里、什么梗，以及为什么会好笑或值得关注。

要求：
- 保持解释简洁有趣，像给朋友讲解一样，再简短一点。
- 如果内容涉及特定圈子（如 ACG、游戏、时事、网络文化等），请简要说明圈内语境。
- 不要输出 Markdown 格式符号（不要用 **、*、# 等标记），用纯文本自然表达。`

// NewMemeClient 创建指向 Gemini OpenAI 兼容端点的客户端。
func NewMemeClient(apiKey string) *llm.Client {
	config := llm.DefaultConfig(apiKey)
	config.BaseURL = GEMINI_BASE
	return llm.NewClientWithConfig(config)
}

// ExplainText 用 Gemini 解释一段文字。
func ExplainText(client *llm.Client, text string) (string, error) {
	resp, err := client.CreateChatCompletion(context.Background(), llm.ChatCompletionRequest{
		Model: GEMINI_MODEL,
		Messages: []llm.ChatCompletionMessage{
			{Role: "system", Content: MemePrompt},
			{Role: "user", Content: text},
		},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("empty response from Gemini (可能被安全过滤)")
	}
	return resp.Choices[0].Message.Content, nil
}

// ExplainImage 用 Gemini 解释一张图片，可附带图片的文字说明（caption）。
func ExplainImage(client *llm.Client, imageData []byte, caption string) (string, error) {
	promptText := "请解释这张图片的内容和梗。"
	if caption != "" {
		promptText = "请解释这张图片的内容和梗。图片附带的文字：" + caption
	}

	base64Str := base64.StdEncoding.EncodeToString(imageData)
	mimeType := http.DetectContentType(imageData)
	imageURL := "data:" + mimeType + ";base64," + base64Str

	resp, err := client.CreateChatCompletion(context.Background(), llm.ChatCompletionRequest{
		Model: GEMINI_MODEL,
		Messages: []llm.ChatCompletionMessage{
			{Role: "system", Content: MemePrompt},
			{
				Role: "user",
				MultiContent: []llm.ChatMessagePart{
					{Type: llm.ChatMessagePartTypeText, Text: promptText},
					{
						Type: llm.ChatMessagePartTypeImageURL,
						ImageURL: &llm.ChatMessageImageURL{
							URL: imageURL,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("empty response from Gemini (可能被安全过滤)")
	}
	return resp.Choices[0].Message.Content, nil
}

// isImageDoc 判断 Document 是否为支持的图片格式。
func isImageDoc(mime string) bool {
	return imageMIMEs[mime]
}
