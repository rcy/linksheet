package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/migration"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func init() {
	log.Printf("init db")
	initMigrator()

	var err error
	dbFile, ok := os.LookupEnv("DB_FILE")
	if !ok {
		log.Fatalf("$DB_FILE not found in environment")
	}
	DB, err = migration.Open("sqlite", dbFile, Migrator)
	if err != nil {
		panic(fmt.Sprintf("couldn't open database: %v", err))
	}
}

func TrackRequest(ip, alias, target string, status int) {
	_, err := DB.Exec("insert into requests(ip, alias, target, status) values(?, ?, ?, ?)", ip, alias, target, status)
	if err != nil {
		panic(fmt.Sprintf("failed to track request: %v", err))
	}
}

type Request struct {
	CreatedAt string
	Ip        string
	Status    string
	Alias     string
	Target    string
}

func Requests() ([]Request, error) {
	rows, err := DB.Query("select created_at, ip, alias, target, status from requests order by created_at desc")
	if err != nil {
		log.Printf("Requests: %v", err)
		return nil, err
	}
	defer rows.Close()

	requests := []Request{}
	for rows.Next() {
		request := Request{}
		rows.Scan(&request.CreatedAt, &request.Ip, &request.Alias, &request.Target, &request.Status)
		requests = append(requests, request)
	}
	return requests, nil
}
