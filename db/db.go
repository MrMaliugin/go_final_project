package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type Task struct {
	ID        int64  `json:"id"`
	Date      string `json:"date"`
	Title     string `json:"title"`
	Comment   string `json:"comment"`
	Repeat    string `json:"repeat"`
	CreatedAt string `json:"created_at"`
}

func InitializeDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./scheduler.db")
	if err != nil {
		log.Println("Error opening database:", err)
		return nil, err
	}

	createTableSQL := `
    CREATE TABLE IF NOT EXISTS scheduler (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        date TEXT NOT NULL,
        title TEXT NOT NULL,
        comment TEXT,
        repeat TEXT,
        created_at TEXT NOT NULL
    );
    `
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Println("Error creating table:", err)
		return nil, err
	}

	log.Println("Таблица scheduler успешно создана или уже существует.")

	return db, nil
}

func AddTask(db *sql.DB, task Task) (int64, error) {
	query := `
        INSERT INTO scheduler (date, title, comment, repeat, created_at)
        VALUES (?, ?, ?, ?, ?)
    `
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.CreatedAt)
	if err != nil {
		log.Println("Error inserting task into database:", err)
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Println("Error getting last insert ID:", err)
		return 0, err
	}

	return id, nil
}

func GetTaskByID(db *sql.DB, id int64) (Task, error) {
	query := `
        SELECT id, date, title, comment, repeat, created_at
        FROM scheduler
        WHERE id = ?
    `
	var task Task
	err := db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat, &task.CreatedAt)
	if err != nil {
		log.Println("Error querying task by ID:", err)
		return Task{}, err
	}
	return task, nil
}

func UpdateTask(db *sql.DB, task Task) error {
	query := `
        UPDATE scheduler
        SET date = ?, title = ?, comment = ?, repeat = ?
        WHERE id = ?
    `
	_, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		log.Println("Error updating task:", err)
		return err
	}
	return nil
}

func GetTasks(db *sql.DB, limit int) ([]Task, error) {
	query := `
        SELECT id, date, title, comment, repeat, created_at
        FROM scheduler
        ORDER BY date ASC
        LIMIT ?
    `
	rows, err := db.Query(query, limit)
	if err != nil {
		log.Println("Error querying tasks from database:", err)
		return nil, err
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat, &task.CreatedAt)
		if err != nil {
			log.Println("Error scanning task row:", err)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error iterating over task rows:", err)
		return nil, err
	}

	return tasks, nil
}

func SearchTasks(db *sql.DB, search string) ([]Task, error) {
	searchPattern := "%" + search + "%"
	query := `
        SELECT id, date, title, comment, repeat, created_at
        FROM scheduler
        WHERE title LIKE ? OR comment LIKE ?
        ORDER BY date ASC
    `
	rows, err := db.Query(query, searchPattern, searchPattern)
	if err != nil {
		log.Println("Error searching tasks:", err)
		return nil, err
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat, &task.CreatedAt)
		if err != nil {
			log.Println("Error scanning task row:", err)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error iterating over task rows:", err)
		return nil, err
	}

	return tasks, nil
}

func GetTasksByDate(db *sql.DB, date string) ([]Task, error) {
	query := `
        SELECT id, date, title, comment, repeat, created_at
        FROM scheduler
        WHERE date = ?
        ORDER BY date ASC
    `
	rows, err := db.Query(query, date)
	if err != nil {
		log.Println("Error querying tasks by date:", err)
		return nil, err
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat, &task.CreatedAt)
		if err != nil {
			log.Println("Error scanning task row:", err)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error iterating over task rows:", err)
		return nil, err
	}

	return tasks, nil
}

func DeleteTask(db *sql.DB, id int64) error {
	query := "DELETE FROM scheduler WHERE id = ?"
	_, err := db.Exec(query, id)
	if err != nil {
		log.Println("Error deleting task:", err)
		return err
	}
	return nil
}

func RemoveTask(db *sql.DB, id int64) error {
	query := "DELETE FROM scheduler WHERE id = ?"
	_, err := db.Exec(query, id)
	if err != nil {
		log.Println("Error deleting task:", err)
		return err
	}
	return nil
}
