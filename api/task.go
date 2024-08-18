package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/MrMaliugin/go_final_project/db"
)

var storeInstance *db.Store

func SetStoreInstance(store *db.Store) {
	storeInstance = store
}

// TaskHandler обрабатывает запросы для работы с задачами (GET, POST, DELETE)
func TaskHandler(w http.ResponseWriter, r *http.Request) {
	if storeInstance == nil {
		http.Error(w, "Store not initialized", http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case "GET":
		// Проверка, запрашивается ли поиск по дате
		searchDate := r.URL.Query().Get("date")
		if searchDate != "" {
			tasks, err := storeInstance.GetTasksByDate(searchDate)
			if err != nil {
				log.Printf("Error retrieving tasks by date: %v", err)
				http.Error(w, "Failed to retrieve tasks by date", http.StatusInternalServerError)
				return
			}

			log.Printf("Tasks retrieved by date: %v", tasks)
			respondWithJSON(w, http.StatusOK, tasks)
			return
		}

		// Получение задачи по ID или возврат списка всех задач, если date не передана
		idParam := r.URL.Query().Get("id")
		if idParam != "" {
			id, err := strconv.ParseInt(idParam, 10, 64)
			if err != nil {
				http.Error(w, "Invalid task ID", http.StatusBadRequest)
				return
			}

			task, err := storeInstance.GetTaskByID(id)
			if err != nil {
				http.Error(w, "Task not found", http.StatusNotFound)
				return
			}

			respondWithJSON(w, http.StatusOK, formatTaskForFrontend(task))
		} else {
			tasks, err := storeInstance.GetTasks(10)
			if err != nil {
				http.Error(w, "Failed to retrieve tasks", http.StatusInternalServerError)
				return
			}

			respondWithJSON(w, http.StatusOK, tasks)
		}

	case "POST":
		// Создание новой задачи
		var task db.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, "Invalid task data", http.StatusBadRequest)
			return
		}

		taskID, err := storeInstance.CreateTask(task)
		if err != nil {
			http.Error(w, "Failed to create task", http.StatusInternalServerError)
			return
		}

		task.ID = taskID
		respondWithJSON(w, http.StatusCreated, formatTaskForFrontend(task))

	case "PUT":
		var task db.Task

		// Попытка извлечь ID задачи из URL параметров, если нет, использовать ID из тела
		idParam := r.URL.Query().Get("id")
		if idParam != "" {
			id, err := strconv.ParseInt(idParam, 10, 64)
			if err != nil {
				log.Printf("Неверный идентификатор задачи из URL: %v", err)
				http.Error(w, "Invalid task ID", http.StatusBadRequest)
				return
			}
			task.ID = id
		}

		// Парсим тело запроса для данных задачи
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			log.Printf("Не удалось декодировать данные задачи: %v", err)
			http.Error(w, "Invalid task data", http.StatusBadRequest)
			return
		}

		// Если ID отсутствует в URL, извлекаем его из тела запроса
		if task.ID == 0 {
			log.Printf("Идентификатор задачи отсутствует как в URL, так и в теле")
			http.Error(w, "Task ID is required", http.StatusBadRequest)
			return
		}

		log.Printf("Получено обновление задачи: %+v", task)

		// Обновляем задачу в базе данных
		if err := storeInstance.UpdateTask(task); err != nil {
			log.Printf("Не удалось обновить задачу в базе данных: %v", err)
			http.Error(w, "Failed to update task", http.StatusInternalServerError)
			return
		}

		// Возвращаем обновлённые данные задачи клиенту
		respondWithJSON(w, http.StatusOK, formatTaskForFrontend(task))

	case "DELETE":
		// Удаление задачи
		idParam := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			http.Error(w, "Invalid task ID", http.StatusBadRequest)
			return
		}

		if err := storeInstance.DeleteTask(id); err != nil {
			http.Error(w, "Failed to delete task", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// TaskListHandler обрабатывает запросы для получения списка задач
func TaskListHandler(w http.ResponseWriter, r *http.Request) {
	if storeInstance == nil {
		http.Error(w, "Store not initialized", http.StatusInternalServerError)
		return
	}

	// Логируем весь запрос и параметры
	log.Printf("URL-адрес запроса: %s", r.URL.String())

	// Изменено: Получаем параметр "search", который приходит с фронтенда
	searchDate := r.URL.Query().Get("search")
	log.Printf("Получена дата поиска: %s", searchDate)

	if searchDate != "" {
		// Пробуем распарсить дату
		parsedDate, err := time.Parse("02.01.2006", searchDate)
		if err != nil {
			log.Printf("Не удалось проанализировать дату: %v", err)
			http.Error(w, "Invalid date format. Use dd.mm.yyyy", http.StatusBadRequest)
			return
		}

		formattedDate := parsedDate.Format("20060102")
		log.Printf("Форматированная дата для запроса к базе данных: %s", formattedDate)

		// Выполняем поиск задач по дате
		tasks, err := storeInstance.GetTasksByDate(formattedDate)
		if err != nil {
			log.Printf("Ошибка при получении задач по дате: %v", err)
			http.Error(w, "Failed to retrieve tasks by date", http.StatusInternalServerError)
			return
		}

		log.Printf("Задачи найдены: %d", len(tasks))

		formattedTasks := make([]map[string]interface{}, len(tasks))
		for i, task := range tasks {
			formattedTasks[i] = formatTaskForFrontend(task)
		}

		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"tasks": formattedTasks,
		})
		return
	}

	log.Println("Дата не указана, получение списка задач по умолчанию.")
	tasks, err := storeInstance.GetTasks(10)
	if err != nil {
		http.Error(w, "Failed to retrieve tasks", http.StatusInternalServerError)
		return
	}

	formattedTasks := make([]map[string]interface{}, len(tasks))
	for i, task := range tasks {
		formattedTasks[i] = formatTaskForFrontend(task)
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"tasks": formattedTasks,
	})
}

// TaskDoneHandler отмечает задачу как выполненную (в данном случае удаляется или обновляется)
func TaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	if storeInstance == nil {
		http.Error(w, "Store not initialized", http.StatusInternalServerError)
		return
	}

	idParam := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := storeInstance.GetTaskByID(id)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Если задача повторяющаяся, переносим её на следующую дату
	if task.Repeat != "" {
		nextDate, err := calculateNextDate(task)
		if err != nil {
			http.Error(w, "Failed to calculate next date", http.StatusInternalServerError)
			return
		}

		task.Date = nextDate
		err = storeInstance.UpdateTask(task)
		if err != nil {
			http.Error(w, "Failed to update task", http.StatusInternalServerError)
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]string{"status": "task updated"})
		return
	}

	// Если задача обычная, удаляем её
	err = storeInstance.DeleteTask(id)
	if err != nil {
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "task done"})
}

// Отправки ответа в формате JSON
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Ошибка маршалинга JSON: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Расчет следующей даты выполнения задачи
func calculateNextDate(task db.Task) (string, error) {
	now := time.Now()
	nextDate, err := NextDate(now, task.Date, task.Repeat)
	if err != nil {
		return "", err
	}
	return nextDate, nil
}

// Функция для форматирования задачи под требования фронтенда
func formatTaskForFrontend(task db.Task) map[string]interface{} {
	return map[string]interface{}{
		"id":      task.ID,      // ID задачи
		"title":   task.Title,   // Название задачи
		"comment": task.Comment, // Комментарий
		"date":    task.Date,    // Дата задачи (может потребоваться форматирование)
		"repeat":  task.Repeat,  // Поле для повторяющихся задач
	}
}
