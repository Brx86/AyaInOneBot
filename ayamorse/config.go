package ayamorse

// dot timing in seconds at 1 WPM (1 word = 50 units → 1.2s per unit)
const baseUnit = 1.2

// SampleRate is the WAV sample rate.
const SampleRate = 8000

// ToneFreq is the Morse audio tone frequency in Hz.
const ToneFreq = 700

// ToneAmp is the peak amplitude for 8-bit unsigned PCM (0-255). 192 is loud but not clipped.
const ToneAmp = 192

// Timing helpers: return sample counts at given WPM.
func dotSamples(wpm int) int {
	if wpm < 1 {
		wpm = 20
	}
	return int(float64(SampleRate) * baseUnit / float64(wpm))
}

func dashSamples(wpm int) int       { return dotSamples(wpm) * 3 }
func elementGapSamples(wpm int) int { return dotSamples(wpm) }
func charGapSamples(wpm int) int    { return dotSamples(wpm) * 3 }
func wordGapSamples(wpm int) int    { return dotSamples(wpm) * 7 }

// Morse lookup: A-Z, 0-9, common punctuation.
var morse = map[byte]string{
	'A': ".-", 'B': "-...", 'C': "-.-.", 'D': "-..", 'E': ".",
	'F': "..-.", 'G': "--.", 'H': "....", 'I': "..", 'J': ".---",
	'K': "-.-", 'L': ".-..", 'M': "--", 'N': "-.", 'O': "---",
	'P': ".--.", 'Q': "--.-", 'R': ".-.", 'S': "...", 'T': "-",
	'U': "..-", 'V': "...-", 'W': ".--", 'X': "-..-", 'Y': "-.--",
	'Z': "--..",
	'0': "-----", '1': ".----", '2': "..---", '3': "...--", '4': "....-",
	'5': ".....", '6': "-....", '7': "--...", '8': "---..", '9': "----.",
	'.': ".-.-.-", ',': "--..--", '?': "..--..", '!': "-.-.--",
	'/': "-..-.", '(': "-.--.", ')': "-.--.-", '&': ".-...",
	':': "---...", ';': "-.-.-.", '=': "-...-", '+': ".-.-.",
	'-': "-....-", '_': "..--.-", '"': ".-..-.", '$': "...-..-",
	'@': ".--.-.", '\'': ".----.",
	' ': "/",
}

// Xiaohe double-pinyin lookup tables (from dp.py).
var firstPy = map[string]string{
	"ch": "i", "sh": "u", "zh": "v",
}

var secondPy = map[string]string{
	"ua": "x", "ei": "w", "e": "e", "ou": "z", "iu": "q",
	"ve": "t", "ue": "t", "u": "u", "i": "i", "o": "o",
	"uo": "o", "ie": "p", "a": "a", "ong": "s", "iong": "s",
	"ai": "d", "ing": "k", "uai": "k", "ang": "h", "uan": "r",
	"an": "j", "en": "f", "ia": "x", "iang": "l", "uang": "l",
	"eng": "g", "in": "b", "ao": "c", "v": "v", "ui": "v",
	"un": "y", "iao": "n", "ian": "m",
}

var specialPy = map[string]string{
	"a": "aa", "ai": "ai", "an": "an", "ang": "ah", "ao": "ao",
	"e": "ee", "ei": "ei", "en": "en", "er": "er",
	"o": "oo", "ou": "ou",
}
