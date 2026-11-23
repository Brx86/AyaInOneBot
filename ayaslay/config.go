package ayaslay

import (
	_ "embed"
	"encoding/json"
	"log"
)

type DataConfig struct {
	BaseURL   string
	CardList  []string
	OtherList []string
	FileList  []string
}

var Data DataConfig

//go:embed data.json
var fileData []byte

func init() {
	err := json.Unmarshal(fileData, &Data)
	if err != nil {
		panic("SLAY│ data.json 解析失败: " + err.Error())
	}
	Data.FileList = append(Data.CardList, Data.OtherList...)
	log.Println("SLAY│", "Loads", len(Data.CardList), "cards,", len(Data.OtherList), "others.")
}
