package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/MrMaliugin/go_final_project/api"
	"github.com/MrMaliugin/go_final_project/db"
)

var dbInstance *sql.DB
var jwtKey = []byte("my_secret_key")

type Credentials struct {
	Password string `json:"password"`
}

type Claims struct {
	PasswordHash string `json:"password_hash"`
	jwt.StandardClaims
}

func main() {
	storeInstance, err := db.NewStore()
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer storeInstance.Close()
	fmt.Println("TODO_PASSWORD:", os.Getenv("TODO_PASSWORD"))

	api.SetStoreInstance(storeInstance)

	http.HandleFunc("/api/task", authMiddleware(api.TaskHandler)) // GET, POST, DELETE
	http.HandleFunc("/api/tasks", authMiddleware(api.TaskListHandler))
	http.HandleFunc("/api/task/done", authMiddleware(api.TaskDoneHandler))
	http.HandleFunc("/api/nextdate", api.NextDateHandler)
	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.HandleFunc("/api/signin", SignInHandler)

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	log.Printf("Запуск сервера на порту %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// SignInHandler - обработчик для входа и получения JWT токена
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	TODO_PASSWORD := os.Getenv("TODO_PASSWORD")
	if TODO_PASSWORD == "" || creds.Password != TODO_PASSWORD {
		http.Error(w, `{"error": "Неверный пароль"}`, http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(8 * time.Hour)
	claims := &Claims{
		PasswordHash: TODO_PASSWORD,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, `{"error": "Не удалось создать токен"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}
