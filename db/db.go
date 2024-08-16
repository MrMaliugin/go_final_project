package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

const (
	MaxTitleLength   = 100
	MaxCommentLength = 500
)

// Task - структура для хранения информации о задаче
type Task struct {
	ID        int64  `json:"id"`
	Date      string `json:"date"`
	Title     string `json:"title"`
	Comment   string `json:"comment"`
	Repeat    string `json:"repeat"`
	CreatedAt string `json:"created_at"`
}

// Store - структура для работы с базой данных и выполнения операций над задачами
type Store struct {
	db *sql.DB
	mu sync.Mutex
}

// NewStore инициализирует базу данных и возвращает новый экземпляр Store
func NewStore() (*Store, error) {
	db, err := sql.Open("sqlite3", "./scheduler.db")
	if err != nil {
		return nil, err
	}

	store := &Store{db: db}
	if err := store.initialize(); err != nil {
		return nil, err
	}

	return store, nil
}

// Закрытие базы данных
func (s *Store) Close() error {
	return s.db.Close()
}

// Создание новой задачи
func (s *Store) CreateTask(task Task) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Валидация задачи перед добавлением
	if err := validateTask(&task); err != nil {
		return 0, err
	}

	query := `INSERT INTO scheduler (date, title, comment, repeat, created_at) VALUES (?, ?, ?, ?, ?)`
	result, err := s.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.CreatedAt)
	if err != nil {
		log.Println("Ошибка создания задачи:", err)
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Получение задачи по ID
func (s *Store) GetTaskByID(id int64) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
        SELECT id, date, title, comment, repeat, created_at
        FROM scheduler
        WHERE id = ?
    `
	var task Task
	err := s.db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat, &task.CreatedAt)
	if err != nil {
		log.Println("Ошибка получения задачи по ID:", err)
		return Task{}, err
	}
	return task, nil
}

// Функция поиска задач по дате в базе данных
func (s *Store) GetTasksByDate(date string) ([]Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
        SELECT id, date, title, comment, repeat, created_at
        FROM scheduler
        WHERE date = ?
        ORDER BY date ASC
    `
	rows, err := s.db.Query(query, date)
	if err != nil {
		log.Printf("Ошибка запроса задач по дате: %v", err)
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat, &task.CreatedAt); err != nil {
			log.Printf("Ошибка сканирования строки задачи: %v", err)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Ошибка при переборе строк задач: %v", err)
		return nil, err
	}

	return tasks, nil
}

// Получение всех задач с ограничением
func (s *Store) GetTasks(limit int) ([]Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
        SELECT id, date, title, comment, repeat, created_at
        FROM scheduler
        ORDER BY date ASC
        LIMIT ?
    `
	rows, err := s.db.Query(query, limit)
	if err != nil {
		log.Println("Ошибка получения списка задач:", err)
		return nil, err
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat, &task.CreatedAt)
		if err != nil {
			log.Println("Ошибка сканирования строки задачи:", err)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		log.Println("Ошибка при итерации по строкам задач:", err)
		return nil, err
	}

	return tasks, nil
}

// Поиск задач по ключевым словам
func (s *Store) SearchTasks(search string) ([]Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	searchPattern := "%" + search + "%"
	query := `
        SELECT id, date, title, comment, repeat, created_at
        FROM scheduler
        WHERE title LIKE ? OR comment LIKE ?
        ORDER BY date ASC
    `
	rows, err := s.db.Query(query, searchPattern, searchPattern)
	if err != nil {
		log.Println("Ошибка поиска задач:", err)
		return nil, err
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat, &task.CreatedAt)
		if err != nil {
			log.Println("Ошибка сканирования строки задачи:", err)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		log.Println("Ошибка при итерации по строкам задач:", err)
		return nil, err
	}

	return tasks, nil
}

// Обновление задачи по ID
func (s *Store) UpdateTask(task Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `UPDATE scheduler SET title = ?, date = ?, comment = ?, repeat = ? WHERE id = ?`
	_, err := s.db.Exec(query, task.Title, task.Date, task.Comment, task.Repeat, task.ID)
	if err != nil {
		log.Printf("Не удалось обновить задачу: %v", err)
		return fmt.Errorf("failed to update task: %v", err)
	}

	log.Printf("Задача обновлена в базе данных: %+v", task)
	return nil
}

// Удаление задачи по ID
func (s *Store) DeleteTask(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `DELETE FROM scheduler WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		log.Println("Ошибка удаления задачи:", err)
		return err
	}

	return nil
}

// Инициализация базы данных (создание таблиц, если необходимо)
func (s *Store) initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
    CREATE TABLE IF NOT EXISTS scheduler (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        date TEXT NOT NULL,
        title TEXT NOT NULL,
        comment TEXT,
        repeat TEXT,
        created_at TEXT NOT NULL
    );
    `
	_, err := s.db.Exec(query)
	if err != nil {
		log.Println("Ошибка создания таблицы:", err)
		return err
	}

	log.Println("Таблица scheduler успешно создана или уже существует.")
	return nil
}

// Валидация длины полей задачи
func validateTask(task *Task) error {
	if len(task.Title) > MaxTitleLength {
		return errors.New("Заголовок слишком длинный, максимальная длина — 100 символов")
	}
	if len(task.Comment) > MaxCommentLength {
		return errors.New("Комментарий слишком длинный, максимальная длина — 500 символов")
	}
	return nil
}
