# ADR-0028: Named Credentials — безопасное хранение секретов для интеграций

**Статус:** Принято
**Дата:** 2026-02-15
**Участники:** @roman_myakotin

## Контекст

### Проблема: HTTP-интеграции требуют аутентификации

Procedure Engine (ADR-0024) предоставляет command type `integration.http` для вызова внешних API. Каждый HTTP-вызов требует аутентификации — API-ключ, логин/пароль, OAuth2-токен.

Текущая абстракция в ADR-0024 — `$.secrets` namespace:

```json
{
  "type": "integration.http",
  "method": "POST",
  "url": "https://api.payment.com/charge",
  "headers": {
    "Authorization": "Bearer $.secrets.stripe_key"
  },
  "body": { "amount": "$.input.amount" }
}
```

Это создаёт **четыре проблемы**:

| Проблема | Описание |
|----------|----------|
| **Где хранятся `$.secrets`?** | ADR-0024 не определяет storage. Env vars? DB? Config file? |
| **Как ротировать?** | Обновление секрета требует знания, где он хранится и кто использует |
| **Как аудировать?** | Нет записи, какая Procedure когда использовала какой секрет |
| **Как защитить от SSRF?** | URL собирается из строк — Procedure может вызвать `http://localhost:5432` |

### Salesforce Named Credentials

Salesforce решает эту проблему через **Named Credentials** — именованные учётные записи, привязанные к endpoint:

```
Named Credential = endpoint URL + auth method + secrets
```

Procedure ссылается на credential по имени, не зная деталей аутентификации. При ротации секрета обновляется только credential — все Procedures продолжают работать.

### Почему не env vars

| Аспект | Environment Variables | Named Credentials |
|--------|----------------------|-------------------|
| Хранение | Файл / CI secrets | БД (зашифровано) |
| Управление | DevOps + деплой | Администратор через UI |
| Ротация | Передеплой | Обновление через API без деплоя |
| Аудит | Невозможен | Полный лог использования |
| SSRF | Нет защиты | base_url ограничивает target |
| OAuth2 | Ручное обновление токенов | Автоматический refresh |
| Видимость | Доступны всему приложению | Только через Credential Service |

## Рассмотренные варианты

### Вариант A — Named Credentials в БД (выбран)

Именованные credentials хранятся в `metadata.credentials`, секреты зашифрованы AES-256-GCM. Procedure ссылается на credential по коду. Credential Service расшифровывает секреты в runtime и формирует auth header.

**Плюсы:**
- Централизованное управление: один источник правды для всех секретов
- Шифрование at rest: AES-256-GCM с уникальным nonce
- SSRF protection: base_url ограничивает допустимые хосты
- Ротация без деплоя: администратор обновляет через UI
- OAuth2 auto-refresh: платформа сама обновляет истёкшие токены
- Аудит: каждое использование логируется
- Интеграция с Procedure: `credential` field вместо inline secrets

**Минусы:**
- Master key management: нужен отдельный секрет для шифрования (ENV)
- Дополнительная сложность: encryption service, token cache, audit log
- Single point of failure: компрометация master key = все секреты

### Вариант B — Environment Variables

Секреты хранятся в env vars, доступны через `$.env.STRIPE_KEY` в Procedure.

**Плюсы:**
- Простота: стандартный подход, нет нового кода
- Совместимость: работает с любым CI/CD

**Минусы:**
- Нет UI: администратор не может управлять секретами
- Нет ротации без деплоя: изменение env var = перезапуск
- Нет аудита: невозможно отследить использование
- Нет SSRF protection: URL полностью контролируется Procedure
- Нет OAuth2: ручное обновление токенов
- Глобальная видимость: env vars доступны всему приложению

### Вариант C — External Vault (HashiCorp Vault / AWS Secrets Manager)

Интеграция с внешним secret manager.

**Плюсы:**
- Enterprise-grade: battle-tested решения
- Key rotation: встроенная ротация
- Fine-grained ACL: per-secret permissions

**Минусы:**
- Внешняя зависимость: Vault нужно деплоить, настраивать, поддерживать
- Latency: network call на каждый доступ к секрету
- Сложность для self-hosted: наша целевая модель — single-tenant self-hosted (ADR-0016)
- Overkill для MVP: десятки-сотни секретов, не тысячи

### Вариант D — Inline secrets в Procedure JSON

Секреты хранятся прямо в JSON definition Procedure.

**Плюсы:**
- Нет абстракции: всё в одном месте

**Минусы:**
- Секреты видны всем, кто может читать Procedure definitions
- Логируются как часть Procedure (в логах, аудите, дампах)
- Дублирование: один API key в N Procedures
- Нет ротации: обновление = редактирование всех Procedures

## Решение

**Выбран вариант A: Named Credentials в БД с AES-256-GCM шифрованием.**

### Определение Named Credential

Named Credential — именованная учётная запись, которая инкапсулирует:
- **Endpoint** (base URL) — куда можно обращаться
- **Auth method** (тип аутентификации) — как аутентифицироваться
- **Secrets** (зашифрованные данные) — чем аутентифицироваться

```json
{
  "code": "stripe_api",
  "name": "Stripe API",
  "description": "Production Stripe account",
  "type": "api_key",
  "base_url": "https://api.stripe.com",
  "auth": {
    "placement": "header",
    "header_name": "Authorization",
    "header_value": "Bearer sk_live_xxx"
  }
}
```

### Типы Credentials

| Тип | Описание | Auth flow |
|-----|----------|-----------|
| `api_key` | Статический токен | Header / Query param |
| `basic` | Username + Password | `Authorization: Basic base64(user:pass)` |
| `oauth2_client` | Client Credentials Grant | Автоматический token fetch + refresh |

#### API Key

```json
{
  "code": "sendgrid_api",
  "type": "api_key",
  "base_url": "https://api.sendgrid.com",
  "auth": {
    "placement": "header",
    "header_name": "Authorization",
    "header_value": "Bearer SG.xxx"
  }
}
```

Credential Service формирует заголовок: `Authorization: Bearer SG.xxx`.

#### Basic Auth

```json
{
  "code": "legacy_erp",
  "type": "basic",
  "base_url": "https://erp.company.com",
  "auth": {
    "username": "api_user",
    "password": "secret123"
  }
}
```

Credential Service формирует: `Authorization: Basic YXBpX3VzZXI6c2VjcmV0MTIz`.

#### OAuth2 Client Credentials

```json
{
  "code": "salesforce_api",
  "type": "oauth2_client",
  "base_url": "https://company.my.salesforce.com",
  "auth": {
    "token_url": "https://login.salesforce.com/services/oauth2/token",
    "client_id": "3MVG9...",
    "client_secret": "xxx",
    "scope": "api refresh_token"
  }
}
```

Credential Service автоматически:
1. Запрашивает access_token через Client Credentials Grant
2. Кэширует до `expires_at` в `metadata.credential_tokens`
3. Обновляет при истечении (прозрачно для Procedure)

### Использование в Procedure

`integration.http` command получает поле `credential` вместо inline auth:

```json
{
  "type": "integration.http",
  "credential": "stripe_api",
  "method": "POST",
  "path": "/v1/charges",
  "body": {
    "amount": "$.input.amount",
    "currency": "usd"
  },
  "as": "charge"
}
```

**Что происходит при выполнении:**

```
integration.http command
    │
    ├── 1. Resolve credential by code ("stripe_api")
    │       → metadata.credentials WHERE code = 'stripe_api'
    │
    ├── 2. Decrypt auth data
    │       → AES-256-GCM decrypt (master_key, nonce, ciphertext)
    │
    ├── 3. Build auth header
    │       → For api_key: header from auth config
    │       → For basic: Authorization: Basic base64(user:pass)
    │       → For oauth2: resolve/refresh token, Authorization: Bearer <token>
    │
    ├── 4. Validate URL (SSRF protection)
    │       → base_url + path → HTTPS only, host match, no internal IPs
    │
    ├── 5. Execute HTTP request
    │       → request.Header.Set(auth_header)
    │       → client.Do(request) [timeout: 10s]
    │
    └── 6. Log usage (without secrets)
            → credential_id, procedure_code, url, status, duration
```

**`$.secrets` заменяется на `credential`:** ADR-0024 определял `$.secrets` как runtime namespace для секретов. Named Credentials делают этот namespace ненужным — Procedure не обращается к секретам напрямую, а ссылается на credential по коду. Credential Service инжектирует auth прозрачно.

### Шифрование

```
                    ┌─────────────┐
                    │ Master Key  │
                    │ (from ENV)  │
                    └──────┬──────┘
                           │
    ┌──────────────────────┼──────────────────────┐
    │                      │                      │
    │  ┌──────────┐   ┌────┴────┐   ┌──────────┐ │
    │  │  Nonce   │   │ AES-256 │   │Auth Data │ │
    │  │ (random) │──▶│   GCM   │◀──│(plaintext)│ │
    │  └──────────┘   └────┬────┘   └──────────┘ │
    │                      │                      │
    │                      ▼                      │
    │         ┌───────────────────────┐           │
    │         │  auth_data_encrypted  │           │
    │         │  + auth_data_nonce    │           │
    │         │  (stored in DB)       │           │
    │         └───────────────────────┘           │
    └─────────────────────────────────────────────┘
```

**Почему AES-256-GCM:**
- Authenticated encryption: integrity + confidentiality в одном алгоритме
- Индустриальный стандарт (NIST рекомендация)
- Hardware acceleration (AES-NI) на всех современных CPU
- Уникальный nonce per-record предотвращает replay attacks

**Master Key:**

| Среда | Хранение |
|-------|----------|
| Development | `.env` (gitignored) |
| Production | Environment variable `CREDENTIAL_ENCRYPTION_KEY` |
| Enterprise (future) | HashiCorp Vault / KMS через интерфейс |

```bash
# Генерация (32 bytes = 256 bits)
openssl rand -base64 32
```

### SSRF Protection

Каждый credential имеет обязательный `base_url`. При выполнении `integration.http`:

1. Полный URL = `base_url` + `path`
2. Только HTTPS (HTTP запрещён)
3. Host полного URL должен совпадать с host из `base_url`
4. Internal IP заблокированы (127.0.0.0/8, 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, ::1)

```
credential.base_url = "https://api.stripe.com"

path: "/v1/charges"        → ✅ https://api.stripe.com/v1/charges
path: "/../internal"       → ❌ host mismatch (path traversal)
url: "http://localhost:5432" → ❌ HTTP + internal IP
```

### Хранение

Таблица `metadata.credentials`:

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | Уникальный ID |
| code | VARCHAR(100) UNIQUE | Код для ссылки из Procedure |
| name | VARCHAR(255) | Человекочитаемое имя |
| description | TEXT | Описание назначения |
| type | VARCHAR(20) | `api_key` / `basic` / `oauth2_client` |
| base_url | VARCHAR(500) NOT NULL | Base URL (SSRF protection) |
| auth_data_encrypted | BYTEA NOT NULL | Зашифрованные auth данные (AES-256-GCM) |
| auth_data_nonce | BYTEA NOT NULL | Уникальный nonce |
| is_active | BOOLEAN DEFAULT true | Активна ли credential (деактивация без удаления) |
| created_at | TIMESTAMPTZ | Время создания |
| updated_at | TIMESTAMPTZ | Время обновления |

Таблица `metadata.credential_tokens` (OAuth2 token cache):

| Колонка | Тип | Описание |
|---------|-----|----------|
| credential_id | UUID PK FK→credentials | 1:1 с credential |
| access_token_encrypted | BYTEA | Зашифрованный access token |
| access_token_nonce | BYTEA | Nonce для access token |
| token_type | VARCHAR(50) DEFAULT 'Bearer' | Тип токена |
| expires_at | TIMESTAMPTZ | Время истечения |
| created_at | TIMESTAMPTZ | Время получения |
| updated_at | TIMESTAMPTZ | Время обновления |

Таблица `metadata.credential_usage_log` (аудит использования):

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | Уникальный ID |
| credential_id | UUID FK→credentials | Какой credential использовался |
| procedure_code | VARCHAR(100) | Какая Procedure инициировала |
| request_url | VARCHAR(500) | URL запроса (без query params) |
| response_status | INT | HTTP-статус ответа |
| success | BOOLEAN | Успешность |
| error_message | TEXT | Сообщение об ошибке (если есть) |
| duration_ms | INT | Длительность запроса |
| created_at | TIMESTAMPTZ | Время использования |
| user_id | UUID FK→users | Кто инициировал Procedure |

**Примечание о `is_active`:** В отличие от soft delete бизнес-записей (которого у нас нет, ADR-0003), `is_active` для credentials — это **механизм безопасности**: временная деактивация при подозрении на компрометацию, без потери конфигурации. Деактивированный credential блокирует все Procedures, которые его используют.

### API

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/api/v1/admin/credentials` | Список credentials (без секретов) |
| POST | `/api/v1/admin/credentials` | Создать credential |
| GET | `/api/v1/admin/credentials/:id` | Получить credential (без секретов) |
| PUT | `/api/v1/admin/credentials/:id` | Обновить credential |
| DELETE | `/api/v1/admin/credentials/:id` | Удалить (409 если используется в Procedures) |
| POST | `/api/v1/admin/credentials/:id/test` | Тест подключения (GET на base_url) |
| GET | `/api/v1/admin/credentials/:id/usage` | Лог использования |
| POST | `/api/v1/admin/credentials/:id/deactivate` | Деактивировать |
| POST | `/api/v1/admin/credentials/:id/activate` | Активировать |

**GET/PUT response — auth_data маскируется:**

```json
{
  "id": "uuid",
  "code": "stripe_api",
  "name": "Stripe API",
  "type": "api_key",
  "base_url": "https://api.stripe.com",
  "is_active": true,
  "auth_masked": {
    "placement": "header",
    "header_name": "Authorization",
    "header_value": "Bearer sk_l***xxx"
  },
  "created_at": "2026-02-15T10:00:00Z"
}
```

Секреты **никогда** не возвращаются в API responses целиком. При обновлении auth данных — передаётся новый plaintext, шифруется при сохранении.

### Безопасность

| Угроза | Защита |
|--------|--------|
| Утечка из БД (SQL dump) | AES-256-GCM шифрование; без master key данные бесполезны |
| Утечка master key | Хранение в ENV (не в коде, не в БД); single point — осознанный trade-off |
| SSRF (Server-Side Request Forgery) | base_url constraint; host match; internal IP blocklist; HTTPS only |
| Логирование секретов | Auth data **никогда** не логируется; usage log хранит только URL + status |
| Несанкционированный доступ к API | Admin-only endpoints (middleware); будущее: OLS на credentials |
| Credential в неактивных Procedure | Валидация при сохранении Procedure: credential должен существовать |
| Удаление используемого credential | 409 Conflict: dependency check (где используется) |
| Компрометация одного credential | Деактивация (`is_active = false`) блокирует все использования мгновенно |

### Ограничения

| Ограничение | Обоснование |
|-------------|-------------|
| Только HTTPS | Безопасность (by design) |
| Нет mTLS | Сложность; добавляется в будущем |
| Нет automatic key rotation | Требует интеграции с каждым провайдером; ручное обновление |
| Single master key | Простота; key rotation через re-encrypt всех записей |
| Нет per-user credentials | Все Procedures используют один credential; per-user OAuth2 — будущее |
| Max 100 credentials | Достаточно для production; предотвращает раздувание |

### Валидация при сохранении Procedure

При сохранении Procedure, если command type = `integration.http`:
1. Поле `credential` обязательно (inline URL/auth запрещены)
2. Credential с указанным кодом должен существовать
3. Credential должен быть `is_active = true`

```json
// ✅ Правильно
{
  "type": "integration.http",
  "credential": "stripe_api",
  "method": "POST",
  "path": "/v1/charges",
  "body": { "amount": "$.input.amount" }
}

// ❌ Запрещено: inline URL без credential
{
  "type": "integration.http",
  "method": "POST",
  "url": "https://api.stripe.com/v1/charges",
  "headers": { "Authorization": "Bearer sk_xxx" },
  "body": { "amount": "$.input.amount" }
}
```

### Constructor UI

Admin-страница для управления credentials:

1. **Список credentials**: code, name, type, base_url, is_active, last_used_at
2. **Создание/редактирование**: форма с полями по типу credential (api_key/basic/oauth2)
3. **Test connection**: кнопка для проверки подключения (GET base_url с auth)
4. **Usage log**: таблица использований с фильтрами (date range, procedure, status)
5. **Deactivate/Activate**: toggle с подтверждением (показывает affected Procedures)

## Последствия

### Позитивные

- **Централизация** — все секреты для интеграций в одном месте
- **Шифрование at rest** — AES-256-GCM, secrets не читаемы из БД без master key
- **SSRF protection** — base_url + host match + internal IP blocklist
- **Ротация без деплоя** — обновление через UI/API, Procedures не меняются
- **OAuth2 auto-refresh** — платформа сама обновляет токены
- **Аудит** — полный лог: кто, когда, какой credential, с каким результатом
- **DRY** — один credential используется в N Procedures
- **Интеграция с Procedure** — `credential` field вместо `$.secrets`; валидация при сохранении
- **Мгновенная деактивация** — `is_active = false` блокирует все использования

### Негативные

- **Master key** — единая точка; компрометация = все секреты. Mitigation: ENV + access control
- **Дополнительная сложность** — encryption service, token cache, usage log
- **Нет per-user OAuth2** — все пользователи используют один credential; Connected Apps — будущее
- **Manual rotation** — нет автоматической ротации API keys (только OAuth2 tokens auto-refresh)

## Связанные ADR

- **ADR-0024** — Procedure Engine: `integration.http` command использует `credential` field. Named Credentials заменяют `$.secrets` namespace
- **ADR-0025** — Scenario Engine: steps могут вызывать Procedures с `integration.http`, которые используют Named Credentials
- **ADR-0016** — Single-tenant: master key per instance; нет multi-tenant key management
- **ADR-0009** — Security: Admin-only доступ к credentials API
