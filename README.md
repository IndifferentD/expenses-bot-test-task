# Тестовое задание: ИИ-бот учёта расходов

Решение [тестового задания](./task.md) на позицию Go-разработчика.

## Запуск

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

Далее:

```sh
docker compose up
```

Откройте `http://localhost:8080/users` (порт можно изменить через `WEB_PORT`). На вкладке «Семья» добавьте члена семьи и его Telegram user ID. **Бот обрабатывает сообщения только от добавленных пользователей.** После этого можно взаимодействовать с ботом в личке.

Список трат доступен по `http://localhost:8080/expenses`

*Опционально* можно наполнить БД тестовыми тратами:

```sh
docker compose exec api seed-expenses
```

Команда создаст 100 случайных трат за последние шесть месяцев для существующих пользователей. Категории и описания берутся из фиксированного набора, суммы — из заданных диапазонов.

## Локальная модель

В `compose.local-llm.yaml` настроен `llama.cpp` с моделью `Qwen3-14B` в формате Q4_K_M. Этот вариант использует CUDA-образ и запрашивает NVIDIA GPU через Docker Compose. Docker должен предоставлять видеокарту контейнеру; для Linux Docker Engine нужны драйвер NVIDIA и настроенный [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html). Для запуска:

```sh
docker compose -f compose.yaml -f compose.local-llm.yaml up --build
```

Этот вариант направляет бота на локальный `llama.cpp` вместо сервера из `OPENAI_BASE_URL`. При первом запуске скачивается GGUF-файл размером около 9 ГБ. Он сохраняется в Docker volume для следующих запусков. Бот стартует после загрузки модели; затем ему можно отправить сообщение с тратой.

Сама модель не требует NVIDIA: `llama.cpp` поддерживает [CPU и другие GPU-бэкенды](https://github.com/ggml-org/llama.cpp#supported-backends). Для них потребуется другой способ запуска сервера; готовый Compose-файл настроен только для CUDA/NVIDIA.

Я пробовал бота с `gpt-6-luna`, `Qwen3.5-9B` и `Qwen3-14B`. Младшая `Qwen3.5-9B` заметно чаще ошибалась в выборе инструмента и его аргументов; в Compose оставлена `Qwen3-14B`.
