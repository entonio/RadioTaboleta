package taboleta

import (
	"io/ioutil"
	"strings"
)

func TextAtPath(path string) (string, error) {
	bytes, err := ioutil.ReadFile(path)
	return string(bytes), err
}

func ContentLines(text string) (lines []string) {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "⭕") {
			continue
		}
		lines = append(lines, line)
	}
	return
}

func NameAndAddress(line string) (name string, address string) {
	parts := strings.Split(line, " ")
	address = parts[len(parts)-1]
	name = strings.TrimSpace(line[0 : len(line)-len(address)])
	return
}
