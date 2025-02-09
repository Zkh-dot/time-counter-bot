# Шаг 1: Используем официальный образ Golang для сборки
FROM golang:1.23 AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем все файлы проекта в контейнер
COPY . .

# Загружаем зависимости и собираем приложение
RUN go mod tidy
RUN GOOS=linux GOARCH=amd64 go build -o bot .
RUN chmod +x ./bot

# Шаг 2: Минимальный образ для запуска (без Golang)
FROM ubuntu:latest

# Устанавливаем необходимые зависимости
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /root/

# Копируем скомпилированное бинарное приложение из builder-контейнера
COPY --from=builder /app/bot .
COPY --from=builder /app/fonts .
# COPY --from=builder /app/config.yaml .

# Указываем команду для запуска бота
CMD ["./bot"]
