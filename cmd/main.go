package main

import (
	"database/sql"
	"db/internal/sqlite"
	"log"
	"net/http"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type app struct {
	users *sqlite.USERMODEL
	posts *sqlite.POSTMODEL
	mu    sync.Mutex
}

func main() {
	db, err := sql.Open("sqlite3", "./internal/sqlite/app.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	app := app{
		users: &sqlite.USERMODEL{
			DB: db,
		},
		posts: &sqlite.POSTMODEL{
			DB : db,
		},
	}
	go app.CleanupExpiredSessions()
	server := http.Server{
		Addr:    ":3535",
		Handler: app.routes(),
	}
	log.Println("Server Start at :3535")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

}
