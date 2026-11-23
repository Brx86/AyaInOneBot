package ayamorse

import (
	"bytes"
	"encoding/binary"
	"math"
	"strings"

	"github.com/mozillazg/go-pinyin"
)

// pinyinToDoublePinyin converts a pinyin string (no tone marks) to
// Xiaohe double-pinyin encoding (2 letters).
func pinyinToDoublePinyin(py string) string {
	if len(py) == 0 {
		return ""
	}

	// Special case: no initial (starts with a/e/o), length <= 3
	if len(py) <= 3 && (py[0] == 'a' || py[0] == 'e' || py[0] == 'o') {
		if v, ok := specialPy[py]; ok {
			return v
		}
		return py
	}

	if v, ok := firstPy[py[:2]]; ok {
		// Compound initial: first[py[:2]] + second[py[2:]]
		if v2, ok2 := secondPy[py[2:]]; ok2 {
			return v + v2
		}
		return v + py[2:]
	}

	// Single initial: py[0] + second[py[1:]]
	if len(py) >= 2 {
		if v, ok := secondPy[py[1:]]; ok {
			return string(py[0]) + v
		}
		return string(py[0]) + py[1:]
	}

	return py
}

// chineseToDoublePinyin converts Chinese characters to double-pinyin using go-pinyin.
func chineseToDoublePinyin(text string) []byte {
	pinyins := pinyin.LazyPinyin(text, pinyin.NewArgs())
	var result []byte
	for _, py := range pinyins {
		dp := pinyinToDoublePinyin(py)
		for _, ch := range dp {
			upper := byte(ch)
			if upper >= 'a' && upper <= 'z' {
				upper -= 32
			}
			result = append(result, upper)
		}
	}
	return result
}

// textToMorse converts input text to Morse code string.
// Chinese characters → go-pinyin → double-pinyin → Morse.
// English/numbers → standard Morse.
// Unknown characters are skipped.
func textToMorse(text string) string {
	var buf bytes.Buffer
	prevWasContent := false
	for _, r := range text {
		if r == ' ' {
			if prevWasContent {
				buf.WriteString(" / ")
			}
			prevWasContent = false
		} else if r >= 'A' && r <= 'Z' {
			if prevWasContent {
				buf.WriteByte(' ')
			}
			buf.WriteString(morse[byte(r)])
			prevWasContent = true
		} else if r >= 'a' && r <= 'z' {
			if prevWasContent {
				buf.WriteByte(' ')
			}
			buf.WriteString(morse[byte(r-32)])
			prevWasContent = true
		} else if r >= '0' && r <= '9' {
			if prevWasContent {
				buf.WriteByte(' ')
			}
			buf.WriteString(morse[byte(r)])
			prevWasContent = true
		} else if r >= 0x4E00 && r <= 0x9FFF {
			// CJK Unified Ideographs: convert via go-pinyin
			letters := chineseToDoublePinyin(string(r))
			if len(letters) == 0 {
				continue
			}
			if prevWasContent {
				buf.WriteString(" / ")
			}
			for j, b := range letters {
				if j > 0 {
					buf.WriteByte(' ')
				}
				if mc, ok := morse[b]; ok {
					buf.WriteString(mc)
				}
			}
			prevWasContent = true
		} else if code, ok := morse[byte(r)]; ok {
			if prevWasContent {
				buf.WriteByte(' ')
			}
			buf.WriteString(code)
			prevWasContent = true
		}
		// else: skip unknown characters
	}
	return buf.String()
}

// audioElement describes one on/off segment of audio.
type audioElement struct {
	samples int
	on      bool
}

// textToAudioElements converts text to a sequence of audio elements at given WPM.
func textToAudioElements(text string, wpm int) []audioElement {
	var elems []audioElement
	runes := []rune(text)
	prevWasContent := false

	for _, r := range runes {
		if r == ' ' {
			if prevWasContent {
				elems = append(elems, audioElement{samples: wordGapSamples(wpm), on: false})
			}
			prevWasContent = false
			continue
		}

		var letters []byte

		if r >= 'A' && r <= 'Z' {
			letters = []byte{byte(r)}
		} else if r >= 'a' && r <= 'z' {
			letters = []byte{byte(r - 32)}
		} else if r >= '0' && r <= '9' {
			letters = []byte{byte(r)}
		} else if r >= 0x4E00 && r <= 0x9FFF {
			letters = chineseToDoublePinyin(string(r))
		}

		if len(letters) == 0 {
			continue
		}

		// Word gap between content characters (not before first)
		if prevWasContent {
			elems = append(elems, audioElement{samples: wordGapSamples(wpm), on: false})
		}

		// Each letter in the double-pinyin code is a Morse character
		for j, b := range letters {
			if j > 0 {
				elems = append(elems, audioElement{samples: charGapSamples(wpm), on: false})
			}

			code := ""
			if mc, ok := morse[b]; ok {
				code = mc
			}

			for k, ch := range code {
				if k > 0 {
					elems = append(elems, audioElement{samples: elementGapSamples(wpm), on: false})
				}
				if ch == '.' {
					elems = append(elems, audioElement{samples: dotSamples(wpm), on: true})
				} else if ch == '-' {
					elems = append(elems, audioElement{samples: dashSamples(wpm), on: true})
				} else if ch == '/' {
					elems = append(elems, audioElement{samples: wordGapSamples(wpm), on: false})
				}
			}
		}
		prevWasContent = true
	}
	return elems
}

// generateWAV produces an 8-bit PCM WAV file at given sample rate.
func generateWAV(text string, wpm int) []byte {
	elems := textToAudioElements(text, wpm)

	// Calculate total samples
	totalSamples := 0
	for _, e := range elems {
		totalSamples += e.samples
	}
	// Add a trailing silence
	totalSamples += SampleRate / 4

	samples := make([]byte, totalSamples)
	pos := 0
	for _, e := range elems {
		if e.on {
			// Generate sine wave
			for i := 0; i < e.samples; i++ {
				t := float64(i) / float64(SampleRate)
				s := int(ToneAmp + ToneAmp*math.Sin(2*math.Pi*ToneFreq*t))
				if s > 255 {
					s = 255
				}
				samples[pos] = byte(s)
				pos++
			}
		} else {
			// Silence
			for i := 0; i < e.samples; i++ {
				samples[pos] = 128 // center line for 8-bit unsigned
				pos++
			}
		}
	}
	// Trailing silence
	for i := pos; i < totalSamples; i++ {
		samples[i] = 128
	}

	// Build WAV file
	dataSize := len(samples)
	fileSize := 36 + dataSize

	var buf bytes.Buffer
	// RIFF header
	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(fileSize))
	buf.WriteString("WAVE")
	// fmt chunk
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16)) // chunk size
	binary.Write(&buf, binary.LittleEndian, uint16(1))  // PCM
	binary.Write(&buf, binary.LittleEndian, uint16(1))  // mono
	binary.Write(&buf, binary.LittleEndian, uint32(SampleRate))
	binary.Write(&buf, binary.LittleEndian, uint32(SampleRate)) // byte rate
	binary.Write(&buf, binary.LittleEndian, uint16(1))          // block align
	binary.Write(&buf, binary.LittleEndian, uint16(8))          // bits per sample
	// data chunk
	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, uint32(dataSize))
	buf.Write(samples)

	return buf.Bytes()
}

// parseWPM extracts WPM from input like "=20 some text" → (20, "some text").
// Returns default WPM (20) if no valid =N prefix found.
func parseWPM(text string) (int, string) {
	const defaultWPM = 20
	if !strings.HasPrefix(text, "=") {
		return defaultWPM, text
	}
	if text == "=" {
		return defaultWPM, ""
	}
	rest := text[1:]
	// Find where the number ends
	end := 0
	for end < len(rest) && rest[end] >= '0' && rest[end] <= '9' {
		end++
	}
	if end == 0 {
		return defaultWPM, text
	}
	n := 0
	for i := 0; i < end; i++ {
		n = n*10 + int(rest[i]-'0')
	}
	if n < 5 || n > 30 {
		return defaultWPM, strings.TrimSpace(rest[end:])
	}
	return n, strings.TrimSpace(rest[end:])
}
