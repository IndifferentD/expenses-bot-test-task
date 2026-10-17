# Тестовое задание: ИИ-бот учёта расходов

Решение [тестового задания](./task.md) на позицию Go-разработчика.

## Запуск

```sh
git clone https://github.com/IndifferentD/expenses-tg-bot-test-task
cd expenses-tg-bot-test-task
cp .env.example .env
```

В `.env` укажите `TELEGRAM_BOT_TOKEN` и `OPENAI_API_KEY`. Если нужны прокси, заполните `TELEGRAM_PROXY_URL` и `OPENAI_PROXY_URL`. Затем запустите:

```sh
docker compose up --build
```

Откройте `http://localhost:8080` (порт можно изменить через `WEB_PORT`). На вкладке «Семья» добавьте члена семьи и его Telegram chat ID. **Бот обрабатывает сообщения только от добавленных пользователей.**
