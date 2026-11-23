package ayalex

import (
	_ "embed"
	"encoding/json"
	"log"
)

type DataConfig struct {
	Prefix   string
	BaseURL  string
	FileList []string
}

var Data DataConfig

//go:embed data.json
var fileData []byte

func init() {
	err := json.Unmarshal(fileData, &Data)
	if err != nil {
		panic(" LEX│ data.json 解析失败: " + err.Error())
	}
	log.Println(" LEX│", Data.Prefix, len(Data.FileList), "data.json loaded.")
}
