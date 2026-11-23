// Package tgutil 收敛各 bot 重复使用的 Telegram 相关小工具。
package tgutil

import (
	"regexp"
	"strings"

	tele "gopkg.in/telebot.v4"
)

// UserName 返回用户显示名：优先用户名，否则 FirstName。
func UserName(u *tele.User) string {
	if u == nil {
		return "unknown"
	}
	if u.Username != "" {
		return u.Username
	}
	return u.FirstName
}

// EscapeHTML 转义 Telegram HTML 模式下必须转义的字符。
func EscapeHTML(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// Truncate 按 rune 截断文本，避免把多字节字符截成非法 UTF-8。
// max <= 0 时不截断。
func Truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "......"
}

// reMDLink 匹配 markdown 链接 [文本](http(s)://url)。
var reMDLink = regexp.MustCompile(`\[([^\[\]]+)\]\((https?://[^)\s]+)\)`)

// MarkdownLinksToHTML 把 markdown 链接转成 HTML <a> 链接。
// Telegram 的 HTML 模式不渲染 markdown 链接，需显式转换。
func MarkdownLinksToHTML(s string) string {
	return reMDLink.ReplaceAllString(s, `<a href="$2">$1</a>`)
}

// FoldQuote 转义并按 rune 截断后，用可展开的 blockquote 包裹。
func FoldQuote(s string, max int) string {
	return "<blockquote expandable>" + MarkdownLinksToHTML(EscapeHTML(Truncate(s, max))) + "</blockquote>"
}
