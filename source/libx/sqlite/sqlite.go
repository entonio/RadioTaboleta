package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func Exec(path string, queries ...[]any) (results []sql.Result, err error) {
	if len(path) == 0 {
		err = fmt.Errorf("No database path configured")
		log.Println(err)
		return
	}
	err = os.MkdirAll(filepath.Dir(path), os.ModeDir)
	if err != nil {
		log.Printf("Could not ensure parent folder for [%s]: %s\n", path, err)
		return
	}
	cs := fmt.Sprintf(
		"file:%s?parseTime=true",
		path,
	)
	log.Printf("Connecting SQLite... %s\n", cs)
	db, err := sql.Open("sqlite3", cs)
	if err != nil {
		log.Printf("Could not connect to SQLite %s: %s\n", cs, err)
		return
	}
	defer db.Close()

	for _, query := range queries {
		result, err := db.Exec(query[0].(string), query[1:]...)
		if err != nil {
			log.Printf("Could not run query on SQLite %s: %s\n", cs, err)
			break
		}
		results = append(results, result)
	}
	return
}
