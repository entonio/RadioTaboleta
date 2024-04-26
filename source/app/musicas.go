package app

import (
	"fmt"
	"os"
	"strings"

	"main/libx/sqlite"
)

type Musicas string

func NewMusicas() Musicas {
	path := SETTINGS.SQLite
	if len(path) > 0 && strings.Contains("$HOME", path) {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
		} else {
			if len(home) == 0 {
				home = "."
			}
			path = strings.Replace(path, "$HOME", home, 1)
		}
	}
	return Musicas(path)
}

var ensureDatabase = `
	CREATE TABLE IF NOT EXISTS Musicas (
		Artist TEXT    NOT NULL,
		Song   TEXT    NOT NULL,
		Times  INTEGER NOT NULL,
		UNIQUE (
			Artist,
			Song
		)
	)
`

func (self Musicas) Add(artist string, song string) {
	sqlite.Exec(string(self),
		A(ensureDatabase),
		A("INSERT OR IGNORE INTO Musicas (Artist, Song, Times) VALUES (?, ?, 0)", artist, song),
		A("UPDATE Musicas SET Times = Times + 1 WHERE Artist = ? AND Song = ?", artist, song),
	)
}
