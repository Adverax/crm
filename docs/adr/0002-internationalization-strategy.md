# ADR-0002: Internationalization Strategy

**Status:** Accepted
**Date:** 2026-02-08
**Participants:** @roman_myakotin

## Context

The CRM platform targets the enterprise segment, where users may work in
different languages. Internationalization affects two types of content:

- **Dynamic** — object metadata, field metadata, picklist values, validation rules.
  Created by administrators at runtime, stored in the database, quantity grows over time.
- **Static** — buttons, menus, system errors, notification templates.
  Created by developers, deployed with code, quantity is fixed.

A unified approach to i18n is needed that accounts for the different nature of these content types.

## Considered Options

### Option A: i18n keys everywhere

All strings (both in the database and in the UI) are replaced with keys (`object.contact.label`),
which are resolved through a single dictionary.

**Pros:**
- Full unification

**Cons:**
- Without the dictionary, users see keys instead of text
- Complicates the admin UI — creating a custom object requires populating the dictionary
- UI strings in the database — extra queries on every render for data that does not change between deploys
- The lifecycles of dynamic and static content are fundamentally different

### Option B: Plain strings without i18n

All labels are stored as plain strings in a single language.

**Pros:**
- Maximum simplicity

**Cons:**
- Multilingual support is impossible
- Unsuitable for enterprises with international teams

### Option C: Default value + translation overlay (chosen)

A unified pattern for the entire platform, but with two storage mechanisms:

1. **Dynamic content** — default value inline in the DB record + polymorphic `translations` table
2. **Static content** — locale JSON files (vue-i18n on the frontend, embedded files in the Go binary)

**Pros:**
- Works out of the box in the default language — translations are optional
- Fallback chain: translation -> default -> key — the user never sees blank content
- One pattern, clear to both developers and administrators
- Dynamic content does not slow down UI rendering
- Static content is version-controlled in git

**Cons:**
- Two storage mechanisms (database + files), although the conceptual pattern is unified

## Decision

We adopt **Option C** — a unified "default value + optional translation" pattern with two storage backends.

### Dynamic content — `translations` table

```sql
CREATE TABLE translations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type   VARCHAR(100) NOT NULL,  -- 'ObjectDef', 'FieldDef', 'PicklistValue', ...
    resource_id     UUID         NOT NULL,  -- Resource ID
    field_name      VARCHAR(100) NOT NULL,  -- 'label', 'plural_label', 'description', 'help_text'
    locale          VARCHAR(10)  NOT NULL,  -- 'en', 'ru', 'de', ...
    value           TEXT         NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (resource_type, resource_id, field_name, locale)
);
```

Covers:
- Object metadata (`label`, `plural_label`, `description`)
- Field metadata (`label`, `help_text`, `description`)
- Picklist values
- Validation rule messages
- Layout and page names

### Static content — locale files

```
web/src/locales/ru.json   ← default language
web/src/locales/en.json   ← translation
internal/locales/ru.json  ← server-side messages (errors, notifications)
internal/locales/en.json
```

### Unified resolve pattern

```
1. Determine locale from user context
2. Look up translation for locale
3. If not found — fall back to default value
4. If not found — return key (only for static content)
```

### User locale

Stored in the user profile (`users.locale`). Passed in the context of every request.

## Consequences

- Fields `label`, `plural_label`, `description`, `help_text` in metadata tables store the default value directly
- The `translations` table is created when needed (not required for MVP)
- Frontend uses vue-i18n with JSON files
- Backend uses embedded JSON files for system messages
- The metadata API returns resolved values based on the locale from the request
