package ayab23

import (
	"testing"
)

func TestFilterXsecToken(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "保留token和source，去掉其余参数",
			in:   "https://www.xiaohongshu.com/explore/abc?xsec_token=A=&xsec_source=app_share&type=normal",
			want: "https://www.xiaohongshu.com/explore/abc?xsec_source=app_share&xsec_token=A%3D",
		},
		{
			name: "无query时不追加问号",
			in:   "https://www.xiaohongshu.com/explore/abc",
			want: "https://www.xiaohongshu.com/explore/abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterXsecToken(tt.in); got != tt.want {
				t.Errorf("filterXsecToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetTraceID(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "32位hex路径取末段",
			in:   "https://sns-webpic-qc.xhscdn.com/202501/aabbccdd11223344aabbccdd11223344/img!nd_dft_webp",
			want: "img",
		},
		{
			name: "非hex路径回退到去掉!后缀",
			in:   "https://sns-webpic-qc.xhscdn.com/2025010114/1040g00830thf!nd_dft_wgth_webp_3",
			want: "2025010114/1040g00830thf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTraceID(tt.in); got != tt.want {
				t.Errorf("getTraceID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCleanHashtags(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"内容 #测试[话题]# 结束", "内容#测试 结束"},
		{"无话题", "无话题"},
		{"#开头[话题]#", "#开头"},
	}
	for _, tt := range tests {
		if got := cleanHashtags(tt.in); got != tt.want {
			t.Errorf("cleanHashtags(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

const h5StateHTML = `<script>window.__INITIAL_STATE__={"noteData":{"data":{"noteData":{"title":"H5标题","desc":"正文 #测试[话题]#","imageList":[{"urlDefault":"https://sns-webpic-qc.xhscdn.com/202501/aabbccdd11223344aabbccdd11223344/img!nd_dft_webp"}]}}}};</script>`

const pcStateHTML = `<script>window.__INITIAL_STATE__={"note":{"firstNoteId":"n1","noteDetailMap":{"n1":{"note":{"title":"PC标题","desc":"正文","imageList":[]}}}}};</script>`

func TestExtractInitialStateAndNote(t *testing.T) {
	data, err := extractInitialState(h5StateHTML)
	if err != nil {
		t.Fatalf("extractInitialState(h5) error: %v", err)
	}
	note := extractNote(data)
	if note == nil {
		t.Fatal("extractNote(h5) = nil")
	}
	if title, _ := note["title"].(string); title != "H5标题" {
		t.Errorf("h5 title = %q, want %q", title, "H5标题")
	}

	data, err = extractInitialState(pcStateHTML)
	if err != nil {
		t.Fatalf("extractInitialState(pc) error: %v", err)
	}
	note = extractNote(data)
	if note == nil {
		t.Fatal("extractNote(pc) = nil")
	}
	if title, _ := note["title"].(string); title != "PC标题" {
		t.Errorf("pc title = %q, want %q", title, "PC标题")
	}

	if _, err := extractInitialState("<html>没有数据</html>"); err == nil {
		t.Error("extractInitialState(无标记) 应返回错误")
	}
}
