package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Попытка получить токен из куки
		var tokenStr string
		cookie, err := r.Cookie("token")
		if err == nil {
			tokenStr = cookie.Value
		} else {
			// Если куки нет, пытаемся взять токен из заголовка Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			// Заголовок должен начинаться с "Bearer "
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		// Парсим и валидируем токен
		claims := &jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil // jwtKey должен быть определен
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Если всё в порядке, передаем управление следующему обработчику
		next.ServeHTTP(w, r)
	}
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TODO_PASSWORD := os.Getenv("TODO_PASSWORD")
		if TODO_PASSWORD != "" {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}

			tokenStr := cookie.Value
			claims := &Claims{}

			tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil
			})
			if err != nil || !tkn.Valid || claims.PasswordHash != TODO_PASSWORD {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
