# RedColarTest

Ядро сервиса опасных зон. Клиент отправляет координаты, сервис синхронно
возвращает список ближайших опасных зон и асинхронно отправляет вебхук
на новостной портал при наличии опасностей.

## Запуск (Docker)

```
docker compose up --build
```

Сервис будет доступен на `http://localhost:8080`.

## Переменные окружения

- `DATABASE_URL` — строка подключения Postgres.
- `OPERATOR_API_KEY` — ключ оператора (CRUD и статистика).
- `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB` — Redis для очереди и кэша.
- `WEBHOOK_URL` — URL вебхука (например, `http://<ngrok>/webhook`).
- `STATS_TIME_WINDOW_MINUTES` — окно статистики.
- `CACHE_INCIDENTS_TTL_SECONDS` — TTL кэша активных инцидентов.
- `WEBHOOK_MAX_RETRIES`, `WEBHOOK_RETRY_BASE_SECONDS` — retry для вебхуков.

## Миграции

SQL лежит в `migrations/`. Можно применить, например, `golang-migrate`:

```
migrate -path ./migrations -database "$DATABASE_URL" up
```

## API

### POST `/api/v1/location/check` (публичный)

```
curl -X POST http://localhost:8080/api/v1/location/check \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"u-123","latitude":55.7522,"longitude":37.6156}'
```

Ответ:
```
{"dangerous":true,"incidents":[{"id":1,"title":"...","distance_m":42.1}]}
```

### CRUD инцидентов (оператор, `x-api-key`)

```
curl -X POST http://localhost:8080/api/v1/incidents \
  -H 'Content-Type: application/json' \
  -H 'x-api-key: dev-operator-key' \
  -d '{"title":"Test","latitude":55.75,"longitude":37.61,"danger_radius_m":100}'
```

```
curl -X GET 'http://localhost:8080/api/v1/incidents?page=1&page_size=20' \
  -H 'x-api-key: dev-operator-key'
```

```
curl -X PUT http://localhost:8080/api/v1/incidents/1 \
  -H 'Content-Type: application/json' \
  -H 'x-api-key: dev-operator-key' \
  -d '{"title":"Updated","latitude":55.75,"longitude":37.61,"danger_radius_m":150,"is_active":true}'
```

```
curl -X DELETE http://localhost:8080/api/v1/incidents/1 \
  -H 'x-api-key: dev-operator-key'
```

### Статистика

```
curl -X GET http://localhost:8080/api/v1/incidents/stats \
  -H 'x-api-key: dev-operator-key'
```

### Health-check

```
curl http://localhost:8080/api/v1/system/health
```

## Вебхуки и ngrok

1. Поднять тестовый сервер (заглушку) на `:9090`:
   ```
   python3 - <<'PY'
import json
from http.server import BaseHTTPRequestHandler, HTTPServer

class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        length = int(self.headers.get("content-length", "0"))
        body = self.rfile.read(length).decode()
        print("webhook:", body)
        self.send_response(200)
        self.end_headers()

HTTPServer(("", 9090), Handler).serve_forever()
PY
   ```
2. Пробросить порт:
   ```
   ngrok http 9090
   ```
3. Прописать `WEBHOOK_URL` в `.env`/docker-compose.

