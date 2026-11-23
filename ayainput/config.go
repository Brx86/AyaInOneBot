package ayainput

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/tidwall/gjson"
)

// 输入: string (例如 "shurufaceshi")
// 输出: []string (候选词列表)
var (
	ErrUnexpectedJSON = errors.New("ErrUnexpectedJSON")
)

func GoogleInput(method, text string) ([]string, []int64, error) {
	// 1. 构建 URL 和查询参数
	baseURL := "https://inputtools.google.com/request"
	params := url.Values{}
	params.Add("text", text)
	params.Add("itc", method)
	params.Add("num", "10")

	// 生成完整 URL
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// 2. 发送 HTTP GET 请求
	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("bad status: %s", resp.Status)
	}
	// 3. 读取响应体（Body）中的所有数据，转换为字符串
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading response body failed: %s", err.Error())
	}
	responseText := string(bodyBytes)

	// 4. 解析 JSON
	candidatesRaw := gjson.Get(responseText, "1.0.1").Array()
	matchedLengthRaw := gjson.Get(responseText, "1.0.3.matched_length").Array()
	if len(candidatesRaw) == 0 || len(matchedLengthRaw) == 0 {
		return nil, nil, fmt.Errorf("%w: candidates is empty", ErrUnexpectedJSON)
	}
	candidates := make([]string, 0, len(candidatesRaw))
	for _, result := range candidatesRaw {
		candidates = append(candidates, result.String())
	}
	matchedLength := make([]int64, 0, len(matchedLengthRaw))
	for _, result := range matchedLengthRaw {
		matchedLength = append(matchedLength, result.Int())
	}
	return candidates, matchedLength, nil
}
