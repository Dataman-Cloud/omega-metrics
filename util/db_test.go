package util

import (
	"log"
	"testing"
)

func TestDB(t *testing.T) {
	db := DB()
	rows, err := db.Query("select name from application")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		t.Logf("%s", name)
	}
}
