#!/bin/bash

# Скрипт для пересборки и запуска бота

echo "🔨 Собираем Docker образ..."
docker build -t time-counter-bot:latest .

if [ $? -eq 0 ]; then
    echo "✅ Образ собран успешно!"
    
    echo "🛑 Останавливаем старые контейнеры..."
    docker-compose down
    
    echo "🚀 Запускаем обновленный бот..."
    docker-compose up -d
    
    echo "📋 Логи бота:"
    docker-compose logs -f telegram-bot
else
    echo "❌ Ошибка при сборке образа!"
    exit 1
fi