# SOQL: Полное руководство

## Введение

**SOQL (Salesforce Object Query Language)** — специализированный язык запросов, разработанный компанией Salesforce для работы с данными в их CRM-платформе. SOQL похож на SQL, но имеет существенные отличия, адаптированные под объектную модель Salesforce.

### Для кого этот язык?

SOQL используют:
- Разработчики, пишущие код на Apex (внутренний язык Salesforce)
- Администраторы для анализа и отладки данных
- Интеграторы при обмене данными с внешними системами
- Аналитики для построения отчётов

### Ключевое отличие от SQL

SQL работает с таблицами и строками. SOQL работает с **объектами и записями**. Это не просто терминологическая разница — SOQL возвращает иерархические структуры данных, а не плоские таблицы.

---

## Базовый синтаксис

### Структура запроса

```sql
SELECT поля
FROM Объект
[WHERE условия]
[ORDER BY поле [ASC|DESC] [NULLS FIRST|LAST]]
[LIMIT число]
[OFFSET число]
```

### Простейший пример

```sql
SELECT Name, Email
FROM Contact
```

Этот запрос вернёт имена и email-адреса всех контактов в системе.

### Пример с условиями

```sql
SELECT Name, Email, Phone
FROM Contact
WHERE Email != null
  AND CreatedDate > 2024-01-01T00:00:00Z
ORDER BY Name ASC
LIMIT 100
```

---

## Выборка полей

### Простые поля

```sql
SELECT Id, Name, Email, Phone, CreatedDate
FROM Contact
```

### Поля связанных объектов (точечная нотация)

Через точку можно обратиться к полям родительского объекта:

```sql
SELECT LastName, Email, Account.Name, Account.Industry
FROM Contact
```

Здесь `Account` — родительский объект для `Contact`. Можно пройти до **5 уровней вверх**:

```sql
SELECT Name, Account.Owner.Manager.Name
FROM Contact
```

### Нельзя использовать SELECT *

В отличие от SQL, SOQL **не поддерживает** `SELECT *`. Необходимо явно перечислить все нужные поля.

---

## Условия WHERE

### Операторы сравнения

| Оператор | Описание | Пример |
|----------|----------|--------|
| `=` | Равно | `Status = 'Active'` |
| `!=` | Не равно | `Email != null` |
| `<` | Меньше | `Amount < 1000` |
| `>` | Больше | `Amount > 1000` |
| `<=` | Меньше или равно | `Age <= 30` |
| `>=` | Больше или равно | `Age >= 18` |

### Логические операторы

```sql
SELECT Name FROM Account
WHERE Industry = 'Technology'
  AND AnnualRevenue > 1000000
  OR Type = 'Partner'
```

Для управления приоритетом используются скобки:

```sql
SELECT Name FROM Account
WHERE (Industry = 'Technology' OR Industry = 'Finance')
  AND AnnualRevenue > 1000000
```

### Оператор IN

Проверка вхождения в список значений:

```sql
SELECT Name FROM Account
WHERE Industry IN ('Technology', 'Finance', 'Healthcare')
```

```sql
SELECT Name FROM Account
WHERE Industry NOT IN ('Government', 'Non-Profit')
```

### Оператор LIKE

Поиск по шаблону:

| Символ | Значение |
|--------|----------|
| `%` | Любое количество любых символов |
| `_` | Ровно один любой символ |

```sql
-- Все аккаунты, начинающиеся на "Acme"
SELECT Name FROM Account
WHERE Name LIKE 'Acme%'

-- Все аккаунты с "Corp" в названии
SELECT Name FROM Account
WHERE Name LIKE '%Corp%'

-- Названия из ровно 5 символов
SELECT Name FROM Account
WHERE Name LIKE '_____'
```

### Работа с NULL

```sql
-- Контакты без email
SELECT Name FROM Contact
WHERE Email = null

-- Контакты с заполненным телефоном
SELECT Name FROM Contact
WHERE Phone != null
```

---

## Работа с датами

### Форматы дат

SOQL использует формат ISO 8601:

```sql
-- Только дата
WHERE CloseDate > 2024-01-15

-- Дата и время (в UTC)
WHERE CreatedDate > 2024-01-15T10:30:00Z
```

### Литералы дат

Одна из самых удобных возможностей SOQL — встроенные литералы для работы с датами:

| Литерал | Значение |
|---------|----------|
| `TODAY` | Сегодня |
| `YESTERDAY` | Вчера |
| `TOMORROW` | Завтра |
| `THIS_WEEK` | Текущая неделя |
| `LAST_WEEK` | Прошлая неделя |
| `NEXT_WEEK` | Следующая неделя |
| `THIS_MONTH` | Текущий месяц |
| `LAST_MONTH` | Прошлый месяц |
| `NEXT_MONTH` | Следующий месяц |
| `THIS_QUARTER` | Текущий квартал |
| `THIS_YEAR` | Текущий год |
| `LAST_YEAR` | Прошлый год |
| `NEXT_YEAR` | Следующий год |
| `LAST_90_DAYS` | Последние 90 дней |
| `NEXT_90_DAYS` | Следующие 90 дней |

### Динамические литералы

Для гибких временных диапазонов:

```sql
-- Записи за последние 30 дней
WHERE CreatedDate = LAST_N_DAYS:30

-- Сделки, закрывающиеся в ближайшие 3 месяца
WHERE CloseDate = NEXT_N_MONTHS:3

-- Активность за последние 2 недели
WHERE ActivityDate = LAST_N_WEEKS:2
```

Доступные варианты: `LAST_N_DAYS`, `NEXT_N_DAYS`, `LAST_N_WEEKS`, `NEXT_N_WEEKS`, `LAST_N_MONTHS`, `NEXT_N_MONTHS`, `LAST_N_QUARTERS`, `NEXT_N_QUARTERS`, `LAST_N_YEARS`, `NEXT_N_YEARS`.

### Фискальные периоды

Для компаний с нестандартным финансовым годом:

```sql
WHERE CloseDate = THIS_FISCAL_QUARTER
WHERE CloseDate = LAST_FISCAL_YEAR
WHERE CloseDate = NEXT_N_FISCAL_QUARTERS:2
```

---

## Сортировка и пагинация

### ORDER BY

```sql
SELECT Name, CreatedDate
FROM Account
ORDER BY CreatedDate DESC
```

Множественная сортировка:

```sql
SELECT Name, Industry, AnnualRevenue
FROM Account
ORDER BY Industry ASC, AnnualRevenue DESC
```

### Обработка NULL при сортировке

```sql
-- NULL-значения в начале
ORDER BY Phone ASC NULLS FIRST

-- NULL-значения в конце (по умолчанию)
ORDER BY Phone ASC NULLS LAST
```

### LIMIT и OFFSET

```sql
-- Первые 10 записей
SELECT Name FROM Account
LIMIT 10

-- Записи с 21 по 30 (для пагинации)
SELECT Name FROM Account
LIMIT 10 OFFSET 20
```

---

## Relationship Queries (связанные запросы)

Это ключевая особенность SOQL, отличающая его от SQL. Вместо JOIN используются два типа связанных запросов.

### Child-to-Parent (снизу вверх)

Доступ к полям родительского объекта через точечную нотацию:

```sql
SELECT 
    FirstName,
    LastName,
    Account.Name,
    Account.Industry,
    Account.Owner.Name
FROM Contact
WHERE Account.Industry = 'Technology'
```

**Ограничение:** максимум 5 уровней вложенности.

### Parent-to-Child (сверху вниз)

Вложенный подзапрос для получения дочерних записей:

```sql
SELECT 
    Name,
    Industry,
    (SELECT FirstName, LastName, Email FROM Contacts)
FROM Account
WHERE Industry = 'Technology'
```

**Важно:** результат — не плоская таблица, а иерархическая структура:

```json
[
  {
    "Name": "Acme Corp",
    "Industry": "Technology",
    "Contacts": [
      {"FirstName": "John", "LastName": "Smith", "Email": "john@acme.com"},
      {"FirstName": "Jane", "LastName": "Doe", "Email": "jane@acme.com"}
    ]
  },
  {
    "Name": "Globex Inc",
    "Industry": "Technology",
    "Contacts": [
      {"FirstName": "Bob", "LastName": "Wilson", "Email": "bob@globex.com"}
    ]
  }
]
```

### Ограничения вложенных запросов

- Только **1 уровень** вложенности (нельзя вложить подзапрос в подзапрос)
- Максимум **20 подзапросов** в одном запросе
- Максимум **200 дочерних записей** на каждую родительскую запись

### Имена связей

Для стандартных объектов имена связей предопределены (Contacts, Opportunities, Cases и т.д.). Для кастомных объектов имя связи формируется как `ИмяОбъекта__r` (с суффиксом `__r`):

```sql
SELECT Name, (SELECT Name FROM Custom_Items__r)
FROM Parent_Object__c
```

---

## Агрегатные функции

### Доступные функции

| Функция | Описание |
|---------|----------|
| `COUNT()` | Количество всех записей |
| `COUNT(поле)` | Количество записей с непустым значением поля |
| `COUNT_DISTINCT(поле)` | Количество уникальных значений |
| `SUM(поле)` | Сумма |
| `AVG(поле)` | Среднее значение |
| `MIN(поле)` | Минимум |
| `MAX(поле)` | Максимум |

### Простые агрегации

```sql
-- Общее количество аккаунтов
SELECT COUNT() FROM Account

-- Статистика по сделкам
SELECT 
    COUNT(Id),
    SUM(Amount),
    AVG(Amount),
    MAX(Amount)
FROM Opportunity
WHERE StageName = 'Closed Won'
```

### GROUP BY

```sql
SELECT 
    StageName,
    COUNT(Id),
    SUM(Amount)
FROM Opportunity
GROUP BY StageName
```

Группировка по нескольким полям:

```sql
SELECT 
    Account.Industry,
    StageName,
    SUM(Amount)
FROM Opportunity
GROUP BY Account.Industry, StageName
```

### HAVING

Фильтрация после группировки:

```sql
SELECT 
    Account.Name,
    COUNT(Id) contactCount
FROM Contact
GROUP BY Account.Name
HAVING COUNT(Id) > 5
```

### GROUP BY ROLLUP

Добавляет итоговые строки:

```sql
SELECT 
    LeadSource,
    COUNT(Name)
FROM Lead
GROUP BY ROLLUP(LeadSource)
```

### GROUP BY CUBE

Добавляет все возможные комбинации итогов:

```sql
SELECT 
    Type,
    BillingCountry,
    COUNT(Id)
FROM Account
GROUP BY CUBE(Type, BillingCountry)
```

---

## Скалярные функции

### Строковые функции

| Функция | Описание | Пример |
|---------|----------|--------|
| `COALESCE(expr1, expr2, ...)` | Возвращает первое не-NULL значение | `COALESCE(Name, 'Unknown')` |
| `NULLIF(expr1, expr2)` | Возвращает NULL если expr1 = expr2 | `NULLIF(Status, 'Inactive')` |
| `CONCAT(str1, str2, ...)` | Объединяет строки | `CONCAT(FirstName, ' ', LastName)` |
| `UPPER(str)` | Преобразует в верхний регистр | `UPPER(Name)` |
| `LOWER(str)` | Преобразует в нижний регистр | `LOWER(Email)` |
| `TRIM(str)` | Удаляет пробелы по краям | `TRIM(Description)` |
| `LENGTH(str)` / `LEN(str)` | Возвращает длину строки | `LENGTH(Name)` |
| `SUBSTRING(str, start [, len])` / `SUBSTR` | Возвращает подстроку | `SUBSTRING(Name, 1, 10)` |

### Математические функции

| Функция | Описание | Пример |
|---------|----------|--------|
| `ABS(num)` | Абсолютное значение | `ABS(Amount)` |
| `ROUND(num [, decimals])` | Округление | `ROUND(Price, 2)` |
| `FLOOR(num)` | Округление вниз | `FLOOR(Amount)` |
| `CEIL(num)` / `CEILING(num)` | Округление вверх | `CEIL(Amount)` |

### Примеры использования

```sql
-- Получить полное имя
SELECT CONCAT(FirstName, ' ', LastName) AS FullName
FROM Contact

-- Использовать значение по умолчанию
SELECT COALESCE(Phone, Mobile, 'No phone') AS ContactNumber
FROM Contact

-- Нормализация для поиска
SELECT Name FROM Account WHERE UPPER(Name) = 'ACME'

-- Работа с числами
SELECT Name, ROUND(Amount * 0.1, 2) AS Commission
FROM Opportunity

-- Фильтрация по длине
SELECT Name FROM Account WHERE LENGTH(Name) > 10

-- Вложенные функции
SELECT UPPER(TRIM(Name)) FROM Account
```

### Оператор конкатенации строк

Также поддерживается оператор `||` для конкатенации строк:

```sql
SELECT FirstName || ' ' || LastName AS FullName FROM Contact
```

---

## Полиморфные поля и TYPEOF

Некоторые поля могут ссылаться на разные типы объектов. Например, поле `What` в объекте Task может указывать на Account, Opportunity или другие объекты.

### Конструкция TYPEOF

```sql
SELECT 
    Subject,
    TYPEOF What
        WHEN Account THEN Name, Industry
        WHEN Opportunity THEN Name, StageName, Amount
        ELSE Name
    END
FROM Task
WHERE What != null
```

Это позволяет выбирать разные поля в зависимости от типа связанного объекта.

---

## Специальные возможности

### FOR VIEW и FOR REFERENCE

Обновляют служебные поля при выполнении запроса:

```sql
-- Обновляет поле LastViewedDate
SELECT Name FROM Account FOR VIEW

-- Обновляет поле LastReferencedDate
SELECT Name FROM Account FOR REFERENCE
```

### FOR UPDATE

Блокирует записи для обновления (предотвращает конкурентные изменения):

```sql
SELECT Name FROM Account WHERE Id = '001xxx' FOR UPDATE
```

### USING SCOPE

Фильтрация по области видимости:

```sql
-- Только записи текущего пользователя
SELECT Name FROM Account USING SCOPE Mine

-- Записи команды пользователя
SELECT Name FROM Account USING SCOPE Team
```

### WITH SECURITY_ENFORCED

Применяет проверку прав доступа на уровне полей:

```sql
SELECT Name, Email FROM Contact WITH SECURITY_ENFORCED
```

Если у пользователя нет доступа к какому-либо полю, запрос выдаст ошибку.

---

## Полный список ограничений

### Количественные лимиты

| Параметр | Ограничение |
|----------|-------------|
| Максимум записей в результате | 50,000 |
| Максимум записей в подзапросе (на родителя) | 200 |
| Уровней вложенности (child-to-parent) | 5 |
| Уровней вложенности (parent-to-child) | 1 |
| Подзапросов в одном запросе | 20 |
| Символов в тексте запроса | 100,000 |
| OFFSET максимум | 2,000 |

### Арифметические выражения и конкатенация

SOQL поддерживает арифметические операции и конкатенацию строк в SELECT и WHERE:

**Арифметические операторы:**

| Оператор | Описание | Пример |
|----------|----------|--------|
| `+` | Сложение | `Amount + Tax` |
| `-` | Вычитание | `Price - Discount` |
| `*` | Умножение | `Amount * 0.1` |
| `/` | Деление | `Total / 100` |
| `%` | Остаток от деления | `Quantity % 10` |
| `\|\|` | Конкатенация строк | `FirstName \|\| ' ' \|\| LastName` |

**Примеры:**

```sql
-- Арифметика в SELECT
SELECT Amount * 0.1 taxAmount FROM Opportunity
SELECT Price + Tax totalPrice FROM LineItem
SELECT (Amount - Discount) * 0.9 finalPrice FROM Order

-- Арифметика в WHERE
SELECT Name FROM Opportunity WHERE Amount * 2 > 1000000
SELECT Name FROM Account WHERE AnnualRevenue / 12 > Budget

-- Конкатенация строк
SELECT FirstName || ' ' || LastName fullName FROM Contact
SELECT Name || ' (' || Industry || ')' displayName FROM Account

-- Унарные операторы
SELECT -Amount FROM Opportunity
```

**Ограничения:**
- Арифметика поддерживается только для числовых полей (`Integer`, `Float`)
- Конкатенация работает со строковыми полями и литералами
- Скобки используются для управления приоритетом операций

---

### Чего НЕТ в SOQL

В отличие от SQL, SOQL **не поддерживает**:

| Возможность | Статус в SOQL |
|-------------|---------------|
| `SELECT *` | ❌ Не поддерживается |
| `JOIN` | ❌ Только Relationship Queries |
| `UNION`, `INTERSECT`, `EXCEPT` | ❌ Не поддерживается |
| `DISTINCT` | ❌ Только `COUNT_DISTINCT()` |
| `CASE WHEN` | ❌ Только `TYPEOF` для полиморфных полей |
| Алиасы таблиц | ❌ Не поддерживается |

### Подзапросы в WHERE (Semi-Join)

SOQL поддерживает подзапросы в операторах `IN` и `NOT IN` для фильтрации записей на основе связанных данных:

```sql
-- Аккаунты, у которых есть контакты
SELECT Name FROM Account
WHERE Id IN (SELECT AccountId FROM Contact)

-- Аккаунты без закрытых сделок
SELECT Name FROM Account
WHERE Id NOT IN (SELECT AccountId FROM Opportunity WHERE StageName = 'Closed Won')

-- Подзапрос с условиями
SELECT Name FROM Account
WHERE Id IN (SELECT AccountId FROM Contact WHERE Email IS NOT NULL AND Status = 'Active')

-- Подзапрос с лимитом
SELECT Name FROM Account
WHERE Id IN (SELECT AccountId FROM Opportunity WHERE Amount > 100000 LIMIT 1000)

-- Комбинация с другими условиями
SELECT Name, Industry FROM Account
WHERE Industry = 'Technology'
  AND Id IN (SELECT AccountId FROM Contact WHERE Title LIKE '%CEO%')
```

**Ограничения подзапросов в WHERE:**

| Ограничение | Описание |
|-------------|----------|
| Одно поле в SELECT | Подзапрос должен выбирать ровно одно поле |
| Без агрегатов | Нельзя использовать `COUNT()`, `SUM()` и т.д. в подзапросе |
| Без вложенности | Нельзя вкладывать один WHERE-подзапрос в другой |
| Без lookups | Нельзя использовать точечную нотацию в SELECT подзапроса |
| Без ORDER BY | ORDER BY не поддерживается (бессмысленно для semi-join) |

---

## Примеры типичных запросов

### Поиск дубликатов

```sql
SELECT Email, COUNT(Id)
FROM Contact
WHERE Email != null
GROUP BY Email
HAVING COUNT(Id) > 1
```

### Аккаунты без контактов

```sql
SELECT Name FROM Account
WHERE Id NOT IN (SELECT AccountId FROM Contact)
```

### Активные сделки с контактами

```sql
SELECT 
    Name,
    Amount,
    StageName,
    Account.Name,
    (SELECT Name, Email FROM OpportunityContactRoles)
FROM Opportunity
WHERE IsClosed = false
  AND Amount > 50000
ORDER BY Amount DESC
LIMIT 20
```

### Статистика по менеджерам

```sql
SELECT 
    Owner.Name,
    COUNT(Id) totalOpps,
    SUM(Amount) totalAmount,
    AVG(Amount) avgDeal
FROM Opportunity
WHERE StageName = 'Closed Won'
  AND CloseDate = THIS_YEAR
GROUP BY Owner.Name
ORDER BY SUM(Amount) DESC
```

### Последняя активность по аккаунтам

```sql
SELECT 
    Name,
    (SELECT Subject, ActivityDate 
     FROM Tasks 
     ORDER BY ActivityDate DESC 
     LIMIT 1),
    (SELECT Subject, ActivityDate 
     FROM Events 
     ORDER BY ActivityDate DESC 
     LIMIT 1)
FROM Account
WHERE Industry = 'Technology'
```

---

## Контексты выполнения

SOQL возвращает разные форматы в зависимости от контекста:

### В коде Apex

Типизированные объекты (sObjects):

```java
List<Account> accounts = [
    SELECT Name, (SELECT LastName FROM Contacts)
    FROM Account
    LIMIT 10
];

for (Account acc : accounts) {
    System.debug('Account: ' + acc.Name);
    for (Contact con : acc.Contacts) {
        System.debug('  Contact: ' + con.LastName);
    }
}
```

### Через REST API

JSON с вложенной структурой:

```json
{
  "totalSize": 2,
  "done": true,
  "records": [
    {
      "attributes": {"type": "Account", "url": "/services/data/v59.0/sobjects/Account/001..."},
      "Name": "Acme Corp",
      "Contacts": {
        "totalSize": 2,
        "done": true,
        "records": [
          {"attributes": {"type": "Contact"}, "LastName": "Smith"},
          {"attributes": {"type": "Contact"}, "LastName": "Doe"}
        ]
      }
    }
  ]
}
```

### Через SOAP API

XML с аналогичной иерархической структурой.

### В Developer Console

Табличное представление для простых запросов, вложенные данные отображаются как разворачиваемые блоки.

---

## SOQL vs SOSL

Salesforce также предоставляет **SOSL (Salesforce Object Search Language)** — язык полнотекстового поиска. Ключевые отличия:

| Характеристика | SOQL | SOSL |
|----------------|------|------|
| Назначение | Структурированные запросы | Полнотекстовый поиск |
| Поиск по объектам | Один объект | Несколько объектов одновременно |
| Тип поиска | Точное совпадение | Нечёткий поиск, синонимы |
| Индексация | По полям | По поисковому индексу |

Пример SOSL:

```sql
FIND {Acme*} IN ALL FIELDS
RETURNING Account(Name), Contact(FirstName, LastName)
```

---

## Рекомендации по оптимизации

### Используйте индексированные поля

Поля Id, Name, Owner, CreatedDate, SystemModstamp индексируются автоматически. Фильтрация по ним работает быстрее.

### Избегайте негативных операторов в начале

```sql
-- Медленно (начинается с негативного условия)
WHERE NOT Status = 'Closed'

-- Быстрее (позитивное условие первым)
WHERE Status IN ('Open', 'In Progress', 'Pending')
```

### Ограничивайте результаты

Всегда используйте LIMIT, если не нужны все записи:

```sql
SELECT Name FROM Account LIMIT 100
```

### Выбирайте только нужные поля

```sql
-- Плохо: выбираем лишние поля
SELECT Id, Name, Description, BillingAddress, ShippingAddress, ...

-- Хорошо: только необходимое
SELECT Id, Name FROM Account
```

---

## Заключение

SOQL — мощный инструмент для работы с данными в Salesforce. Его главные особенности:

1. **Объектная модель** вместо табличной
2. **Relationship Queries** вместо JOIN
3. **Иерархические результаты** вместо плоских таблиц
4. **Встроенные литералы дат** для удобной фильтрации
5. **Жёсткие ограничения** для защиты производительности

При проектировании собственного языка запросов для CRM стоит заимствовать удачные решения SOQL (литералы дат, relationship queries), но учитывать его ограничения и, возможно, добавить недостающие возможности SQL там, где они действительно нужны.
