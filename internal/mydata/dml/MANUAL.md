# DML: Полное руководство

## Введение

**DML (Data Manipulation Language)** — язык для модификации данных в CRM-системе. Поддерживает четыре основные операции: INSERT, UPDATE, DELETE и UPSERT.

### Для кого этот язык?

DML используют:
- Разработчики для программного доступа к данным
- Интеграторы при обмене данными с внешними системами
- Администраторы для массовых операций с данными

### Ключевые отличия от SQL

DML работает с **объектами и записями**, а не с таблицами и строками. Имена объектов и полей используют CamelCase (как в API), а не snake_case (как в SQL).

---

## INSERT: Вставка записей

### Базовый синтаксис

```sql
INSERT INTO Объект (Поле1, Поле2, ...)
VALUES (значение1, значение2, ...),
       (значение3, значение4, ...)
```

### Примеры

**Вставка одной записи:**

```sql
INSERT INTO Account (Name, Industry, AnnualRevenue)
VALUES ('Acme Corporation', 'Technology', 1000000)
```

**Batch-вставка нескольких записей:**

```sql
INSERT INTO Contact (FirstName, LastName, Email, AccountId)
VALUES
    ('John', 'Smith', 'john@acme.com', 'acc-001'),
    ('Jane', 'Doe', 'jane@acme.com', 'acc-001'),
    ('Bob', 'Wilson', 'bob@globex.com', 'acc-002')
```

**Вставка с NULL значениями:**

```sql
INSERT INTO Contact (FirstName, LastName, Email, Phone)
VALUES ('John', 'Smith', 'john@example.com', NULL)
```

### Ограничения

- Все обязательные поля должны быть указаны
- Нельзя записывать в read-only поля (Id, CreatedDate и т.д.)
- Количество значений должно совпадать с количеством полей
- Максимум 10,000 строк в одном INSERT (настраивается)

---

## UPDATE: Обновление записей

### Базовый синтаксис

```sql
UPDATE Объект
SET Поле1 = значение1, Поле2 = значение2, ...
[WHERE условие]
```

### Примеры

**Обновление одной записи:**

```sql
UPDATE Contact
SET Status = 'Active', Phone = '+1-555-0100'
WHERE Email = 'john@example.com'
```

**Обновление нескольких записей:**

```sql
UPDATE Opportunity
SET Stage = 'Closed Won', CloseDate = 2024-01-15
WHERE AccountId = 'acc-001' AND Amount > 100000
```

**Установка NULL:**

```sql
UPDATE Contact
SET Phone = NULL
WHERE Id = 'cnt-001'
```

### Важно

- Если WHERE не указан, будут обновлены **все записи** (по умолчанию разрешено, но можно запретить в конфигурации)
- Нельзя обновлять read-only поля
- Нельзя указывать одно поле дважды в SET

---

## DELETE: Удаление записей

### Базовый синтаксис

```sql
DELETE FROM Объект
WHERE условие
```

### Примеры

**Удаление по ID:**

```sql
DELETE FROM Task
WHERE Id = 'task-001'
```

**Удаление по условию:**

```sql
DELETE FROM Task
WHERE Status = 'Completed' AND CreatedDate < 2023-01-01
```

### Важно

- **WHERE обязателен по умолчанию** — это защита от случайного удаления всех записей
- Для удаления всех записей необходимо явно указать условие, например: `WHERE 1 = 1`

---

## UPSERT: Вставка или обновление

UPSERT (INSERT + UPDATE) вставляет новую запись или обновляет существующую, если найдено совпадение по внешнему идентификатору.

### Базовый синтаксис

```sql
UPSERT Объект (Поле1, Поле2, ...)
VALUES (значение1, значение2, ...),
       (значение3, значение4, ...)
ON ПолеВнешнегоИД
```

### Примеры

**UPSERT с внешним ID:**

```sql
UPSERT Account (ExternalId, Name, Industry)
VALUES
    ('ext-001', 'Acme Corp', 'Technology'),
    ('ext-002', 'Globex Inc', 'Finance')
ON ExternalId
```

**Логика работы:**
1. Если запись с `ExternalId = 'ext-001'` существует — обновить `Name` и `Industry`
2. Если не существует — вставить новую запись

### Требования

- Поле для ON должно быть помечено как `IsExternalId` или `IsUnique`
- Поле ON должно быть включено в список полей
- Поле ON не обновляется при конфликте (остаётся прежним)

---

## Условия WHERE

WHERE поддерживается в UPDATE и DELETE для фильтрации записей.

### Операторы сравнения

| Оператор | Описание | Пример |
|----------|----------|--------|
| `=` | Равно | `Status = 'Active'` |
| `!=`, `<>` | Не равно | `Status != 'Closed'` |
| `<` | Меньше | `Amount < 1000` |
| `>` | Больше | `Amount > 1000` |
| `<=` | Меньше или равно | `Priority <= 3` |
| `>=` | Больше или равно | `Priority >= 1` |

### Логические операторы

```sql
-- AND: все условия должны выполняться
WHERE Status = 'Active' AND Priority = 'High'

-- OR: хотя бы одно условие
WHERE Status = 'Active' OR Status = 'Pending'

-- NOT: отрицание
WHERE NOT Status = 'Closed'

-- Группировка скобками
WHERE (Status = 'Active' OR Status = 'Pending') AND Priority = 'High'
```

### Оператор IN

```sql
-- Проверка вхождения в список
WHERE Status IN ('Active', 'Pending', 'InProgress')

-- Отрицание
WHERE Status NOT IN ('Closed', 'Cancelled')
```

### Оператор LIKE

Поиск по шаблону:

| Символ | Значение |
|--------|----------|
| `%` | Любое количество любых символов |
| `_` | Ровно один символ |

```sql
-- Начинается с "Acme"
WHERE Name LIKE 'Acme%'

-- Содержит "Corp"
WHERE Name LIKE '%Corp%'

-- Заканчивается на ".com"
WHERE Email LIKE '%.com'

-- Отрицание
WHERE Name NOT LIKE '%Test%'
```

### Работа с NULL

```sql
-- Поле не заполнено
WHERE Phone IS NULL

-- Поле заполнено
WHERE Phone IS NOT NULL
```

---

## Функции

DML поддерживает скалярные функции для обработки значений в INSERT, UPDATE и UPSERT.

### Строковые функции

| Функция | Описание | Пример |
|---------|----------|--------|
| `UPPER(str)` | Преобразует в верхний регистр | `UPPER('hello')` → `'HELLO'` |
| `LOWER(str)` | Преобразует в нижний регистр | `LOWER('HELLO')` → `'hello'` |
| `TRIM(str)` | Удаляет пробелы по краям | `TRIM('  hi  ')` → `'hi'` |
| `CONCAT(str1, str2, ...)` | Объединяет строки | `CONCAT('Hello', ' ', 'World')` → `'Hello World'` |
| `LENGTH(str)` или `LEN(str)` | Длина строки | `LENGTH('hello')` → `5` |
| `SUBSTRING(str, start, len)` или `SUBSTR` | Подстрока | `SUBSTRING('hello', 2, 3)` → `'ell'` |

### Математические функции

| Функция | Описание | Пример |
|---------|----------|--------|
| `ABS(num)` | Абсолютное значение | `ABS(5)` → `5` |
| `ROUND(num)` | Округление | `ROUND(3.7)` → `4` |
| `FLOOR(num)` | Округление вниз | `FLOOR(3.7)` → `3` |
| `CEIL(num)` или `CEILING(num)` | Округление вверх | `CEIL(3.2)` → `4` |

### Функции для работы с NULL

| Функция | Описание | Пример |
|---------|----------|--------|
| `COALESCE(val1, val2, ...)` | Первое не-NULL значение | `COALESCE(NULL, 'default')` → `'default'` |
| `NULLIF(val1, val2)` | NULL если значения равны | `NULLIF('', '')` → `NULL` |

### Примеры использования функций

**INSERT с функцией:**

```sql
INSERT INTO Account (Name, Industry)
VALUES (UPPER('acme corporation'), TRIM('  Technology  '))
```

**UPDATE с функцией:**

```sql
UPDATE Contact
SET FirstName = UPPER(FirstName), Email = LOWER(Email)
WHERE Id = 'cnt-001'
```

**Вложенные функции:**

```sql
INSERT INTO Account (Name)
VALUES (UPPER(TRIM('  acme  ')))
```

**COALESCE для значений по умолчанию:**

```sql
INSERT INTO Contact (FirstName, LastName, Status)
VALUES ('John', 'Doe', COALESCE(NULL, 'Active'))
```

**CONCAT для объединения строк:**

```sql
UPDATE Contact
SET FullName = CONCAT(FirstName, ' ', LastName)
WHERE AccountId = 'acc-001'
```

### Ограничения функций

- Функции работают только со значениями (не в WHERE)
- Отрицательные числа не поддерживаются как литералы внутри функций
- Типы аргументов проверяются на совместимость с целевым полем

---

## Типы данных

### Строки

Строки заключаются в одинарные кавычки:

```sql
'Hello, World!'
'It''s a test'   -- Экранирование кавычки: ''
```

### Числа

```sql
42              -- Целое число
3.14            -- Дробное число
-100            -- Отрицательное число
1.5e10          -- Экспоненциальная запись
```

### Булевы значения

```sql
TRUE
FALSE
```

### Даты

Формат ISO 8601:

```sql
-- Только дата (YYYY-MM-DD)
2024-01-15

-- Дата и время (RFC 3339)
2024-01-15T10:30:00Z
2024-01-15T10:30:00+03:00
```

### NULL

```sql
NULL
```

---

## Идентификаторы

### Стандартные идентификаторы

Имена объектов и полей состоят из букв, цифр и подчёркиваний:

```sql
Account
FirstName
custom_field__c
```

### Идентификаторы с пробелами

Используйте двойные кавычки:

```sql
INSERT INTO "Custom Object" ("Field Name", "Another Field")
VALUES ('value1', 'value2')
```

---

## Примеры типичных операций

### Создание контактов для аккаунта

```sql
INSERT INTO Contact (FirstName, LastName, Email, AccountId, Status)
VALUES
    ('John', 'Smith', 'john.smith@acme.com', 'acc-12345', 'Active'),
    ('Jane', 'Doe', 'jane.doe@acme.com', 'acc-12345', 'Active')
```

### Массовое обновление статуса

```sql
UPDATE Task
SET Status = 'Cancelled', CancelReason = 'Project closed'
WHERE ProjectId = 'proj-001' AND Status != 'Completed'
```

### Архивация старых записей

```sql
UPDATE Lead
SET IsArchived = TRUE
WHERE Status = 'Unqualified' AND CreatedDate < 2023-01-01
```

### Удаление тестовых данных

```sql
DELETE FROM Contact
WHERE Email LIKE '%@test.com' AND IsTest = TRUE
```

### Синхронизация из внешней системы

```sql
UPSERT Contact (ExternalSystemId, FirstName, LastName, Email, Phone)
VALUES
    ('CRM-001', 'John', 'Smith', 'john@example.com', '+1-555-0100'),
    ('CRM-002', 'Jane', 'Doe', 'jane@example.com', '+1-555-0200'),
    ('CRM-003', 'Bob', 'Wilson', 'bob@example.com', NULL)
ON ExternalSystemId
```

---

## Ограничения и лимиты

### Количественные лимиты

| Параметр | Значение по умолчанию |
|----------|----------------------|
| Максимум строк в INSERT/UPSERT | 10,000 |
| Максимум полей на строку | Без ограничения |
| Максимум символов в запросе | 100,000 |

### Ограничения безопасности

| Ограничение | По умолчанию |
|-------------|--------------|
| WHERE обязателен для DELETE | Да |
| WHERE обязателен для UPDATE | Нет |

### Чего НЕТ в DML

| Возможность | Статус |
|-------------|--------|
| Подзапросы | Не поддерживается |
| JOIN | Не поддерживается |
| RETURNING (кроме ID) | Не поддерживается |
| Транзакции | Не поддерживается (каждый запрос атомарен) |

---

## Обработка ошибок

### Синтаксические ошибки

Если запрос содержит синтаксическую ошибку, вы получите `ParseError` с указанием позиции:

```
ParseError at line 1, column 15: expected ')', got ','
```

### Ошибки валидации

Ошибки проверки метаданных и типов:

```
ValidationError [UnknownField]: unknown field: Account.InvalidField
ValidationError [TypeMismatch]: type mismatch for Account.Amount: expected float, got string
ValidationError [MissingRequired]: required field Account.Name is missing
```

### Ошибки доступа

Если нет прав на операцию:

```
AccessError: INSERT access denied to object: Account
AccessError: write access denied to field: Contact.SensitiveData
```

### Ошибки лимитов

```
LimitError: MaxBatchSize limit exceeded: 15000 (max: 10000)
LimitError: MaxStatementLength limit exceeded: 150000 (max: 100000)
```

---

## Рекомендации

### Используйте batch-операции

Вместо множества отдельных INSERT объединяйте записи:

```sql
-- Хорошо: один запрос с несколькими записями
INSERT INTO Contact (FirstName, LastName)
VALUES ('John', 'Smith'), ('Jane', 'Doe'), ('Bob', 'Wilson')

-- Плохо: три отдельных запроса
INSERT INTO Contact (FirstName, LastName) VALUES ('John', 'Smith')
INSERT INTO Contact (FirstName, LastName) VALUES ('Jane', 'Doe')
INSERT INTO Contact (FirstName, LastName) VALUES ('Bob', 'Wilson')
```

### Будьте осторожны с UPDATE без WHERE

UPDATE без WHERE обновит **все записи** объекта. Всегда дважды проверяйте условия.

### Используйте UPSERT для синхронизации

При интеграции с внешними системами UPSERT гарантирует идемпотентность — повторный запрос не создаст дубликаты.

### Обрабатывайте ошибки

Всегда проверяйте тип ошибки для правильной обработки:

```go
switch {
case engine.IsParseError(err):
    // Синтаксическая ошибка — показать пользователю
case engine.IsValidationError(err):
    // Ошибка данных — показать детали
case engine.IsAccessError(err):
    // Нет доступа — запросить права
case engine.IsLimitError(err):
    // Превышен лимит — разбить на части
}
```

---

## Связанная документация

- [ARCHITECTURE.md](./ARCHITECTURE.md) — архитектура реализации
