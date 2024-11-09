package main

import (
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	app := NewApp()
	go app.CleanupExpiredSessions()
	server := app.StartServer()
	log.Println("Server Start at :3333")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
