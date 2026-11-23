package ayamorse

import "strings"

// normalizeChinesePunctuation replaces common Chinese punctuation with ASCII equivalents.
func normalizeChinesePunctuation(text string) string {
	replacer := strings.NewReplacer(
		"\uff0c", ",", // ，
		"\u3002", ".", // 。
		"\uff1f", "?", // ？
		"\uff01", "!", // ！
		"\uff1b", ";", // ；
		"\uff1a", ":", // ：
		"\u3001", ",", // 、
		"\uff08", "(", // （
		"\uff09", ")", // ）
		"\u3010", "[", // 【
		"\u3011", "]", // ）
		"\u201c", "\"", // "
		"\u201d", "\"", // "
		"\u2014", "-", // —
		"\u2026", "...", // …
	)
	return replacer.Replace(text)
}
