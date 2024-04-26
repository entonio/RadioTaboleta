package app

import (
	"encoding/xml"
	"strings"
)

type RadioInfo struct {
	Table        Table
	AnimadorInfo AnimadorInfo
}

type Table struct {
	DB_LEAD_ARTIST_NAME string
	DB_SONG_NAME        string
}

type AnimadorInfo struct {
	TITLE string
	NAME  string
}

func readTitleFromXml(s string) string {

	var fields []string
	addField := func(f string) {
		if len(f) > 0 {
			fields = append(fields, f)
		}
	}
	joined := func() string {
		return strings.Join(fields, " - ")
	}

	var radioInfo RadioInfo
	err := xml.Unmarshal([]byte(s), &radioInfo)
	if err == nil {
		addField(radioInfo.Table.DB_LEAD_ARTIST_NAME)
		addField(radioInfo.Table.DB_SONG_NAME)
		if len(fields) == 0 {
			addField(radioInfo.AnimadorInfo.TITLE)
			addField(radioInfo.AnimadorInfo.NAME)
		}
		if len(fields) > 0 {
			return joined()
		}
	}

	return joined()
}
