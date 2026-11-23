package ayab23

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/StarkBotsIndustries/telegraph/v2"
	"github.com/go-resty/resty/v2"
)

// telegraphToken 由 Start() 从环境变量注入，用于创建 Telegraph 页面。
var telegraphToken string

const (
	xhsPCUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
	xhsPCAccept    = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"
	// 手机 UA 配 Chrome 的 accept 会被跳到验证码页
	xhsMobileUserAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 18_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.5 Mobile/15E148 Safari/604.1"
	xhsMobileAccept    = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
)

var (
	client      = resty.New()
	reUndefined = regexp.MustCompile(`\bundefined\b`)
	reNewSet    = regexp.MustCompile(`new Set\((\[[^\]]*\])\)`)
	reTraceID   = regexp.MustCompile(`/[a-f0-9]{32}/(.+)`)
	reHashtag   = regexp.MustCompile(` ?#[^#\[\]]+\[话题\]# ?`)
)

type tNode telegraph.NodeElement
type tNodes []telegraph.Node
type tAttrs map[string]string
type Page struct {
	Url    string
	Title  string
	Text   string
	Images []string
	Nodes  []telegraph.Node
}

func (p *Page) genNodes() {
	texts := strings.Split(p.Text, " \t ")
	p.Nodes = make(tNodes, 0, len(texts)+len(p.Images)+2)
	for _, s := range texts {
		p.Nodes = append(p.Nodes, tNode{Tag: "p", Children: tNodes{s}})
	}
	for _, imgurl := range p.Images {
		p.Nodes = append(p.Nodes, tNode{Tag: "img", Attrs: tAttrs{"src": imgurl}})
	}
	clearUrl := strings.Split(p.Url, "?")[0]
	aHref := tNode{Tag: "a", Attrs: tAttrs{"href": p.Url}, Children: tNodes{clearUrl}}
	p.Nodes = append(p.Nodes, tNode{Tag: "hr"})
	p.Nodes = append(p.Nodes, tNode{Tag: "p", Children: tNodes{"原文：", aHref}})
}

func filterXsecToken(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	newValues := url.Values{}
	for _, key := range []string{"xsec_token", "xsec_source"} {
		if value := u.Query().Get(key); value != "" {
			newValues.Set(key, value)
		}
	}
	u.RawQuery = newValues.Encode()
	return u.String()
}

func loadXhsCookie() []*http.Cookie {
	if cookie := strings.TrimSpace(os.Getenv("XHS_COOKIE")); cookie != "" {
		return parseCookieString(cookie)
	}
	paths := []string{"ref/my_cookie.txt", "my_cookie.txt"}
	if path := os.Getenv("XHS_COOKIE_FILE"); path != "" {
		paths = append([]string{path}, paths...)
	}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if cookies := parseCookieString(string(content)); len(cookies) > 0 {
			return cookies
		}
	}
	return nil
}

func parseCookieString(raw string) []*http.Cookie {
	var cookies []*http.Cookie
	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		cookies = append(cookies, &http.Cookie{Name: kv[0], Value: kv[1]})
	}
	return cookies
}

func fetchXhsHTML(link, userAgent, accept string, cookies []*http.Cookie) (string, error) {
	req := client.R().SetHeader("User-Agent", userAgent).SetHeader("Accept", accept)
	if len(cookies) > 0 {
		req.SetCookies(cookies)
	}
	resp, err := req.Get(link)
	if err != nil {
		return "", fmt.Errorf("Link inaccessible: %w", err)
	}
	var finalURL string
	if resp.RawResponse != nil && resp.RawResponse.Request != nil && resp.RawResponse.Request.URL != nil {
		finalURL = resp.RawResponse.Request.URL.String()
	}
	switch {
	case strings.Contains(finalURL, "/login"):
		return "", fmt.Errorf("查看该帖子需要登录。")
	case strings.Contains(finalURL, "/404/sec"):
		return "", fmt.Errorf("被风控拦截。")
	case strings.Contains(finalURL, "/404"):
		return "", fmt.Errorf("帖子不存在。")
	}
	return resp.String(), nil
}

func extractInitialState(content string) (map[string]any, error) {
	idx := strings.Index(content, "window.__INITIAL_STATE__")
	if idx < 0 {
		return nil, fmt.Errorf("No data found")
	}
	rest := content[idx:]
	eq := strings.Index(rest, "=")
	if eq < 0 {
		return nil, fmt.Errorf("No data found")
	}
	rest = rest[eq+1:]
	end := strings.Index(rest, "</script>")
	if end < 0 {
		return nil, fmt.Errorf("No data found")
	}
	raw := strings.TrimSpace(rest[:end])
	raw = strings.TrimSuffix(raw, ";")
	raw = reUndefined.ReplaceAllString(raw, "null")
	raw = reNewSet.ReplaceAllString(raw, "$1")
	var data map[string]any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, fmt.Errorf("Failed to parse page data: %w", err)
	}
	return data, nil
}

func extractNote(data map[string]any) map[string]any {
	if h5, ok := data["noteData"].(map[string]any); ok {
		h5Data, _ := h5["data"].(map[string]any)
		if note, ok := h5Data["noteData"].(map[string]any); ok {
			return note
		}
	}
	if pc, ok := data["note"].(map[string]any); ok {
		firstID, _ := pc["firstNoteId"].(string)
		mapData, _ := pc["noteDetailMap"].(map[string]any)
		if entry, ok := mapData[firstID].(map[string]any); ok {
			if note, ok := entry["note"].(map[string]any); ok {
				return note
			}
		}
	}
	return nil
}

func getTraceID(imgURL string) string {
	path := imgURL
	if u, err := url.Parse(imgURL); err == nil {
		path = u.Path
	}
	if i := strings.Index(path, "!"); i >= 0 {
		path = path[:i]
	}
	if m := reTraceID.FindStringSubmatch(path); len(m) == 2 {
		return m[1]
	}
	return strings.TrimPrefix(path, "/")
}

func getRawImageURL(imgURL string) string {
	return "https://sns-img-hw.xhscdn.com/" + getTraceID(imgURL) + "?imageView2/format/jpg"
}

func cleanHashtags(desc string) string {
	desc = reHashtag.ReplaceAllStringFunc(desc, func(hashtag string) string {
		return strings.ReplaceAll(strings.TrimSpace(hashtag), "[话题]#", "") + " "
	})
	return strings.TrimSpace(desc)
}

func extractImages(note map[string]any) []string {
	var images []string
	imageList, _ := note["imageList"].([]any)
	for _, item := range imageList {
		image, ok := item.(map[string]any)
		if !ok {
			continue
		}
		src, _ := image["urlDefault"].(string)
		if src == "" {
			src, _ = image["url"].(string)
		}
		if src == "" {
			continue
		}
		images = append(images, getRawImageURL(src))
	}
	return images
}

func (p *Page) parseContent(content string) error {
	data, err := extractInitialState(content)
	if err != nil {
		return err
	}
	note := extractNote(data)
	if note == nil {
		return fmt.Errorf("该帖子需要登录后查看")
	}
	p.Title, _ = note["title"].(string)
	desc, _ := note["desc"].(string)
	p.Text = cleanHashtags(desc)
	p.Images = extractImages(note)
	p.genNodes()
	return nil
}

func NewXhsPage(link string) (Page, error) {
	link = filterXsecToken(link)
	page := Page{Url: link}

	if cookies := loadXhsCookie(); len(cookies) > 0 {
		if content, err := fetchXhsHTML(link, xhsPCUserAgent, xhsPCAccept, cookies); err == nil {
			if err := page.parseContent(content); err == nil {
				return page, nil
			}
		}
		page = Page{Url: link}
	}

	content, err := fetchXhsHTML(link, xhsMobileUserAgent, xhsMobileAccept, nil)
	if err != nil {
		return page, err
	}
	if err := page.parseContent(content); err != nil {
		return page, err
	}
	return page, nil
}

func UploadXHSTelegraph(p Page) (*telegraph.Page, error) {
	resp, err := telegraph.CreatePage(telegraph.CreatePageOpts{
		Title:       p.Title,
		Content:     p.Nodes,
		AccessToken: telegraphToken,
	})
	if err != nil {
		return nil, fmt.Errorf("Upload telegraph error: %w", err)
	}
	return resp, nil
}

func FetchAndUpload(link string) string {
	page, err := NewXhsPage(link)
	if err != nil {
		return err.Error()
	}
	resp, err := UploadXHSTelegraph(page)
	if err != nil {
		return err.Error()
	}
	return resp.URL
}
