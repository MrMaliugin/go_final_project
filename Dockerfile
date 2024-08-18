# Сначала используем образ golang для сборки бинарника
FROM golang:1.22.3 AS build

WORKDIR /app

# Копируем исходный код в рабочую директорию
COPY . .

# Собираем Go приложение
RUN go build -o main .

# Используем минимальный образ Ubuntu для запуска готового приложения
FROM ubuntu:latest

WORKDIR /app

# Копируем скомпилированное приложение из предыдущего этапа
COPY --from=build /app/main /app/main
COPY --from=build /app/web /app/web

# Открываем порт
EXPOSE 7540

# Устанавливаем переменные окружения
ENV TODO_PORT=7540
ENV TODO_PASSWORD=supersecretpassword
ENV TODO_DBFILE=scheduler.db

# Запуск приложения
CMD ["./main"]
