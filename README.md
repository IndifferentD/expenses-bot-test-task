# Тестовое задание: ИИ-бот учёта расходов

Решение [тестового задания](./task.md) на позицию Go-разработчика.

## Запуск

```sh
git clone https://github.com/IndifferentD/expenses-tg-bot-test-task
cd expenses-tg-bot-test-task
cp .env.example .env
```

Переменные в `.env`:

- `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD` — база данных и учётные данные PostgreSQL.
- `WEB_PORT` — порт веб-панели, по умолчанию `8080`.
- `LOG_LEVEL` — уровень логов Go-сервисов, по умолчанию `INFO`.
- `TELEGRAM_BOT_TOKEN` — токен бота из BotFather; обязателен.
- `TELEGRAM_PROXY_URL` — необязательный прокси для запросов бота к Telegram.
- `OPENAI_API_KEY` — ключ LLM-провайдера для обычного запуска; при запуске локальной модели Compose задаёт служебное значение `local`.
- `OPENAI_BASE_URL` — адрес OpenAI-совместимого API, по умолчанию официальный API OpenAI.
- `OPENAI_MODEL` — имя модели для запросов, по умолчанию `gpt-6-luna`.
- `OPENAI_PROXY_URL` — необязательный прокси только для запросов к LLM.
- `TIMEZONE` — часовой пояс для относительных дат в сообщениях, по умолчанию `Europe/Moscow`.

Бот отправляет запросы в OpenAI-совместимый `POST /v1/responses` и ожидает вызовы функций (`function_call`). Укажите `OPENAI_BASE_URL` и `OPENAI_MODEL`, если используете другой сервер или модель. Одной поддержки Chat Completions недостаточно; модель должна уметь вызывать переданные ей инструменты. Для локального сервера без авторизации всё равно задайте непустой `OPENAI_API_KEY` — сейчас бот требует его при запуске.

Затем запустите:

```sh
docker compose up --build
```

Откройте `http://localhost:8080/users` (порт можно изменить через `WEB_PORT`). На вкладке «Семья» добавьте члена семьи и его Telegram user ID. **Бот обрабатывает сообщения только от добавленных пользователей.**

## Локальная модель

Чтобы попробовать Qwen3-14B на NVIDIA GPU, установите [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html) и запустите:

```sh
docker compose -f compose.yaml -f compose.local-llm.yaml up --build
```

Этот вариант направляет бота на локальный `llama.cpp` вместо сервера из `OPENAI_BASE_URL`; ключ OpenAI не нужен. При первом запуске скачивается GGUF-модель Q4_K_M размером около 9 ГБ. Она сохраняется в Docker volume для следующих запусков. Бот стартует после загрузки модели; затем ему можно отправить сообщение с тратой.
