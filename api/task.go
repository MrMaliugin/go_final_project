package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"example.com/m/db"
)

var dbInstance *sql.DB

func SetDBInstance(db *sql.DB) {
	dbInstance = db
}

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		addTask(w, r)
	case "GET":
		getTaskByID(w, r)
	case "PUT":
		updateTask(w, r)
	case "DELETE":
		deleteTask(w, r)
	default:
		log.Println("Unsupported HTTP method:", r.Method)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func addTask(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		http.Error(w, "{\"error\": \"Неверный формат данных\"}", http.StatusBadRequest)
		return
	}

	task.CreatedAt = time.Now().Format("2006-01-02 15:04:05")

	id, err := db.AddTask(dbInstance, task)
	if err != nil {
		log.Println("Error adding task to database:", err)
		http.Error(w, "{\"error\": \"Ошибка добавления задачи\"}", http.StatusInternalServerError)
		return
	}

	log.Println("Task added with ID:", id)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("{\"id\": %d}", id)))
}

func getTaskByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		log.Println("Task ID not provided")
		http.Error(w, "{\"error\": \"Не указан идентификатор\"}", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		log.Println("Invalid Task ID:", idParam)
		http.Error(w, "{\"error\": \"Неверный идентификатор\"}", http.StatusBadRequest)
		return
	}

	task, err := db.GetTaskByID(dbInstance, int64(id))
	if err != nil {
		log.Println("Task not found:", id)
		http.Error(w, "{\"error\": \"Задача не найдена\"}", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		http.Error(w, "{\"error\": \"Неверный формат данных\"}", http.StatusBadRequest)
		return
	}

	if task.ID == 0 {
		log.Println("Task ID not provided for update")
		http.Error(w, "{\"error\": \"Не указан идентификатор\"}", http.StatusBadRequest)
		return
	}

	err = db.UpdateTask(dbInstance, task)
	if err != nil {
		log.Println("Error updating task:", err)
		http.Error(w, "{\"error\": \"Ошибка добавления задачи\"}", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}")) // Возвращаем пустой JSON при успешном обновлении
}
func TaskListHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	tasks, err := searchTasks(search)
	if err != nil {
		log.Println("Error retrieving tasks from database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка получения задач"})
		return
	}

	// Форматирование задач в JSON с полями строкового типа
	formattedTasks := []map[string]string{}
	for _, task := range tasks {
		formattedTask := map[string]string{
			"id":      fmt.Sprintf("%d", task.ID),
			"date":    task.Date,
			"title":   task.Title,
			"comment": task.Comment,
			"repeat":  task.Repeat,
		}
		formattedTasks = append(formattedTasks, formattedTask)
	}

	if formattedTasks == nil {
		formattedTasks = []map[string]string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"tasks": formattedTasks})
}

func searchTasks(search string) ([]db.Task, error) {
	if search == "" {
		return db.GetTasks(dbInstance, 50)
	}

	// Проверка на соответствие формату даты 02.01.2006
	if searchDate, err := time.Parse("02.01.2006", search); err == nil {
		return db.GetTasksByDate(dbInstance, searchDate.Format("20060102"))
	}

	// Если не дата, используем поиск по заголовку и комментарию
	searchPattern := "%" + search + "%"
	return db.SearchTasks(dbInstance, searchPattern)
}

func TaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		log.Println("Task ID not provided")
		http.Error(w, "{\"error\": \"Не указан идентификатор\"}", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		log.Println("Invalid Task ID:", idParam)
		http.Error(w, "{\"error\": \"Неверный идентификатор\"}", http.StatusBadRequest)
		return
	}

	task, err := db.GetTaskByID(dbInstance, int64(id))
	if err != nil {
		log.Println("Task not found:", id)
		http.Error(w, "{\"error\": \"Задача не найдена\"}", http.StatusNotFound)
		return
	}

	if task.Repeat == "" {
		// Одноразовая задача, нужно удалить
		err = db.DeleteTask(dbInstance, int64(id))
		if err != nil {
			log.Println("Error deleting task:", err)
			http.Error(w, "{\"error\": \"Ошибка удаления задачи\"}", http.StatusInternalServerError)
			return
		}
	} else {
		// Периодическая задача, нужно обновить дату выполнения
		now := time.Now()
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			log.Println("Error calculating next date:", err)
			http.Error(w, "{\"error\": \"Ошибка расчета следующей даты\"}", http.StatusInternalServerError)
			return
		}

		task.Date = nextDate
		err = db.UpdateTask(dbInstance, task)
		if err != nil {
			log.Println("Error updating task:", err)
			http.Error(w, "{\"error\": \"Ошибка обновления задачи\"}", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		log.Println("Task ID not provided")
		http.Error(w, "{\"error\": \"Не указан идентификатор\"}", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		log.Println("Invalid Task ID:", idParam)
		http.Error(w, "{\"error\": \"Неверный идентификатор\"}", http.StatusBadRequest)
		return
	}

	err = db.RemoveTask(dbInstance, int64(id))
	if err != nil {
		log.Println("Error deleting task:", err)
		http.Error(w, "{\"error\": \"Ошибка удаления задачи\"}", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
}
