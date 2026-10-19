# Тестовое задание: ИИ-бот учёта расходов

Решение [тестового задания](./task.md) на позицию Go-разработчика.

## Запуск

Склонировать репозиторий и подготовить `.env`:

```sh
git clone https://github.com/IndifferentD/expenses-bot-test-task.git
cd expenses-bot-test-task
cp .env.example .env
```

Указать в `.env`:

- `TELEGRAM_BOT_TOKEN` — токен бота.
- `TELEGRAM_PROXY_URL` — (при необходимости) прокси для запросов бота к Telegram (используется long polling).
- `OPENAI_API_KEY` — ключ LLM-провайдера, если сервер требует авторизацию; для локального сервера можно оставить пустым.
- `OPENAI_BASE_URL` — адрес OpenAI-совместимого API.
- `OPENAI_MODEL` — имя модели для запросов.
- `OPENAI_PROXY_URL` — (при необходимости) прокси для запросов к LLM.

Боту нужен OpenAI-совместимый Responses API и модель с поддержкой вызовов инструментов (`function_call`).

Запустить сервисы:

```sh
docker compose up
```

Открыть `http://localhost:8080/users` (порт можно изменить через `WEB_PORT`). На вкладке «Семья» добавить члена семьи и его Telegram user ID. **Бот обрабатывает сообщения только от добавленных пользователей.** После этого можно отправлять боту сообщение в личном чате.

Просмотреть список трат можно по адресу `http://localhost:8080/expenses`.

При необходимости наполнить БД тестовыми тратами:

```sh
docker compose exec api seed-expenses
```

Команда создаст 100 случайных трат за последние шесть месяцев для существующих пользователей. Категории и описания берутся из фиксированного набора, суммы — из заданных диапазонов.

## Локальная модель

`compose.local-llm.yaml` запускает бота с локальной `Qwen3-14B` в формате GGUF `Q4_K_M` через `llama.cpp`. Для этой CUDA-конфигурации нужен доступ контейнера к NVIDIA GPU; в Linux Docker Engine — драйвер и [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html). У меня модель занимала около 10 ГБ видеопамяти. Запустить сервисы:

```sh
docker compose -f compose.yaml -f compose.local-llm.yaml up --build
```

При первом запуске скачивается GGUF-файл размером около 9 ГБ. Он сохраняется в Docker volume для следующих запусков. Бот стартует после загрузки модели. Затем можно отправлять ему сообщение с тратой.

Сама модель не требует NVIDIA: `llama.cpp` поддерживает [CPU и другие GPU-бэкенды](https://github.com/ggml-org/llama.cpp#supported-backends). Для них потребуется другой способ запуска сервера.
___
Я пробовал бота с `gpt-6-luna` через [ProxyAPI](https://proxyapi.ru/), а `Qwen3.5-9B` и `Qwen3-14B` запускал локально через `llama.cpp`. `Qwen3.5-9B` заметно чаще ошибалась в выборе инструмента и его аргументов; в Compose оставлена `Qwen3-14B`.
