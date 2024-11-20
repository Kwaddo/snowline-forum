package main

import (
	"database/sql"
	"db/internal/sqlite"
	"net/http"
	"sync"
	"log"
	"os"
)

type app struct {
	users *sqlite.USERMODEL
	posts *sqlite.POSTMODEL
	mu    sync.Mutex
}

func NewApp() *app {
	db, err := sql.Open("sqlite3", "./internal/sqlite/app.db")
	if err != nil {
		panic(err)
	}
	if err2 := runSQLFile(db, "./internal/sqlite/tables.sql"); err2 != nil {
		log.Fatalf("Error executing tables.sql: %v", err)
	}
	return &app{
		users: &sqlite.USERMODEL{DB: db},
		posts: &sqlite.POSTMODEL{DB: db},
	}
}

func runSQLFile(db *sql.DB, filepath string) error {
	sqlBytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	sqlCommands := string(sqlBytes)
	_, err = db.Exec(sqlCommands)
	return err
}

func (a *app) StartServer() *http.Server {
	return &http.Server{
		Addr:    ":3333",
		Handler: a.routes(),
	}
}
