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

# Устанавливаем необходимые зависимости, включая Python
RUN apt-get update && apt-get install -y \
    ca-certificates \
    python3 \
    python3-venv \
    && rm -rf /var/lib/apt/lists/*

# Создаём виртуальное окружение для Python
RUN python3 -m venv /root/venv
ENV PATH="/root/venv/bin:$PATH"

# Устанавливаем зависимости для Python-скрипта
RUN pip install --no-cache-dir matplotlib

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /root/

RUN mkdir /root/python_scripts

# Копируем скомпилированное бинарное приложение из builder-контейнера
COPY --from=builder /app/bot .
COPY --from=builder /app/python_scripts/generate_pie_chart.py ./python_scripts
# COPY --from=builder /app/config.yaml .

# Указываем команду для запуска бота
CMD ["./bot"]
