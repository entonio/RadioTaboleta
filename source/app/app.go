package app

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"main/libx/taboleta"
	"main/libx/translate"

	"github.com/getlantern/systray"
)

type Radio struct {
	Name       string
	Address    string
	Separator  *string
	Predefined bool
}

type Settings struct {
	Mpd        string
	Volume     int
	Zapping    time.Duration
	Playback   string
	Radio      string
	Trim       *regexp.Regexp
	Language   string
	SQLite     string
	Restart    []string
	Predefined Radio
}

var DEFAULT_SETTINGS = Settings{
	Mpd:      "localhost:32123",
	Volume:   50,
	Zapping:  30,
	Playback: "let",
	Radio:    "let",
	Trim:     nil,
}

var QUERY_STATUS_INTERVAL = time.Duration(2 * 3.141592 * float64(time.Second))

var CONFIG_DIR string
var SETTINGS Settings
var RADIOS []Radio
var LARGEST_TITLE_LENGTH = 20

var MUSICAS = NewMusicas()

var TRANSLATOR translate.Translator

func Start() {
	CONFIG_DIR = findConfigDir()
	SETTINGS = readSettings(CONFIG_DIR)
	log.Printf("SETTINGS: %+v\n", SETTINGS)
	radios := readRadios(CONFIG_DIR)
	log.Printf("RADIOS: %+v\n", radios)
	TRANSLATOR = translate.NewTranslator(filepath.Join(CONFIG_DIR, "translations.csv"), "pt")

	checkAddresses := false
	if checkAddresses {
		N := len(radios)

		found := make([]bool, N)
		semaphore := make(chan int, N)

		for i, radio := range radios {
			go func(i int, radio Radio) {
				found[i] = check(radio.Address)
				semaphore <- 1
			}(i, radio)
		}
		for i := 0; i < N; i += 1 {
			<-semaphore
		}
		for i, radio := range radios {
			if found[i] {
				RADIOS = append(RADIOS, radio)
			}
		}

	} else {
		RADIOS = radios
	}

	for _, radio := range RADIOS {
		if len(radio.Name) > LARGEST_TITLE_LENGTH {
			LARGEST_TITLE_LENGTH = len(radio.Name)
		}
	}

	setupMpdClient()

	systray.Run(buildUI, onStopSystray)
}

func check(s string) bool {
	r, e := http.Head(s)
	return e == nil && r.StatusCode >= 100 && r.StatusCode < 400
}

func findConfigDir() string {
	var dir string
	for _, parameter := range os.Args[1:] {
		// ignore -psn_0_... (process serial number)
		if !strings.HasPrefix(parameter, "-") {
			dir = parameter
		}
	}

	if len(dir) == 0 {
		dir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
		dir += "/config/"
	} else if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	//if runtime.GOOS == Windows {
	//	dir += "win/"
	//} else {
	//	dir += "mac/"
	//}

	node, _ := os.Stat(dir)
	if node == nil || !node.IsDir() {
		log.Fatalf("Could not open config dir at [%s]", dir)
	}
	return dir
}

const (
	Windows = "windows"
)

func readIconFromConfig() []byte {
	var file string
	if runtime.GOOS == Windows {
		file = "systray.ico"
	} else {
		file = "menubar.icns"
	}
	data, err := ioutil.ReadFile(CONFIG_DIR + file)
	if err != nil {
		log.Println(err)
	}
	return data
}

func readSettings(configDir string) (settings Settings) {
	settings = DEFAULT_SETTINGS

	text, err := os.ReadFile(filepath.Join(configDir, "Settings.taboleta"))
	if err != nil {
		log.Println(err)
		return
	}

	for _, line := range strings.Split(string(text), "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.Contains(line, " ") {
			log.Printf("Unexpected line: [%s]", line)
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		key, value := parts[0], strings.TrimSpace(parts[1])
		switch key {
		case "Mpd":
			settings.Mpd = value
		case "Volume":
			settings.Volume, _ = strconv.Atoi(value)
		case "Zapping":
			s, _ := strconv.Atoi(value)
			settings.Zapping = time.Duration(s) * time.Second
		case "Playback":
			settings.Playback = value
		case "Radio":
			settings.Radio = value
		case "Trim":
			settings.Trim, _ = regexp.Compile(value)
		case "Language":
			settings.Language = value
		case "SQLite":
			settings.SQLite = value
		case "Restart":
			settings.Restart = strings.Split(value, " ")
		default:
			log.Printf("Unexpected key: [%s]", key)
		}
	}
	return
}

func readRadios(configDir string) (radios []Radio) {
	text, err := taboleta.TextAtPath(filepath.Join(configDir, "Radios.taboleta"))
	if err != nil {
		log.Println(err)
		return
	}
	var separator *string
	for _, line := range taboleta.ContentLines(text) {
		if strings.HasPrefix(line, "-") {
			s := strings.TrimSpace(strings.TrimPrefix(line, "-"))
			separator = &s
			continue
		}
		name, address := taboleta.NameAndAddress(line)
		if len(name) == 0 {
			name = address
		}
		predefined := false
		if strings.HasPrefix(name, "*") {
			name = strings.TrimSpace(name[1:])
			predefined = true
		}
		radios = append(radios, Radio{Name: name, Address: address, Separator: separator, Predefined: predefined})
		separator = nil
	}
	return
}

func t(key string) string {
	return TRANSLATOR.Translate(SETTINGS.Language, key)
}

func tf(key string, parameters ...any) string {
	return TRANSLATOR.TranslateFormatted(SETTINGS.Language, key, parameters...)
}

func AsInt(s string, onError int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return onError
	}
	return i
}

func A(parameters ...any) []any {
	if len(parameters) > 0 {
		return parameters
	}
	var empty []any
	return empty
}

func ints(start int, step int, limit int) (a []int) {
	for i := start; i <= limit; i += step {
		a = append(a, i)
	}
	return
}
