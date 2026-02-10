# Стратегия продвижения CRM-платформы

**Дата:** 2026-02-08
**Модель:** Open Core (AGPL v3 + Enterprise)
**Бюджет:** Минимальный (преимущественно бесплатные каналы)
**Целевая готовность к запуску:** Phase 5-6 (auth + standard objects)

---

## 1. Что может сделать Claude Code прямо сейчас

### 1.1. Подготовка репозитория к публичному релизу

| Задача | Описание | Статус |
|--------|----------|--------|
| LICENSE | AGPL v3 в корне | Готово |
| ee/LICENSE | Adverax Commercial License | Готово |
| README.md | Продающий README с badges, screenshots, quickstart | Нужно сделать |
| CONTRIBUTING.md | Гайд для контрибьюторов: setup, code style, PR process | Нужно сделать |
| CODE_OF_CONDUCT.md | Стандартный Contributor Covenant | Нужно сделать |
| SECURITY.md | Политика responsible disclosure | Нужно сделать |
| .github/ISSUE_TEMPLATE/ | Шаблоны для bug report, feature request | Нужно сделать |
| .github/PULL_REQUEST_TEMPLATE.md | Шаблон PR с чеклистом | Нужно сделать |
| .github/FUNDING.yml | Ссылки на спонсорство (GitHub Sponsors) | Нужно сделать |

### 1.2. Документация

| Задача | Описание |
|--------|----------|
| docs/getting-started.md | Быстрый старт: Docker → первый запрос за 5 минут |
| docs/architecture.md | Обзор архитектуры для разработчиков (metadata engine, SOQL, security layers) |
| docs/self-hosting.md | Гайд по self-hosted развёртыванию (Docker, docker-compose, env variables) |
| docs/api-reference.md | Описание REST API с примерами (генерируется из OpenAPI spec) |
| docs/custom-objects.md | Как создавать custom objects через metadata engine |
| docs/security-model.md | Подробное описание OLS/FLS/RLS для enterprise-аудитории |
| docs/comparison.md | Сравнение с Salesforce, Bitrix24, SuiteCRM, Twenty — feature matrix |
| docs/faq.md | Часто задаваемые вопросы |

### 1.3. Landing page

Статический сайт (можно на Vue или простом HTML + Tailwind), публикуемый через GitHub Pages:

- Hero-секция: одно предложение о продукте + CTA (GitHub / Docker quickstart)
- Feature highlights: metadata engine, SOQL, 3-layer security, custom objects
- Сравнительная таблица с конкурентами
- Pricing: Community (бесплатно) / Enterprise (по запросу)
- Architecture diagram (SVG)
- Quick demo GIF / embedded video

### 1.4. Демо-материалы

| Задача | Описание |
|--------|----------|
| Seed-данные | Реалистичные тестовые данные: компании, контакты, сделки, задачи |
| Postman/Bruno коллекция | Готовая коллекция API-запросов для быстрого знакомства |
| docker-compose.demo.yml | Один файл для запуска полного демо (DB + seed + API + frontend) |
| Makefile: `make demo` | Одна команда для запуска демо |
| asciinema-запись | Терминальная демонстрация: создание custom object → CRUD через API |
| Screenshots | UI скриншоты для README и landing page |

### 1.5. Контент (статьи, черновики)

Claude Code может написать полные черновики статей:

**Технические статьи (Habr, dev.to):**

1. «Как мы спроектировали metadata-driven CRM на Go» — архитектура, table-per-object, почему не EAV
2. «3 слоя безопасности в CRM: OLS, FLS, RLS на PostgreSQL» — security model с SQL-примерами
3. «Свой SOQL на Go: лексер, парсер, executor» — как построить DSL для запросов (Phase 3+)
4. «Open Core в 2026: AGPL + ee/ — как монетизировать open source» — наш опыт лицензирования
5. «Table-per-object vs EAV vs JSON: как хранить данные для custom objects» — сравнение подходов
6. «PostgreSQL как платформа: pgTAP, closure tables, materialized views для CRM» — advanced PG

**Бизнес-статьи (VC.ru, Spark):**

7. «Зачем бизнесу self-hosted CRM в 2026 году» — импортозамещение, контроль данных, compliance
8. «CRM для разработчиков: почему low-code не работает для сложных процессов» — позиционирование

### 1.6. SEO и discovability

| Задача | Описание |
|--------|----------|
| GitHub Topics | Добавить topics: `crm`, `golang`, `postgresql`, `open-source`, `self-hosted`, `salesforce-alternative` |
| GitHub Description | Лаконичное описание: "Open-source CRM platform with metadata-driven objects, SOQL, and 3-layer security (OLS/FLS/RLS)" |
| Social Preview | OG-image для GitHub (1280×640px) |
| awesome-go PR | Добавить в awesome-go список после стабилизации |
| awesome-selfhosted PR | Добавить в awesome-selfhosted |
| alternativeto.net | Зарегистрировать как альтернативу Salesforce, HubSpot, SuiteCRM |

---

## 2. Стратегия каналов продвижения

### 2.1. Бесплатные каналы

| Канал | Аудитория | Действие | Когда |
|-------|-----------|----------|-------|
| **GitHub** | Разработчики | Quality README, topics, releases, discussions | При публичном релизе |
| **Habr** | Русскоязычные разработчики | Серия технических статей (см. 1.5) | 1 статья в 2 недели после релиза |
| **dev.to** | Международные разработчики | Переводы/адаптации статей с Habr | Параллельно с Habr |
| **Product Hunt** | Ранние adopters, стартапы | Launch day: README + landing + demo | Одноразово, когда демо готово |
| **Hacker News** | Техническая аудитория | Show HN пост | После Product Hunt |
| **Reddit** | Самохостеры, CRM-пользователи | r/selfhosted, r/golang, r/CRM | При релизе + при публикации статей |
| **Telegram** | Русскоязычные Go/CRM-сообщества | Посты в каналах Go, devops, CRM | При релизе |
| **Twitter/X** | Техническая аудитория | Тред про архитектуру, короткие посты | Постоянно |
| **LinkedIn** | B2B аудитория, CTO, IT-директора | Статьи о self-hosted CRM, импортозамещении | При релизе |
| **YouTube** | Разработчики, decision makers | Скринкасты: архитектура, демо, setup | После Phase 6 |
| **Конференции** | Go-разработчики | Доклад на GolangConf / GopherCon | Через 3-6 месяцев после релиза |

### 2.2. Низкобюджетные платные каналы

| Канал | Бюджет | Ожидание |
|-------|--------|----------|
| GitHub Sponsors | Бесплатно (получение, не трата) | Донаты от пользователей |
| Targeted ads (Twitter/LinkedIn) | $50-100/мес | B2B лиды для enterprise |
| Спонсорство Go-подкастов | $100-200 за выпуск | Узнаваемость в Go-сообществе |

---

## 3. Позиционирование

### 3.1. Варианты позиционирования

| Позиция | Слоган | Целевая аудитория | Конкуренты |
|---------|--------|-------------------|------------|
| **Open-source Salesforce** | «Enterprise CRM без vendor lock-in» | Компании, уходящие с Salesforce | SuiteCRM, EspoCRM, Twenty |
| **Self-hosted CRM для РФ** | «CRM с полным контролем данных» | Российский бизнес (импортозамещение) | Bitrix24, amoCRM |
| **Developer-first CRM** | «CRM, который разработчики любят» | Технические команды, стартапы | Twenty, Folk, Attio |
| **CRM-платформа** | «Постройте свою CRM без ограничений» | ISV, интеграторы | Salesforce Platform, Creatio |

**Рекомендация:** Начать с «Developer-first open-source CRM platform» — это привлечёт первых пользователей (разработчиков), которые станут advocates в своих компаниях. Позже расширять на B2B.

### 3.2. Ключевые differentiators

1. **Metadata-driven objects** — создавайте custom objects без миграций и перезапуска
2. **SOQL** — единый язык запросов с автоматическим security enforcement
3. **3-layer security** — OLS/FLS/RLS из коробки, как у Salesforce, но open source
4. **Table-per-object** — нативная производительность PostgreSQL, без EAV-overhead
5. **Self-hosted** — ваши данные на вашем сервере, полный контроль
6. **Go + PostgreSQL** — минимальные ресурсы, один бинарник, простое развёртывание

---

## 4. Roadmap продвижения

### Phase A: Подготовка (сейчас → Phase 5)

Claude Code делает параллельно с разработкой:
- [ ] README.md (продающий, с badges и quickstart)
- [ ] CONTRIBUTING.md, CODE_OF_CONDUCT.md, SECURITY.md
- [ ] GitHub issue/PR templates
- [ ] docs/getting-started.md, docs/architecture.md
- [ ] docs/comparison.md (feature matrix)
- [ ] Landing page (HTML + Tailwind, GitHub Pages)
- [ ] Черновики 3-4 статей для Habr/dev.to
- [ ] Seed-данные и Postman коллекция
- [ ] docker-compose.demo.yml + `make demo`

### Phase B: Soft Launch (Phase 5-6 готовы)

- [ ] Репозиторий переводится в public
- [ ] GitHub Release v0.1.0
- [ ] Публикация первой статьи на Habr
- [ ] Посты в Telegram-каналах и Reddit
- [ ] Регистрация на alternativeto.net
- [ ] Добавление в awesome-selfhosted

### Phase C: Public Launch (стабильный MVP)

- [ ] Product Hunt Launch Day
- [ ] Show HN пост
- [ ] Публикация landing page
- [ ] Серия статей (1 раз в 2 недели)
- [ ] PR в awesome-go
- [ ] YouTube: скринкаст-демо

### Phase D: Рост (после запуска)

- [ ] GitHub Discussions для community support
- [ ] Discord/Telegram-чат для пользователей
- [ ] Сбор feedback и feature requests
- [ ] Case studies от первых пользователей
- [ ] Доклад на GolangConf
- [ ] Первые enterprise-клиенты → blog posts

---

## 5. Метрики успеха

| Метрика | Target (3 мес.) | Target (6 мес.) | Target (12 мес.) |
|---------|-----------------|-----------------|------------------|
| GitHub Stars | 100 | 500 | 2000 |
| Docker pulls | 200 | 1000 | 5000 |
| Contributors | 3 | 10 | 25 |
| Habr views (total) | 5K | 20K | 50K |
| Self-hosted installations | 10 | 50 | 200 |
| Enterprise leads | — | 5 | 20 |

---

## 6. Что Claude Code НЕ может сделать

Для полноты картины — задачи, которые требуют человека:

- **Публикация** — я могу написать текст, но опубликовать на Habr/PH/Reddit должен человек
- **Видео** — могу написать сценарий, но запись и монтаж — вручную
- **Нетворкинг** — конференции, meetups, личные контакты
- **Переговоры** — enterprise sales, партнёрства
- **Юридическая экспертиза** — лицензии стоит проверить с юристом
- **Дизайн** — логотип, брендинг, UI polish (могу сгенерировать код, но не визуальный дизайн)
- **Решения о приоритетах** — что продвигать первым, куда вкладывать время

---

## 7. Приоритетный план действий

Что стоит начать делать прямо сейчас (параллельно с Phase 2):

1. **README.md** — первое, что видят люди. Должен быть идеальным
2. **docs/architecture.md** — для привлечения контрибьюторов
3. **Черновик статьи #1** — «Metadata-driven CRM на Go» — чтобы была готова к релизу
4. **CONTRIBUTING.md** — без него не будет контрибьюторов
5. **docker-compose.demo.yml** — одна команда для запуска
