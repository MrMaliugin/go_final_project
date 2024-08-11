package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"example.com/m/api"
	"example.com/m/db"
	_ "github.com/mattn/go-sqlite3"
)

var dbInstance *sql.DB

func main() {
	var err error
	dbInstance, err = db.InitializeDatabase()
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %s", err)
	}
	defer dbInstance.Close()

	api.SetDBInstance(dbInstance)

	http.HandleFunc("/api/task", api.TaskHandler)
	http.HandleFunc("/api/nextdate", api.NextDateHandler)
	http.HandleFunc("/api/tasks", api.TaskListHandler)
	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.HandleFunc("/api/task/done", api.TaskDoneHandler)

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	log.Printf("Запуск сервера на порту %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
