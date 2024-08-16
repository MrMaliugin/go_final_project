package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const FormatDate = "20060102"

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowParam := r.FormValue("now")
	dateParam := r.FormValue("date")
	repeatParam := r.FormValue("repeat")

	now, err := time.Parse(FormatDate, nowParam)
	if err != nil {
		log.Println("Неверный формат даты 'now':", err)
		http.Error(w, "Invalid 'now' date format", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateParam, repeatParam)
	if err != nil {
		log.Println("Ошибка расчета следующей даты:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte(nextDate))
}

func NextDate(now time.Time, date string, repeat string) (string, error) {
	taskDate, err := time.Parse(FormatDate, date)
	if err != nil {
		return "", fmt.Errorf("Invalid date format")
	}

	if repeat == "" {
		return "", fmt.Errorf("No repeat rule specified")
	}

	rule := strings.Fields(repeat)
	if len(rule) == 0 {
		return "", fmt.Errorf("Invalid repeat rule")
	}

	switch rule[0] {
	case "d":
		if len(rule) != 2 {
			return "", fmt.Errorf("Invalid repeat rule")
		}
		days, err := strconv.Atoi(rule[1])
		if err != nil || days <= 0 || days > 400 {
			return "", fmt.Errorf("Invalid days value")
		}
		for taskDate.Before(now) || taskDate.Equal(now) {
			taskDate = taskDate.AddDate(0, 0, days)
		}
	case "y":
		for taskDate.Before(now) || taskDate.Equal(now) {
			taskDate = taskDate.AddDate(1, 0, 0)
		}
	default:
		return "", fmt.Errorf("Unsupported repeat rule")
	}

	return taskDate.Format(FormatDate), nil
}
