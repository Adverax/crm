# ADR-0017: Auth Module — JWT-аутентификация

**Статус:** Принято
**Дата:** 2026-02-13
**Участники:** @roman_myakotin

## Контекст

Phase 0–4 завершены: metadata engine, security (OLS/FLS/RLS), SOQL, DML. Все запросы аутентифицируются через DevAuth middleware (заголовок `X-Dev-User-Id`), что подходит только для разработки. Phase 5 заменяет DevAuth на полноценную JWT-аутентификацию.

Требования:
1. Аутентификация через username + password
2. Stateless access tokens (JWT) для API-запросов
3. Refresh tokens для продления сессии без повторного ввода пароля
4. Password reset flow (забыли пароль)
5. Rate limiting на login endpoint для защиты от brute-force
6. Совместимость с существующим security engine (UserContext → SOQL/DML)

## Рассмотренные варианты

### Вариант A — Session-based auth (server-side sessions)

Классические HTTP-сессии: session ID в cookie, данные сессии в Redis/DB.

**Плюсы:**
- Простая инвалидация (удалить сессию на сервере)
- Нет проблемы с размером токена

**Минусы:**
- Stateful — требует shared storage для сессий
- Каждый запрос → lookup в Redis/DB
- Не подходит для API-first платформы (mobile, integrations)
- CSRF-защита для cookie-based auth

### Вариант B — JWT access + refresh tokens (выбран)

Short-lived JWT access token (15 min) + long-lived refresh token (7 days) в БД.

**Плюсы:**
- Stateless access tokens — нет lookup на каждый запрос
- Все нужные данные (UserID, ProfileID, RoleID) в claims — совместимость с UserContext
- API-first: удобно для SPA, mobile, integrations
- Стандартный подход для enterprise API

**Минусы:**
- Нельзя мгновенно инвалидировать access token (ждём 15 min expiry)
- Refresh token требует хранения в БД

### Вариант C — OAuth 2.0 / OIDC

Полноценный OAuth 2.0 authorization server.

**Плюсы:**
- Стандарт индустрии
- Поддержка SSO, federated identity

**Минусы:**
- Огромная сложность для MVP
- Требует authorization server (Keycloak, Hydra)
- Overkill для single-tenant self-hosted CRM

## Решение

**Выбран Вариант B — JWT access + refresh tokens.**

### Детали реализации

#### Токены
- **Access token**: JWT, подписан HMAC-SHA256, TTL = 15 минут
- **Refresh token**: crypto/rand 32 bytes → hex string, TTL = 7 дней
- **Refresh token storage**: SHA-256 hash в таблице `iam.refresh_tokens`. Raw token знает только клиент.
- **Token rotation**: при refresh старый token удаляется, выдаётся новый (предотвращает replay)

#### JWT Claims
```json
{
  "sub": "<user_id>",
  "pid": "<profile_id>",
  "rid": "<role_id>",
  "exp": 1234567890,
  "iat": 1234567890
}
```

Claims содержат все поля для `security.UserContext` — middleware создаёт UserContext из JWT без обращения к БД.

#### Пароли
- **Хеширование**: bcrypt, cost = 12
- **Хранение**: колонка `password_hash VARCHAR(255)` в `iam.users`
- **Пустой hash** (`''`) означает "пароль не задан" — login отклоняется

#### Регистрация
- **Admin-only**: администратор создаёт пользователя через существующий CRUD (`POST /admin/security/users`), затем задаёт пароль через `PUT /admin/security/users/:id/password`
- **Self-registration отсутствует** — нетипично для enterprise CRM
- **Начальный пароль admin**: env variable `ADMIN_INITIAL_PASSWORD`, устанавливается при первом запуске

#### Password Reset
- Таблица `iam.password_reset_tokens`: одноразовый token, TTL = 1 час
- `POST /auth/forgot-password` — всегда возвращает 200 (не раскрывает существование email)
- `POST /auth/reset-password` — проверяет token, устанавливает новый пароль, инвалидирует все refresh tokens (force re-login)
- Email sender — interface. Для dev: console-реализация (логирует URL). Для production: SMTP-реализация (подключается позже)

#### Rate Limiting
- In-memory sliding window per IP
- 5 попыток за 15 минут
- Достаточно для single-tenant (ADR-0016)
- При необходимости заменяется на Redis-based

#### Blacklisting access tokens
- **Не реализуется.** Access token short-lived (15 min), истекает естественно
- При logout удаляется только refresh token
- При password reset удаляются все refresh tokens пользователя

### Endpoints

| Method | Path | Auth | Описание |
|--------|------|------|----------|
| POST | `/api/v1/auth/login` | Нет | Вход: username + password → token pair |
| POST | `/api/v1/auth/refresh` | Нет | Обновление: refresh_token → новый token pair |
| POST | `/api/v1/auth/forgot-password` | Нет | Запрос сброса пароля по email |
| POST | `/api/v1/auth/reset-password` | Нет | Сброс пароля по token |
| POST | `/api/v1/auth/logout` | JWT | Выход: удаляет refresh token |
| GET | `/api/v1/auth/me` | JWT | Текущий пользователь |
| PUT | `/api/v1/admin/security/users/:id/password` | JWT | Установка пароля (admin) |

### Совместимость с security engine

JWT middleware создаёт `security.UserContext{UserID, ProfileID, RoleID}` из claims и устанавливает его в Gin + standard context — точно как DevAuth. SOQL/DML engines не требуют изменений.

## Последствия

- DevAuth middleware заменяется на JWTAuth. DevAuth может быть сохранён для тестов (`MODE=dev`)
- Все существующие endpoints получают JWT-защиту
- Frontend получает login page, token management, route guards
- Email infrastructure (SMTP) — заглушка для MVP, реальная реализация при необходимости
- OAuth/OIDC (SSO) — future work, в `ee/` (ADR-0014)

## Связанные решения

- [ADR-0009: Security architecture](0009-security-architecture-overview.md) — 3-слойная безопасность, UserContext
- [ADR-0014: Open Core](0014-licensing-and-business-model.md) — SSO в ee/
- [ADR-0016: Single-tenant](0016-single-tenant-architecture.md) — in-memory rate limiting достаточен