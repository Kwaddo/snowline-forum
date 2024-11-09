package main

import (
	"database/sql"
	"db/internal/sqlite"
	"net/http"
	"sync"
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

	return &app{
		users: &sqlite.USERMODEL{DB: db},
		posts: &sqlite.POSTMODEL{DB: db},
	}
}
func (a *app) StartServer() *http.Server {
	return &http.Server{
		Addr:    ":3333",
		Handler: a.routes(),
	}
}
