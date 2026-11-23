package ayaomo

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// WeatherResponse represents the structure of the API response.
type WeatherResponse struct {
	Daily []struct {
		TextDay   string `json:"textDay"`
		TextNight string `json:"textNight"`
		TempMin   string `json:"tempMin"`
		TempMax   string `json:"tempMax"`
		Humidity  string `json:"humidity"`
	} `json:"daily"`
}

// cacheEntry stores cached data and its expiration time.
type cacheEntry struct {
	temRange string
	expiry   time.Time
}

var (
	// beijing 固定使用东八区，不篡改全局 time.Local。
	beijing     = time.FixedZone("Asia/Shanghai", 8*3600)
	apiKey      string // 由 Start() 从环境变量注入。
	location    = "101010600"
	cache       = make(map[string]*cacheEntry)
	cacheMutex  sync.RWMutex
	apiEndpoint = "https://mb7dn7debh.re.qweatherapi.com/v7/weather/3d"
)

func getWeekdayZh(weekday time.Weekday) string {
	weekdays := [7]string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}
	return weekdays[int(weekday)]
}

// getTemRange fetches weather data and formats the temperature range.
func getTemRange(location string, formatDate bool) (temRange string, err error) {
	resp, err := client.R().
		SetQueryParam("location", location).
		SetQueryParam("key", apiKey).
		Get(apiEndpoint)

	body := resp.Body()

	var weatherResp WeatherResponse
	if err = json.Unmarshal(body, &weatherResp); err != nil {
		return "", fmt.Errorf("decoding JSON response: %w", err)
	}

	today := weatherResp.Daily[0]
	if today.TextDay == today.TextNight {
		temRange = fmt.Sprintf("%s %s~%s℃", today.TextDay, today.TempMin, today.TempMax)
	} else {
		temRange = fmt.Sprintf("%s转%s %s~%s℃", today.TextDay, today.TextNight, today.TempMin, today.TempMax)
	}

	if formatDate {
		now := time.Now().In(beijing)
		weekdayZh := getWeekdayZh(now.Weekday())
		temRange = fmt.Sprintf("%s %s %s 湿度%s%%", now.Format("2006年01月02日"), weekdayZh, temRange, today.Humidity)
	}
	return
}

// cachedGetTemRange retrieves the temperature range from cache or fetches it.
func cachedGetTemRange(location string, formatDate bool) (temRange string, err error) {
	cacheKey := location
	if formatDate {
		cacheKey += "_full" // Differentiate cache keys for different formats
	}

	cacheMutex.RLock()
	entry, ok := cache[cacheKey]
	cacheMutex.RUnlock()

	if ok && time.Now().Before(entry.expiry) {
		log.Println(" OMO│ Cached API:", entry.temRange)
		return entry.temRange, nil
	}

	temRange, err = getTemRange(location, formatDate)
	if err != nil {
		return "", err //  Don't cache errors.  Return immediately.
	}

	cacheMutex.Lock()
	cache[cacheKey] = &cacheEntry{
		temRange: temRange,
		expiry:   time.Now().Add(30 * time.Minute),
	}
	cacheMutex.Unlock()

	log.Println(" OMO│ Request API:", temRange)
	return
}

func getTem() string {
	temRange, _ := cachedGetTemRange(location, true)
	return temRange
}
