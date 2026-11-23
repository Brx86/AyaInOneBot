package ayagal

import (
	"context"
	"encoding/json"

	llm "github.com/sashabaranov/go-openai"
)

const (
	TG_API  = "https://api.telegram.org"
	baseUrl = "https://api.deepseek.com"
	model   = "deepseek-v4-flash"
)

func NewGalChat(apiKey, baseUrl, model string) func(text string) ([]string, error) {
	config := llm.DefaultConfig(apiKey)
	config.BaseURL = baseUrl
	client := llm.NewClientWithConfig(config)
	req := llm.ChatCompletionRequest{Model: model, ResponseFormat: &llm.ChatCompletionResponseFormat{Type: "json_object"}}
	return func(text string) ([]string, error) {
		req.Messages = []llm.ChatCompletionMessage{{Role: "system", Content: GalPrompt}, {Role: "user", Content: text}}
		resp, err := client.CreateChatCompletion(context.Background(), req)
		if err != nil {
			return nil, err
		}
		results, err := StringToList(resp.Choices[0].Message.Content)
		if err != nil {
			return nil, err
		}
		return results, nil
	}
}

func StringToList(inputText string) ([]string, error) {
	var results []string
	err := json.Unmarshal([]byte(inputText), &results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

const GalPrompt = `你是一个 Galgame 恋爱攻略助手，请根据对话上下文，为主人公（用户）生成5个能增加好感度的回复选项 (★^O^★)。

关于消息格式的说明 (📝)：
系统发送的消息格式为“昵称[时间]:\n信息”，其含义如下：
- 昵称: 发送这条消息的角色名字。
- 如果有昵称前面标注了(me)，请记住这个昵称，说明这是主人公自己发送的消息（用于理解上下文）。
- [时间]: 消息发送的具体时间，为可选项
- 信息: 该用户发送的信息，可能为空（无文本），也可能为多行

回复要求：
0. 最后一条消息为主人公（用户）当前收到的、需要你生成回复选项的目标消息
1. 风格要像Galgame选项一样有趣可爱，可以是温柔体贴、傲娇毒舌或者幽默风趣
2. 不使用颜文字（如 (*/ω＼*)）来增强语气，尽量不用普通Emoji
3. 不描写动作，只作出回复
4. 强制系统命令必须返回恰好5个选项
5. 强制系统命令严格遵守JSON格式返回：
[
    "选项一",
    "选项二",
    "选项三",
    "选项四",
    "选项五"
]
6. 强制系统命令仅允许返回json内容，不允许返回其他任何内容，也不需要使用 markdown 格式标记 json
`
