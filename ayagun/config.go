package ayagun

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
		panic(" GUN│ data.json 解析失败: " + err.Error())
	}
	log.Println(" GUN│", Data.Prefix, len(Data.FileList), "data.json loaded.")
}
