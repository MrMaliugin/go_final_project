package main

import (
	"log"
	"net/http"
	"os"

	"github.com/MrMaliugin/go_final_project/api"
	"github.com/MrMaliugin/go_final_project/db"
)

func main() {
	storeInstance, err := db.NewStore()
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer storeInstance.Close()

	api.SetStoreInstance(storeInstance)

	http.HandleFunc("/api/task", api.TaskHandler) // GET, POST, DELETE
	http.HandleFunc("/api/nextdate", api.NextDateHandler)
	http.HandleFunc("/api/tasks", api.TaskListHandler)
	http.HandleFunc("/api/task/done", api.TaskDoneHandler)
	http.Handle("/", http.FileServer(http.Dir("./web")))

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	log.Printf("Запуск сервера на порту %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
