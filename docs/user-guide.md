# CRM Administrator Guide

## Table of Contents

1. [Introduction](#1-introduction)
2. [Authentication](#2-authentication)
   - [Login](#21-login)
   - [Access Tokens](#22-access-tokens)
   - [Password Recovery](#23-password-recovery)
   - [Admin Password Reset](#24-admin-password-reset)
   - [Logout](#25-logout)
   - [Current User](#26-current-user)
3. [Key Concepts](#3-key-concepts)
4. [Metadata Management](#4-metadata-management)
   - [Objects](#41-objects)
   - [Fields](#42-fields)
5. [Security Management](#5-security-management)
   - [Roles](#51-roles)
   - [Permission Sets](#52-permission-sets)
   - [Profiles](#53-profiles)
   - [Users](#54-users)
   - [Groups](#55-groups)
   - [Sharing Rules](#56-sharing-rules)
   - [Manual Sharing](#57-manual-sharing)
6. [SOQL — Query Language](#6-soql--query-language)
   - [Syntax](#61-syntax)
   - [Operators and Functions](#62-operators-and-functions)
   - [Date Literals](#63-date-literals)
   - [Relationships and Subqueries](#64-relationships-and-subqueries)
   - [Security in SOQL](#65-security-in-soql)
   - [API](#66-api)
   - [Limits](#67-limits)
7. [DML — Data Manipulation Language](#7-dml--data-manipulation-language)
   - [INSERT](#71-insert)
   - [UPDATE](#72-update)
   - [DELETE](#73-delete)
   - [UPSERT](#74-upsert)
   - [Functions in DML](#75-functions-in-dml)
   - [Security in DML](#76-security-in-dml)
   - [API](#77-api)
   - [Limits](#78-limits)
8. [Territory Management (Enterprise)](#8-territory-management-enterprise)
   - [Territory Models](#81-territory-models)
   - [Territories](#82-territories)
   - [Object Defaults](#83-object-defaults)
   - [User Assignment](#84-user-assignment)
   - [Record Assignment](#85-record-assignment)
   - [Assignment Rules](#86-assignment-rules)
9. [App Templates](#9-app-templates)
   - [Available Templates](#91-available-templates)
   - [Applying a Template](#92-applying-a-template)
   - [API](#93-api)
   - [Limitations](#94-limitations)
10. [Record Management (Generic CRUD)](#10-record-management-generic-crud)
    - [Describe API](#101-describe-api)
    - [Record CRUD](#102-record-crud)
    - [System Fields](#103-system-fields)
    - [Pagination](#104-pagination)
    - [Security in Record Operations](#105-security-in-record-operations)
11. [Validation Rules & Dynamic Defaults](#11-validation-rules--dynamic-defaults)
    - [Validation Rules](#111-validation-rules)
    - [CEL Expression Language](#112-cel-expression-language)
    - [Dynamic Defaults](#113-dynamic-defaults)
    - [DML Pipeline](#114-dml-pipeline)
    - [Admin UI](#115-admin-ui)
    - [CEL Validation Endpoint](#116-cel-validation-endpoint)
12. [Custom Functions](#12-custom-functions)
    - [Overview](#121-overview)
    - [Creating Functions](#122-creating-functions)
    - [fn.* Namespace](#123-fn-namespace)
    - [Parameters & Return Types](#124-parameters--return-types)
    - [Dependency Management](#125-dependency-management)
    - [Expression Builder](#126-expression-builder)
    - [API](#127-api)
    - [Limits](#128-limits)
13. [Object Views](#13-object-views)
    - [Overview](#131-overview)
    - [Creating an Object View](#132-creating-an-object-view)
    - [Config Structure](#133-config-structure)
    - [Visual Constructor](#134-visual-constructor)
    - [Resolution Logic](#135-resolution-logic)
    - [FLS Intersection](#136-fls-intersection)
    - [Describe API Extension](#137-describe-api-extension)
    - [CRM UI Rendering](#138-crm-ui-rendering)
    - [API](#139-api)
14. [Procedures](#14-procedures)
    - [Overview](#141-overview)
    - [Creating a Procedure](#142-creating-a-procedure)
    - [Definition & Commands](#143-definition--commands)
    - [Versioning](#144-versioning)
    - [Dry Run & Execution](#145-dry-run--execution)
    - [Constructor UI](#146-constructor-ui)
    - [API](#147-api)
    - [Limits](#148-limits)
15. [Named Credentials](#15-named-credentials)
    - [Overview](#151-overview)
    - [Credential Types](#152-credential-types)
    - [Creating a Credential](#153-creating-a-credential)
    - [Test Connection](#154-test-connection)
    - [Usage Log](#155-usage-log)
    - [API](#156-api)
16. [Profile Navigation](#16-profile-navigation)
    - [Overview](#161-overview)
    - [Navigation Config](#162-navigation-config)
    - [Resolution Logic](#163-resolution-logic)
    - [Admin UI](#164-admin-ui)
    - [API](#165-api)
17. [Automation Rules](#17-automation-rules)
    - [Overview](#171-overview)
    - [API](#172-api)
18. [Layouts](#18-layouts)
    - [Overview](#181-overview)
    - [Layout Config](#182-layout-config)
    - [Form Merge & Fallback](#183-form-merge--fallback)
    - [Admin UI](#184-admin-ui)
    - [API](#185-api)
19. [Shared Layouts](#19-shared-layouts)
    - [Overview](#191-overview)
    - [Types](#192-types)
    - [layout_ref & Overrides](#193-layout_ref--overrides)
    - [API](#194-api)
20. [Common Scenarios](#20-common-scenarios)

---

## 1. Introduction

### Purpose

The CRM admin panel is designed for system administrators and allows:

- Managing **metadata** — defining objects and their fields (data schema), configuring object visibility (OWD).
- Managing **security** — configuring roles, profiles, permission sets, users, groups, sharing rules, and manual record sharing.
- Applying **app templates** — bootstrapping the system with pre-configured objects and fields.
- Configuring **validation rules** — CEL expressions that check data on every save.
- Creating **custom functions** — reusable CEL expressions callable from any context.
- Configuring **object views** — role-based UI per profile with read config (fields, actions, queries, computed) and optional write config (validation, defaults, computed, mutations).
- Configuring **profile navigation** — per-profile sidebar with grouped items (objects, links, pages, dividers), OLS intersection for object items, `ov_api_name` for page views, fallback to alphabetical flat list.
- Building **procedures** — named JSON-described business logic sequences (record operations, computations, branching, HTTP integrations) with a visual Constructor UI, versioning (draft/published), and dry-run testing.
- Managing **named credentials** — encrypted secret storage for HTTP integrations (API keys, basic auth, OAuth2 client credentials) with SSRF protection and usage audit logging.
- Configuring **automation rules** — DML triggers (before/after insert/update/delete) with CEL conditions that invoke published procedures.
- Creating **layouts** — per Object View + form factor (desktop/tablet/mobile) + mode (edit/view) to control page structure, section grids, field presentation, and list configuration.
- Managing **shared layouts** — reusable configuration snippets (field/section/list) referenced via `layout_ref` from layouts, with RESTRICT delete protection.
- Working with **records** — creating, editing, and deleting records of any object through a universal CRUD interface.

### Audience

This document is intended for the system administrator responsible for CRM platform configuration: creating data structures, setting up access permissions, and managing users.

### Navigation

The admin panel is available at `/admin`. The interface consists of two parts:

- **Sidebar** (left, 240px width) — section navigation.
- **Main area** (right) — content of the selected section.

Sidebar structure:

| Item | Route |
|------|-------|
| **Objects** | `/admin/metadata/objects` |
| **Templates** | `/admin/templates` |
| **Functions** | `/admin/functions` |
| **Object Views** | `/admin/metadata/object-views` |
| **Navigation** | `/admin/metadata/navigation` |
| **Procedures** | `/admin/metadata/procedures` |
| **Credentials** | `/admin/metadata/credentials` |
| **Layouts** | `/admin/metadata/layouts` |
| **Shared Layouts** | `/admin/metadata/shared-layouts` |
| **Automation Rules** | `/admin/metadata/automation-rules` |
| **Security** (collapsible group) | |
| — Roles | `/admin/security/roles` |
| — Permission Sets | `/admin/security/permission-sets` |
| — Profiles | `/admin/security/profiles` |
| **Users** | `/admin/security/users` |

The "Security" group expands automatically when navigating to any of its subsections. Validation rules are accessed from the object detail page (Objects → select object → Validation Rules tab).

The CRM workspace is available at `/app` with its own sidebar showing all accessible objects for the current user.

All detail pages contain **breadcrumbs** for convenient back-navigation.

### Common UI Elements

- **Tables** with pagination (20 records per page). "Previous" / "Next" buttons appear when multiple pages exist.
- **Action menu** (three vertical dots) in each table row — contains "Open" and "Delete" items.
- **Delete confirmation dialog** — the system prompts for confirmation before deleting any entity.
- **Toast notifications** — messages for successful operations and errors.

---

## 2. Authentication

Working with the admin panel requires authentication. The system uses JWT (JSON Web Token) — a standard authorization mechanism for REST APIs.

### 2.1. Login

**Route:** `/login`

The login page contains a form:

| Field | Type | Required | Description |
|-------|------|:---:|-------------|
| Username | Text | Yes | Account username |
| Password | Password | Yes | Account password |

**"Login"** button — sends an authentication request. On success — redirects to the admin panel (`/admin`).

**"Forgot password?"** link — navigates to the password recovery page.

**Rate limiting:** The login endpoint is rate-limited — no more than 5 attempts per 15 minutes from a single IP address. When the limit is exceeded, requests are rejected with a 429 (Too Many Requests) error.

### 2.2. Access Tokens

The system uses JWT authentication with two token types:

| Token | Lifetime | Purpose |
|-------|----------|---------|
| **Access token** | 15 minutes | API request authorization (`Authorization: Bearer <token>` header) |
| **Refresh token** | 7 days | Obtaining a new token pair without re-entering the password |

**Automatic refresh:** When the access token expires, the frontend automatically requests a new token pair using the refresh token. If the refresh token has also expired — the user is redirected to the login page.

**Token rotation:** Each refresh invalidates the old refresh token and issues a new one. This protects against reuse of intercepted tokens.

**Storage:** The refresh token is stored on the server as a SHA-256 hash. The raw token is not saved — in case of a database leak, the tokens cannot be used.

### 2.3. Password Recovery

#### Reset Request

**Route:** `/forgot-password`

| Field | Type | Required | Description |
|-------|------|:---:|-------------|
| Email | Email | Yes | Email address associated with the account |

**"Send"** button — initiates sending an email with a password reset link. The system always responds with success, even if the specified email is not registered (to protect against user enumeration).

**"Back to login"** link — returns to the login page.

#### Password Reset

**Route:** `/reset-password?token=<token>`

The user follows the link from the email. The form contains:

| Field | Type | Required | Description |
|-------|------|:---:|-------------|
| New password | Password | Yes | New password (minimum 8 characters, maximum 128) |
| Confirm password | Password | Yes | Repeat of the new password |

**"Reset password"** button — sets the new password. After the reset, all active user sessions are invalidated (refresh tokens are deleted), and the user is redirected to the login page.

The reset token is valid for **1 hour** and can only be used once.

### 2.4. Admin Password Reset

An administrator can set a password for any user via the API:

**Route:** `PUT /api/v1/admin/security/users/:id/password`

Request body:
```json
{
  "password": "new_password"
}
```

Password requirements: minimum 8 characters, maximum 128 characters.

#### Initial Administrator Password

On the first system startup, the system administrator account password is set via an environment variable:

```
ADMIN_INITIAL_PASSWORD=your-secure-password
```

The password is set once: if the administrator already has a password, the environment variable is ignored.

### 2.5. Logout

The **"Logout"** button is located at the bottom of the sidebar (next to the current user's name).

On logout:
- The current refresh token is invalidated on the server.
- Local tokens are removed from the browser.
- The user is redirected to the login page.

To terminate **all** active sessions (e.g., if a compromise is suspected), use the API:

```
POST /api/v1/auth/logout-all
```

This will delete all of the user's refresh tokens.

### 2.6. Current User

**API Route:** `GET /api/v1/auth/me`

Returns information about the currently authenticated user:

| Field | Description |
|-------|-------------|
| `id` | User UUID |
| `username` | Username |
| `email` | Email address |
| `first_name` | First name |
| `last_name` | Last name |
| `profile_id` | UUID of the assigned profile |
| `role_id` | UUID of the assigned role (or `null`) |
| `is_active` | Activity status |

The current user's name is displayed at the bottom of the admin panel sidebar.

---

## 3. Key Concepts

### Security Model

The CRM security system implements three layers of access control:

1. **OLS (Object-Level Security)** — object-level permissions. Determines which CRUD operations are allowed for a given object.
2. **FLS (Field-Level Security)** — field-level permissions. Determines which fields a user can read and/or edit.
3. **RLS (Row-Level Security)** — record-level permissions. Determines which specific records a user can see based on OWD visibility, role hierarchy, groups, sharing rules, and manual sharing.

### Security Entity Hierarchy

```
Profile
  └── Base Permission Set (grant type)
        ├── OLS: object permissions
        └── FLS: field permissions

User
  ├── Profile (exactly one, required)
  ├── Role (optional, used for RLS)
  ├── Additional Permission Sets (grant and/or deny)
  └── Groups (automatic and manual)

Groups
  ├── Personal — one per user (auto-created)
  ├── Role — one per role (auto-created)
  ├── Role & Subordinates — one per role (auto-created)
  └── Public — created by administrator

Sharing Rules
  ├── Owner-based — based on record ownership
  └── Criteria-based — based on record field values
```

### OWD (Organization-Wide Defaults) — Object Visibility

Visibility (OWD) determines the **baseline access level** to all records of an object for all users. This is the starting point for RLS — access is then extended through role hierarchy, groups, and sharing rules.

| Visibility | Description |
|------------|-------------|
| `private` | Users can only see their own records (by `owner_id` field). Access is extended through role hierarchy, groups, and sharing. |
| `public_read` | All users can read all records. Updates — only by owner, hierarchy, or sharing. |
| `public_read_write` | Full access for everyone. RLS filtering is disabled. No share table is created. |
| `controlled_by_parent` | Access is determined by the parent object (for composition relationships). |

Visibility is set when creating an object and can be changed later. When changing visibility, the system automatically creates or deletes the share table.

### Groups

Groups are a mechanism for combining users for sharing purposes. All sharing operations (manual, via rules) are performed through groups.

| Group Type | Description | Creation |
|------------|-------------|----------|
| `personal` | Contains a single user | Automatically when creating a user |
| `role` | Contains all users with a given role | Automatically when creating a role |
| `role_and_subordinates` | Contains users with a given role and all subordinate roles | Automatically when creating a role |
| `public` | Arbitrary group of users | Manually by administrator |

When a user's role changes, their membership in role groups is recalculated automatically.

### Effective Permissions Calculation

```
effective = grants & ~denies
```

1. Base permissions are determined by the user's **profile** (through its base permission set).
2. Additional **grant sets** extend permissions (bitwise OR).
3. **Deny sets** globally suppress specified permissions (bitwise AND with inversion).

Deny always overrides Grant.

### Bitmask Permission Model

**OLS (object permissions):**

| Bit | Value | Operation |
|-----|-------|-----------|
| 1   | `0001` | Read    |
| 2   | `0010` | Create  |
| 4   | `0100` | Update  |
| 8   | `1000` | Delete  |

Full access = `15` (all 4 bits set).

**FLS (field permissions):**

| Bit | Value | Operation |
|-----|-------|-----------|
| 1   | `01`  | Read    |
| 2   | `10`  | Write   |

Full access = `3` (both bits set).

### Roles

Roles form a **hierarchy** (tree). They are used for Row-Level Security:

- A user with a higher role in the hierarchy can see records of subordinates (read only).
- A user can have exactly one role (or none).

---

## 4. Metadata Management

### 4.1. Objects

An object is a description of an entity in the system (analogous to a table). Each object stores metadata: name, type, behavioral flags.

#### Object List

**Route:** `/admin/metadata/objects`

The page displays a table with the following columns:

| Column | Description |
|--------|-------------|
| API Name | Unique object identifier (link to detail page) |
| Label | Human-readable name |
| Type | `standard` or `custom` (displayed as a badge) |
| Created | Creation date |

**Filtering:** dropdown to filter by object type:
- All types
- Standard
- Custom

**Actions:**
- **"Create Object"** button — navigates to the creation form.
- Row context menu — "Open" and "Delete" (deletion is only available if `isDeleteableObject = true` and `isPlatformManaged = false`).

#### Creating an Object

**Route:** `/admin/metadata/objects/new`

The form consists of two cards:

**"Basic Information" card:**

| Field | Type | Required | Description |
|-------|------|:---:|-------------|
| API Name | Text | Yes | Unique object name (e.g., `Invoice__c`). Set at creation, not editable afterwards. |
| Label | Text | Yes | Display name in singular (e.g., "Invoice") |
| Plural Label | Text | Yes | Display name in plural (e.g., "Invoices") |
| Object Type | Select | Yes | `standard` or `custom`. Set at creation, not editable afterwards. |
| Visibility (OWD) | Select | No | Baseline record access level: `private` (default), `public_read`, `public_read_write`, `controlled_by_parent`. See [OWD](#owd-organization-wide-defaults--object-visibility). |
| Description | Textarea | No | Free-form object description |

**Flags card** (three groups of checkboxes):

**"Record Permissions" group:**

| Flag | Description |
|------|-------------|
| `isCreateable` | Record creation is allowed for this object |
| `isUpdateable` | Record updates are allowed |
| `isDeleteable` | Record deletion is allowed |
| `isQueryable` | Object is available for SOQL queries |
| `isSearchable` | Object is available for full-text search |

**"Object Settings" group:**

| Flag | Description |
|------|-------------|
| `isVisibleInSetup` | Object is displayed in the setup interface |
| `isCustomFieldsAllowed` | Custom fields can be added to the object |
| `isDeleteableObject` | The object itself can be deleted |

**"Capabilities" group:**

| Flag | Description |
|------|-------------|
| `hasActivities` | Activities support (tasks, events) |
| `hasNotes` | Notes support |
| `hasHistoryTracking` | Change history tracking |
| `hasSharingRules` | Sharing rules support (RLS) |

After clicking **"Create"**, the user is redirected to the detail page of the created object.

#### Editing an Object

**Route:** `/admin/metadata/objects/:objectId`

The detail page contains two tabs:

**"General" tab:**
- All the same fields as during creation, but **API Name** and **Object Type** are read-only (disabled).
- Label, Plural Label, Description, **Visibility (OWD)**, and all flags are editable.
- When changing visibility to/from `public_read_write`, the system automatically creates or deletes the object's share table.
- **"Save"** button — saves changes.

**"Fields (N)" tab:**
- Displays the object's field table (see [4.2. Fields](#42-fields)).
- The tab header shows the field count.

**Deleting an object:**
- **"Delete Object"** button (red, top-right corner) is visible only if `isDeleteableObject = true` and `isPlatformManaged = false`.
- Deletion requires confirmation. Warning text: *"Object '{label}' ({apiName}) and all its fields will be permanently deleted."*

### 4.2. Fields

A field is an attribute of an object (analogous to a table column). Each field has a type, subtype, and additional configuration.

#### Field Table

The field table is displayed on the "Fields" tab of the object detail page. For each field, the following are shown: API Name, Label, Type, Subtype, "Required" and "Unique" flags.

Actions:
- **"Add Field"** button — opens the creation dialog.
- **"Edit"** button — opens the editing dialog.
- **"Delete"** button — deletes the field (with confirmation).

#### Field Types and Subtypes

| Type | Display Name | Available Subtypes |
|------|-------------|-------------------|
| `text` | Text | `plain` (Plain text), `area` (Multiline), `rich` (Rich text), `email` (Email), `phone` (Phone), `url` (URL) |
| `number` | Number | `integer` (Integer), `decimal` (Decimal), `currency` (Currency), `percent` (Percent), `auto_number` (Auto-number) |
| `boolean` | Boolean | — (no subtypes) |
| `datetime` | Date/Time | `date` (Date), `datetime` (Date and time), `time` (Time) |
| `picklist` | Picklist | `single` (Single-select), `multi` (Multi-select) |
| `reference` | Reference | `association` (Association), `composition` (Composition), `polymorphic` (Polymorphic) |

#### Creating a Field

The field creation dialog contains the following fields:

| Field | Type | Required | Description |
|-------|------|:---:|-------------|
| API Name | Text | Yes | Unique field name within the object |
| Label | Text | Yes | Display name of the field |
| Type | Select | Yes | One of six types: text, number, boolean, datetime, picklist, reference |
| Subtype | Select | Depends on type | Available subtypes depend on the selected type. No subtypes for boolean. |
| Reference Object | Select | For reference | Object that the field references |
| Description | Text | No | Field description |
| Help Text | Text | No | Help text for users |
| Required | Checkbox | No | Makes the field required |
| Unique | Checkbox | No | Field value must be unique |
| Custom | Checkbox | No | Marks the field as custom |
| Sort Order | Number | No | Display order of the field |

#### Type/Subtype-Specific Configuration

Depending on the selected type/subtype combination, additional configuration fields are available:

**Text fields (text/plain, text/area, text/rich):**

| Parameter | Description |
|-----------|-------------|
| Max Length | Maximum text length |
| Default Value | Value substituted when creating a record |

**Text fields (email, phone, url):**

| Parameter | Description |
|-----------|-------------|
| Max Length | Maximum text length |

**Integer (number/integer):**

| Parameter | Description |
|-----------|-------------|
| Default Value | Default value |

**Decimal / Currency / Percent (number/decimal, number/currency, number/percent):**

| Parameter | Description |
|-----------|-------------|
| Precision (total digits) | Total number of significant digits |
| Scale (after decimal) | Number of digits after the decimal point |
| Default Value | Default value |

**Auto-number (number/auto_number):**

| Parameter | Description |
|-----------|-------------|
| Format | Numbering template, e.g., `INV-{0000}` |
| Starting Value | Counter starting value |

**Boolean:**

| Parameter | Description |
|-----------|-------------|
| Default Value | `true` or `false` |

**Date/Time (datetime/date, datetime/datetime, datetime/time):**

| Parameter | Description |
|-----------|-------------|
| Default Value | Default value |

**Association (reference/association):**

| Parameter | Description |
|-----------|-------------|
| Relationship Name | Reverse relationship name |
| On Delete | `set_null` (Clear) or `restrict` (Prevent) |

**Composition (reference/composition):**

| Parameter | Description |
|-----------|-------------|
| Relationship Name | Reverse relationship name |
| On Delete | `cascade` (Cascade delete) or `restrict` (Prevent) |
| Allow Reparent | Whether parent reassignment is allowed |

**Polymorphic reference (reference/polymorphic):**

| Parameter | Description |
|-----------|-------------|
| Relationship Name | Reverse relationship name |
| On Delete | `set_null` (Clear) or `restrict` (Prevent) |

#### Editing a Field

The edit dialog allows changing:
- Label
- Description
- Help Text
- Required (checkbox)
- Unique (checkbox)
- Configuration parameters (depend on type/subtype)
- Sort Order

**API Name**, **Type**, and **Subtype** are not editable after creation.

#### Deleting a Field

Field deletion is performed via the "Delete" button in the field table. The operation is irreversible.

---

## 5. Security Management

### 5.1. Roles

Roles define a user's position in the organizational hierarchy. They are used for Row-Level Security: users with a higher role gain read access to records of subordinates.

#### Role List

**Route:** `/admin/security/roles`

Table with the following columns:

| Column | Description |
|--------|-------------|
| API Name | Unique role identifier (link to detail page) |
| Label | Human-readable name |
| Parent Role | Parent role name or "—" for root roles |
| Created | Creation date |

Actions:
- **"Create Role"** button — navigates to the creation form.
- Row context menu — "Open" and "Delete".

#### Creating a Role

**Route:** `/admin/security/roles/new`

| Field | Type | Required | Description |
|-------|------|:---:|-------------|
| API Name | Text | Yes | Unique role name (e.g., `sales_manager`). Set at creation, not editable afterwards. |
| Label | Text | Yes | Display name (e.g., "Sales Manager") |
| Parent Role | Select | No | Select from existing roles. "No parent" option — root role. |
| Description | Textarea | No | Free-form role description |

After creation, the user is redirected to the role detail page.

> **Automatic groups:** When a role is created, the system automatically creates two groups: `role_{api_name}` (all users with this role) and `role_and_sub_{api_name}` (users with this role and all subordinate roles).

#### Editing a Role

**Route:** `/admin/security/roles/:roleId`

- **API Name** — read-only.
- **Label**, **Parent Role**, **Description** — editable.
- The current role is excluded from the parent roles list (a role cannot be its own parent).
- **"Save"** button — saves changes.

#### Deleting a Role

**"Delete Role"** button (red) in the top-right corner of the detail page. Requires confirmation.

### 5.2. Permission Sets

A Permission Set is a container for OLS and FLS permissions. There are two types:

- **Grant** — extends permissions (allows operations).
- **Deny** — globally suppresses permissions (forbids operations, even if they are allowed in other sets).

#### Permission Set List

**Route:** `/admin/security/permission-sets`

Table with the following columns:

| Column | Description |
|--------|-------------|
| API Name | Unique identifier (link to detail page) |
| Label | Human-readable name |
| Type | `Grant` or `Deny` (displayed as a badge) |
| Created | Creation date |

**Filtering:** dropdown to filter by type:
- All types
- Grant
- Deny

Actions:
- **"Create Permission Set"** button — navigates to the creation form.
- Row context menu — "Open" and "Delete".

#### Creating a Permission Set

**Route:** `/admin/security/permission-sets/new`

| Field | Type | Required | Description |
|-------|------|:---:|-------------|
| API Name | Text | Yes | Unique name (e.g., `sales_read_access`). Not editable after creation. |
| Label | Text | Yes | Display name (e.g., "Sales Read Access") |
| Type | Select | Yes | `Grant (allows)` or `Deny (forbids)`. Not editable after creation. |
| Description | Textarea | No | Free-form description |

#### Permission Set Detail Page

**Route:** `/admin/security/permission-sets/:permissionSetId`

Contains three tabs:

**"General" tab:**
- API Name (read-only), Label, Type (read-only), Description.
- "Save" button to save label and description changes.

**"Object Permissions" (OLS) tab:**

A table where each row is an object from metadata. For each object, a group of checkboxes is displayed:

| Checkbox | Bit | Description |
|----------|-----|-------------|
| Read | 1 | Permission to read object records |
| Create | 2 | Permission to create records |
| Update | 4 | Permission to update records |
| Delete | 8 | Permission to delete records |

Changes are saved **instantly** when toggling a checkbox (no "Save" button needed).

**"Field Permissions" (FLS) tab:**

1. First, select an object from the dropdown.
2. After selection, a table of the object's fields is displayed. For each field — a group of checkboxes:

| Checkbox | Bit | Description |
|----------|-----|-------------|
| Read | 1 | Permission to read the field value |
| Write | 2 | Permission to edit the field value |

Changes are also saved **instantly**.

#### Deleting a Permission Set

**"Delete Permission Set"** button (red) in the top-right corner. Requires confirmation.

### 5.3. Profiles

A Profile is a mandatory security entity assigned to every user. When a profile is created, a **base permission set** (grant type) is automatically created, which defines the user's initial permissions.

#### Profile List

**Route:** `/admin/security/profiles`

Table with the following columns:

| Column | Description |
|--------|-------------|
| API Name | Unique identifier (link to detail page) |
| Label | Human-readable name |
| Created | Creation date |

Actions:
- **"Create Profile"** button — navigates to the creation form.
- Row context menu — "Open" and "Delete".

#### Creating a Profile

**Route:** `/admin/security/profiles/new`

| Field | Type | Required | Description |
|-------|------|:---:|-------------|
| API Name | Text | Yes | Unique name (e.g., `sales_profile`). Not editable after creation. |
| Label | Text | Yes | Display name (e.g., "Sales Profile") |
| Description | Textarea | No | Free-form description |

When a profile is created, the system automatically generates a base permission set (grant).

#### Editing a Profile

**Route:** `/admin/security/profiles/:profileId`

- API Name (read-only), Label, Description — editable.
- **"Open Base Permission Set"** link — navigates to the profile's base permission set editing page, where OLS and FLS are configured.
- **"Save"** button — saves label and description changes.

> **Important:** To configure a profile's OLS/FLS, navigate to its base permission set via the link on the profile detail page.

#### Deleting a Profile

**"Delete Profile"** button (red) in the top-right corner. Requires confirmation.

### 5.4. Users

A User is a person's account in the system. Each user is assigned a profile (required) and optionally — a role and additional permission sets.

#### User List

**Route:** `/admin/security/users`

Table with the following columns:

| Column | Description |
|--------|-------------|
| Username | Username (link to detail page) |
| Email | Email address |
| Name | Full name (first name + last name) |
| Profile | Assigned profile name |
| Role | Assigned role name or "—" |
| Status | "Active" / "Inactive" badge |

Actions:
- **"Create User"** button — navigates to the creation form.
- Row context menu — "Open" and "Delete".

#### Creating a User

**Route:** `/admin/security/users/new`

The form consists of three cards:

**"Credentials" card:**

| Field | Type | Required | Description |
|-------|------|:---:|-------------|
| Username | Text | Yes | Unique login name (e.g., `john.doe`). Not editable after creation. |
| Email | Email | Yes | Email address |

**"Personal Information" card:**

| Field | Type | Required | Description |
|-------|------|:---:|-------------|
| First Name | Text | No | User's first name |
| Last Name | Text | No | User's last name |

**"Security" card:**

| Field | Type | Required | Description |
|-------|------|:---:|-------------|
| Profile | Select | Yes | Select from created profiles |
| Role | Select | No | Select from existing roles. "No role" option — user is not in the hierarchy. |

> **Automatic actions on creation:** The system automatically creates a personal group `personal_{username}` and adds the user to it, as well as to role groups (if a role is specified).

#### Editing a User

**Route:** `/admin/security/users/:userId`

Contains two tabs:

**"General" tab:**

- **Username** — read-only.
- **Email**, **First Name**, **Last Name** — editable.
- **Profile** and **Role** — changeable via dropdowns. When the role changes, membership in role groups is recalculated automatically.
- **Active** — toggle switch. An inactive user cannot log in.
- **"Save"** button — saves changes.

**"Permission Sets" tab:**

Managing additional permission sets assigned to the user.

Table of assigned sets:

| Column | Description |
|--------|-------------|
| Label | Permission set name |
| Type | `Grant` or `Deny` (badge) |
| Assigned | Assignment date |
| Action | "Revoke" button |

Actions:
- **"Assign Permission Set"** button — opens the assignment dialog. The dialog contains a dropdown with permission sets not yet assigned to this user. Each set's type (grant/deny) is shown next to it.
- **"Revoke"** button — revokes the permission set from the user (with confirmation).

#### Deleting a User

**"Delete User"** button (red) in the top-right corner. Requires confirmation.

### 5.5. Groups

Groups combine users for sharing purposes. All records are shared with groups, not individual users.

#### Automatic Groups

Three group types are created automatically and are not managed manually:

- **Personal** — created when a user is created. Contains exactly one user. API Name: `personal_{username}`.
- **Role** — created when a role is created. Contains all users with the given role. API Name: `role_{api_name}`.
- **Role & Subordinates** — created when a role is created. Contains users with the given role and all subordinate roles. API Name: `role_and_sub_{api_name}`.

Membership in these groups is recalculated automatically when a user's role or the role hierarchy changes.

#### Public Groups

Public groups (`public`) are created manually by the administrator and allow combining users in any way.

| Field | Type | Required | Description |
|-------|------|:---:|-------------|
| API Name | Text | Yes | Unique group name |
| Label | Text | Yes | Display name |
| Type | Select | Yes | `personal`, `role`, `role_and_subordinates`, or `public` |

#### Group Membership

The following can be added to a group:
- A **user** (by `member_user_id`).
- **Another group** (by `member_group_id`) — all members of the nested group become members of the parent group.

> **Restriction:** Each addition specifies either a user or a group — not both simultaneously.

Effective membership (including nested groups) is calculated automatically and stored in the `security.effective_group_members` cache.

### 5.6. Sharing Rules

Sharing Rules extend record visibility for specific groups of users. Rules work additively — they only add access, they cannot restrict it.

#### Rule Types

| Type | Description |
|------|-------------|
| `owner_based` | Records owned by users in the **source group** become visible to users in the **target group**. |
| `criteria_based` | Records matching a criterion (field + operator + value) become visible to users in the **target group**. |

#### Creating a Rule

| Field | Type | Required | Description |
|-------|------|:---:|-------------|
| Object | UUID | Yes | Object to which the rule applies |
| Rule Type | Select | Yes | `owner_based` or `criteria_based` |
| Source Group | UUID | Yes | Group of record owners (for owner_based) |
| Target Group | UUID | Yes | Group that receives access |
| Access Level | Select | Yes | `read` — read only, `read_write` — read and write |
| Criteria Field | Text | For criteria_based | Field name for filtering (e.g., `status`) |
| Criteria Operator | Text | For criteria_based | Comparison operator (e.g., `equals`) |
| Criteria Value | Text | For criteria_based | Value for comparison (e.g., `closed`) |

> **Note:** Generation of entries in share tables based on rules is performed in Phase 3/4 (DML engine). Phase 2b stores rule definitions and generates outbox events for cache recalculation.

### 5.7. Manual Sharing

Manual sharing allows granting access to a specific record for a specific group of users. An entry is added to the object's share table with reason `manual`.

#### Granting Access (Share)

| Parameter | Type | Description |
|-----------|------|-------------|
| Table | Text | Object table name (e.g., `obj_invoice`) |
| Record ID | UUID | ID of the record to be shared |
| Group | UUID | ID of the group to receive access |
| Access Level | Select | `read` or `read_write` |

When re-sharing the same record with the same group, the access level is updated.

#### Revoking Access (Revoke)

Deletes the entry from the share table with reason `manual` for the specified (record, group) pair.

#### Viewing Shares

Returns a list of all entries from the share table for a specific record (all reasons: `manual`, `sharing_rule`, `territory`).

### How RLS Determines Record Visibility

For an object with visibility `private` or `controlled_by_parent`, a user can see a record if at least one condition is met:

1. **Owner** — the user is the record owner (`owner_id`).
2. **Role hierarchy** — the user is above the owner in the role hierarchy (read only).
3. **Share table** — the record is shared with a group that includes the user (via manual sharing, rules, or territories).

For `public_read` — everyone can read; updates follow the same rules as `private`.
For `public_read_write` — RLS filtering is completely disabled.

---

## 6. SOQL — Query Language

SOQL (Structured Object Query Language) is the platform's built-in query language. All data read operations go through SOQL with automatic OLS, FLS, and RLS enforcement.

### 6.1. Syntax

#### Basic Structure

```
SELECT <fields> FROM <object>
[WHERE <condition>]
[GROUP BY <fields>]
[HAVING <condition>]
[ORDER BY <fields> [ASC|DESC] [NULLS FIRST|LAST]]
[LIMIT <number>]
[OFFSET <number>]
[FOR UPDATE]
```

#### Query Examples

Simple query:
```
SELECT Id, Name, Email FROM Contact WHERE Status = 'Active'
```

With sorting and limit:
```
SELECT Id, Name, Amount FROM Deal ORDER BY Amount DESC LIMIT 10
```

Grouping with aggregates:
```
SELECT Status, COUNT(), SUM(Amount) FROM Deal GROUP BY Status HAVING COUNT() > 5
```

Field aliases:
```
SELECT Name AS ContactName, Email AS ContactEmail FROM Contact
```

### 6.2. Operators and Functions

#### Comparison Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `=` or `==` | Equal | `WHERE Status = 'Active'` |
| `!=` or `<>` | Not equal | `WHERE Status != 'Closed'` |
| `>` | Greater than | `WHERE Amount > 1000` |
| `<` | Less than | `WHERE Amount < 500` |
| `>=` | Greater than or equal | `WHERE CreatedAt >= 2024-01-01` |
| `<=` | Less than or equal | `WHERE Amount <= 10000` |

#### Logical Operators

| Operator | Example |
|----------|---------|
| `AND` | `WHERE Status = 'Active' AND Amount > 1000` |
| `OR` | `WHERE Status = 'Closed' OR Status = 'Won'` |
| `NOT` | `WHERE NOT IsDeleted` |

#### Special Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `IS NULL` | NULL check | `WHERE Phone IS NULL` |
| `IS NOT NULL` | Not-NULL check | `WHERE Email IS NOT NULL` |
| `IN` | List membership | `WHERE Id IN ('001', '002', '003')` |
| `NOT IN` | List exclusion | `WHERE Status NOT IN ('Closed', 'Archived')` |
| `LIKE` | Pattern matching | `WHERE Name LIKE 'Acme%'` |
| `NOT LIKE` | Inverse LIKE | `WHERE Email NOT LIKE '%test%'` |

`LIKE` patterns: `%` — any characters, `_` — single character.

#### Aggregate Functions

| Function | Description | Example |
|----------|-------------|---------|
| `COUNT()` | Record count | `SELECT COUNT() FROM Contact` |
| `COUNT_DISTINCT(field)` | Unique value count | `SELECT COUNT_DISTINCT(Status) FROM Deal` |
| `SUM(field)` | Sum | `SELECT SUM(Amount) FROM Deal` |
| `AVG(field)` | Average | `SELECT AVG(Amount) FROM Deal` |
| `MIN(field)` | Minimum | `SELECT MIN(CreatedAt) FROM Contact` |
| `MAX(field)` | Maximum | `SELECT MAX(Amount) FROM Deal` |

#### Built-in Functions

**String:**

| Function | Description |
|----------|-------------|
| `UPPER(str)` | Convert to uppercase |
| `LOWER(str)` | Convert to lowercase |
| `TRIM(str)` | Remove leading/trailing whitespace |
| `LENGTH(str)` or `LEN(str)` | String length |
| `SUBSTRING(str, start, length)` or `SUBSTR(...)` | Substring |
| `CONCAT(str1, str2, ...)` | Concatenation |

**Numeric:**

| Function | Description |
|----------|-------------|
| `ABS(num)` | Absolute value |
| `ROUND(num, decimals)` | Rounding |
| `FLOOR(num)` | Floor |
| `CEIL(num)` or `CEILING(num)` | Ceiling |

**NULL handling:**

| Function | Description |
|----------|-------------|
| `COALESCE(expr1, expr2, ...)` | First non-NULL value |
| `NULLIF(expr1, expr2)` | NULL if values are equal |

### 6.3. Date Literals

SOQL supports symbolic date literals that are automatically resolved to specific dates at query execution time.

**Static literals:**

| Literal | Description |
|---------|-------------|
| `TODAY` | Today |
| `YESTERDAY` | Yesterday |
| `TOMORROW` | Tomorrow |
| `THIS_WEEK` / `LAST_WEEK` / `NEXT_WEEK` | Current / previous / next week |
| `THIS_MONTH` / `LAST_MONTH` / `NEXT_MONTH` | Current / previous / next month |
| `THIS_QUARTER` / `LAST_QUARTER` / `NEXT_QUARTER` | Current / previous / next quarter |
| `THIS_YEAR` / `LAST_YEAR` / `NEXT_YEAR` | Current / previous / next year |
| `LAST_90_DAYS` / `NEXT_90_DAYS` | Last / next 90 days |
| `THIS_FISCAL_QUARTER` / `THIS_FISCAL_YEAR` | Current fiscal quarter / year |

**Parameterized literals (with N):**

| Literal | Example |
|---------|---------|
| `LAST_N_DAYS:N` | `WHERE CreatedAt >= LAST_N_DAYS:30` |
| `NEXT_N_DAYS:N` | `WHERE DueDate <= NEXT_N_DAYS:7` |
| `LAST_N_WEEKS:N` / `NEXT_N_WEEKS:N` | Last/next N weeks |
| `LAST_N_MONTHS:N` / `NEXT_N_MONTHS:N` | Last/next N months |
| `LAST_N_QUARTERS:N` / `NEXT_N_QUARTERS:N` | Last/next N quarters |
| `LAST_N_YEARS:N` / `NEXT_N_YEARS:N` | Last/next N years |

### 6.4. Relationships and Subqueries

#### Lookup Queries (child → parent)

Use dot notation to access parent object fields:

```
SELECT Name, Account.Name, Account.Owner.Name FROM Contact
```

Nesting depth — up to 5 levels.

#### Relationship Subqueries (parent → child)

To retrieve child records, use subqueries in SELECT:

```
SELECT Name, (SELECT Email, Phone FROM Contacts) FROM Account
```

Subqueries support WHERE, ORDER BY, and LIMIT.

#### Semi-join (IN with subquery)

```
SELECT Name FROM Contact
WHERE AccountId IN (SELECT Id FROM Account WHERE Industry = 'Tech')
```

#### TYPEOF (polymorphic fields)

For polymorphic reference fields:

```
SELECT TYPEOF What
  WHEN Account THEN Name, Industry
  WHEN Opportunity THEN Name, Amount
  ELSE Name
END
FROM Task
```

### 6.5. Security in SOQL

Every SOQL query automatically passes through three security levels:

1. **OLS** — the `Read` permission on the object in SELECT and FROM is checked. If the user does not have access to the object, the query is rejected.
2. **FLS** — the `Read` permission on each field in SELECT is checked. System fields (`Id`, `OwnerId`, `CreatedAt`, `UpdatedAt`, `CreatedById`, `UpdatedById`) are always accessible.
3. **RLS** — a WHERE condition is automatically injected into the SQL, limiting results to records visible to the user (based on OWD, role hierarchy, groups, and sharing).

### 6.6. API

**GET** `/api/v1/query?q=<SOQL>`

Parameters:
- `q` (required) — SOQL query string.

Example:
```
GET /api/v1/query?q=SELECT Id, Name FROM Account LIMIT 10
```

**POST** `/api/v1/query`

Request body:
```json
{
  "query": "SELECT Id, Name FROM Account WHERE Industry = 'Tech'",
  "pageSize": 100
}
```

Response:
```json
{
  "totalSize": 3,
  "done": true,
  "records": [
    {"Id": "...", "Name": "Acme Inc"},
    {"Id": "...", "Name": "Globex Corp"},
    {"Id": "...", "Name": "TechStart"}
  ]
}
```

### 6.7. Limits

| Parameter | Default Value |
|-----------|--------------|
| Maximum records (LIMIT) | 50,000 |
| Maximum OFFSET | 2,000 |
| Lookup depth (dot notation) | 5 levels |
| Maximum subqueries | 20 |
| Records per subquery (per parent) | 200 |
| Query length | 100,000 characters |

### 6.8. SOQL Editor

Административный интерфейс предоставляет Rich Editor для написания SOQL-запросов (используется в Object View Queries tab и будет переиспользоваться в отчётах).

**Возможности:**

- **Подсветка синтаксиса** — CodeMirror-based редактор с токенизацией SOQL: ключевые слова (SELECT, FROM, WHERE, AND, OR, ORDER BY, GROUP BY, LIMIT и др.), функции (COUNT, SUM, AVG, MIN, MAX, COALESCE, UPPER, LOWER и др.), date literals (TODAY, LAST_N_DAYS:30 и др.), строки, числа, параметры (`:param`), комментарии (`--`).
- **Контекстное автодополнение** — подсказки зависят от позиции курсора: после SELECT — поля и функции, после FROM — имена объектов, после WHERE — поля, date literals, операторы, после ORDER BY — поля, ASC, DESC.
- **Серверная валидация** — кнопка Validate отправляет `POST /api/v1/admin/soql/validate` и показывает ошибки с указанием строки и колонки (клик на ошибку перемещает курсор).
- **Тестовый запуск** — кнопка Test Query выполняет запрос (`POST /api/v1/query`, pageSize: 5) и показывает первые 5 записей в мини-таблице.
- **Object/Field picker** — popover с двумя табами (Objects, Fields), поиск по имени. Клик вставляет имя в позицию курсора.
- **Переключение режимов** — Editor (CodeMirror) / Plain Text (textarea) по кнопке.
- **Автоопределение FROM-объекта** — из текста запроса извлекается имя объекта, автоматически загружаются его поля для автодополнения.

**API endpoint:**

```
POST /api/v1/admin/soql/validate
```

Request:
```json
{"query": "SELECT Id, Name FROM Account WHERE Industry = 'Tech'"}
```

Response (valid):
```json
{"valid": true, "object": "Account", "fields": ["Id", "Industry", "Name"]}
```

Response (error):
```json
{
  "valid": false,
  "errors": [{"message": "unknown object: Foo", "line": 1, "column": 20}]
}
```

---

## 7. DML — Data Manipulation Language

DML (Data Manipulation Language) is the built-in language for write operations. All data changes go through DML with automatic OLS and FLS enforcement. For UPDATE and DELETE, RLS is additionally applied.

### 7.1. INSERT

Inserting one or more records.

```
INSERT INTO <object> (field1, field2, ...) VALUES (value1, value2, ...)
```

**Single record:**
```
INSERT INTO Contact (FirstName, LastName, Email)
VALUES ('John', 'Smith', 'john@example.com')
```

**Multiple records (batch):**
```
INSERT INTO Contact (FirstName, LastName, Email)
VALUES
  ('John', 'Smith', 'john@example.com'),
  ('Jane', 'Doe', 'jane@example.com')
```

### 7.2. UPDATE

Updating records by condition.

```
UPDATE <object> SET field1 = value1, field2 = value2 [WHERE condition]
```

**Examples:**
```
UPDATE Contact SET Status = 'Active' WHERE Id = '550e8400-e29b-41d4-a716-446655440000'

UPDATE Account SET Revenue = 1000000, Industry = 'Tech'
WHERE Name LIKE 'Acme%'
```

> **Important:** RLS automatically restricts UPDATE to only records visible to the current user. Attempting to update another user's record will not cause an error, but the record will not be affected.

### 7.3. DELETE

Deleting records by condition.

```
DELETE FROM <object> WHERE condition
```

**Example:**
```
DELETE FROM Task WHERE Status = 'Completed' AND CreatedAt < 2023-01-01
```

> **Important:** By default, `WHERE` is mandatory for DELETE (protection against accidental deletion of all records). RLS restricts deletion to visible records only.

### 7.4. UPSERT

Insert or update based on an external identifier.

```
UPSERT <object> (field1, field2, ...) VALUES (value1, value2, ...) ON <external_key>
```

If a record with the specified external key value exists — UPDATE is performed. If it doesn't exist — INSERT.

**Example:**
```
UPSERT Account (Name, Revenue, ExternalId)
VALUES ('Acme Inc', 1000000, 'ext_001')
ON ExternalId
```

**Batch UPSERT:**
```
UPSERT Account (Name, Revenue, ExternalId)
VALUES
  ('Acme', 1000000, 'ext_001'),
  ('Globex', 500000, 'ext_002')
ON ExternalId
```

### 7.5. Functions in DML

In INSERT, UPDATE, and UPSERT values, you can use the same functions as in SOQL:

```
INSERT INTO Contact (FirstName, Email)
VALUES (UPPER('john'), COALESCE(NULL, 'default@example.com'))

UPDATE Contact SET Email = LOWER(Email) WHERE Status = 'Pending'
```

### 7.6. Security in DML

| Operation | OLS | FLS | RLS |
|-----------|-----|-----|-----|
| **INSERT** | CanCreate | CanWrite (each field) | — (not needed) |
| **UPDATE** | CanUpdate | CanWrite (each field) | WHERE injection (visible records only) |
| **DELETE** | CanDelete | — | WHERE injection (visible records only) |
| **UPSERT** | CanCreate + CanUpdate | CanWrite (each field) | — (INSERT path) |

System fields (`Id`, `CreatedAt`, `UpdatedAt`, `CreatedById`, `UpdatedById`) are read-only and cannot be specified in DML. The `OwnerId` field is writable.

### 7.7. API

**POST** `/api/v1/data`

Request body:
```json
{
  "statement": "INSERT INTO Contact (FirstName, LastName) VALUES ('John', 'Smith')"
}
```

Response:
```json
{
  "rows_affected": 1,
  "inserted_ids": ["550e8400-e29b-41d4-a716-446655440000"]
}
```

For UPDATE/DELETE — `updated_ids` / `deleted_ids` respectively.

### 7.8. Limits

| Parameter | Default Value |
|-----------|--------------|
| Maximum batch rows (INSERT/UPSERT) | 10,000 |
| Statement length | 100,000 characters |
| WHERE required for DELETE | Yes |

---

## 8. Territory Management (Enterprise)

> **Note:** Territory Management is only available in the Enterprise edition (Adverax Commercial License). These endpoints are not present in the Community edition.

Territories are a mechanism for organizing record access based on geographic, functional, or other criteria. Territories work alongside the role hierarchy and extend the RLS model.

Base route: `/api/v1/admin/territory`

### 8.1. Territory Models

A territory model is a top-level container that defines a set of territories. A model has a lifecycle: `planning` → `active` → `archived`.

#### Routes

| Method | Route | Description |
|--------|-------|-------------|
| POST | `/territory/models` | Create a model |
| GET | `/territory/models` | List models (with pagination) |
| GET | `/territory/models/:id` | Get model by ID |
| PUT | `/territory/models/:id` | Update model |
| DELETE | `/territory/models/:id` | Delete model |
| POST | `/territory/models/:id/activate` | Activate model |
| POST | `/territory/models/:id/archive` | Archive model |

#### Creating a Model

```json
{
  "api_name": "fy2026",
  "label": "Territory Model FY2026",
  "description": "Territory model for fiscal year 2026"
}
```

#### Lifecycle

- **Planning** — initial state. Territory structure can be edited.
- **Active** — model is active, territories affect record visibility.
- **Archived** — model is archived, territories do not affect access.

Transitions: `planning → active` (activation), `active → archived` (archiving).

### 8.2. Territories

A territory is a node in a hierarchical tree within a model. Territories form a tree (parent → children).

#### Routes

| Method | Route | Description |
|--------|-------|-------------|
| POST | `/territory/territories` | Create a territory |
| GET | `/territory/territories?model_id=<uuid>` | List territories of a model |
| GET | `/territory/territories/:id` | Get territory by ID |
| PUT | `/territory/territories/:id` | Update territory |
| DELETE | `/territory/territories/:id` | Delete territory |

#### Creating a Territory

```json
{
  "model_id": "...",
  "parent_id": null,
  "api_name": "west_region",
  "label": "West Region",
  "description": "Includes all western states"
}
```

With `parent_id = null`, a root territory is created. When specifying `parent_id` — a child territory.

### 8.3. Object Defaults

Define what access level a territory grants to records of a specific object.

#### Routes

| Method | Route | Description |
|--------|-------|-------------|
| POST | `/territory/territories/:id/object-defaults` | Set a default |
| GET | `/territory/territories/:id/object-defaults` | List territory defaults |
| DELETE | `/territory/territories/:id/object-defaults/:objectId` | Remove a default |

#### Setting a Default

```json
{
  "object_id": "...",
  "access_level": "read_write"
}
```

Access levels: `read`, `read_write`.

### 8.4. User Assignment

Users are assigned to territories (M:M relationship). A user can be assigned to multiple territories.

#### Routes

| Method | Route | Description |
|--------|-------|-------------|
| POST | `/territory/territories/:id/users` | Assign a user |
| GET | `/territory/territories/:id/users` | List territory users |
| DELETE | `/territory/territories/:id/users/:userId` | Unassign a user |

#### Assignment

```json
{
  "user_id": "..."
}
```

### 8.5. Record Assignment

Specific records can be linked to a territory.

#### Routes

| Method | Route | Description |
|--------|-------|-------------|
| POST | `/territory/territories/:id/records` | Link a record |
| GET | `/territory/territories/:id/records` | List territory records |
| DELETE | `/territory/territories/:id/records/:recordId?object_id=<uuid>` | Unlink a record |

#### Linking a Record

```json
{
  "record_id": "...",
  "object_id": "...",
  "reason": "manual"
}
```

### 8.6. Assignment Rules

Rules automatically assign records to territories based on criteria.

#### Routes

| Method | Route | Description |
|--------|-------|-------------|
| POST | `/territory/assignment-rules` | Create a rule |
| GET | `/territory/assignment-rules?territory_id=<uuid>` | List territory rules |
| GET | `/territory/assignment-rules/:id` | Get rule by ID |
| PUT | `/territory/assignment-rules/:id` | Update rule |
| DELETE | `/territory/assignment-rules/:id` | Delete rule |

### How Territories Affect RLS

A user assigned to a territory gains access to records of that territory (and all child territories) at the level defined in object-defaults. Access is provided through object share tables with reason `territory`.

---

## 9. App Templates

App Templates allow administrators to bootstrap the CRM with pre-configured objects and fields on first launch. Instead of manually creating objects one by one, you can apply a ready-made template that creates all necessary objects, fields, and security permissions in one operation.

### 9.1 Available Templates

The platform ships with two built-in templates:

#### Sales CRM

A classic sales pipeline template with 4 objects and 36 fields.

| Object | Fields | Visibility | Description |
|--------|--------|------------|-------------|
| Account | 9 | Private | Companies and organizations. Fields: name, website, phone, industry, employee_count, annual_revenue, billing_city, billing_country, description |
| Contact | 9 | Private | People associated with accounts. Fields: first_name, last_name, email, phone, title, account_id (reference → Account), department, date_of_birth, description |
| Opportunity | 9 | Private | Sales deals. Fields: name, account_id (ref → Account), contact_id (ref → Contact), amount, stage, probability, close_date, is_won, description |
| Task | 9 | PublicReadWrite | Activities and to-do items. Fields: subject, status, priority, due_date, account_id (ref → Account), contact_id (ref → Contact), opportunity_id (ref → Opportunity), is_completed, description |

Reference fields use `set_null` on delete — if a parent record is deleted, the reference field is set to NULL rather than cascading the deletion.

#### Recruiting

An applicant tracking system with 4 objects and 28 fields.

| Object | Fields | Visibility | Description |
|--------|--------|------------|-------------|
| Position | 8 | PublicRead | Open job positions. Fields: title, department, location, status, salary_min, salary_max, headcount, description |
| Candidate | 9 | Private | Job applicants. Fields: first_name, last_name, email, phone, current_company, current_title, linkedin_url, source, notes |
| Application | 6 | Private | Candidate applications. Fields: position_id (ref → Position), candidate_id (ref → Candidate), stage, applied_date, is_rejected, rejection_reason |
| Interview | 5 | Private | Scheduled interviews. Fields: application_id (ref → Application, cascade delete), scheduled_at, interview_type, result, feedback |

The Interview object uses `cascade` on delete for application_id — deleting an Application automatically removes its Interviews.

### 9.2 Applying a Template

Templates can only be applied to an **empty database** (no objects exist yet). This is a one-time operation.

**Steps:**

1. Navigate to `/admin/templates` in the admin panel.
2. The page displays available templates as cards showing the template name, description, and object/field count.
3. Click the "Apply" button on the desired template.
4. The system creates all objects and fields from the template, grants full OLS permissions to the System Administrator profile, and invalidates the metadata cache.
5. A success notification confirms the template was applied.

**Guard mechanism:** If any objects already exist in the database, the apply operation returns a `409 Conflict` error. This prevents accidental double application or mixing templates.

**Security auto-grant:** After applying a template, the System Administrator profile automatically receives full CRUD permissions (Read + Create + Update + Delete) for all created objects. Other profiles need manual permission configuration.

### 9.3 API

#### List Templates

```
GET /api/v1/admin/templates
```

**Response:**
```json
{
  "data": [
    {
      "id": "sales_crm",
      "label": "Sales CRM",
      "description": "CRM for sales teams: accounts, contacts, opportunities, and tasks",
      "status": "available",
      "objects": 4,
      "fields": 36
    },
    {
      "id": "recruiting",
      "label": "Recruiting",
      "description": "Applicant tracking system: positions, candidates, applications, and interviews",
      "status": "available",
      "objects": 4,
      "fields": 28
    }
  ]
}
```

Template status values: `available` (can be applied), `applied` (already applied), `blocked` (unavailable).

#### Apply Template

```
POST /api/v1/admin/templates/{templateId}/apply
```

Request body: `{}` (empty).

**Success response:**
```json
{
  "data": {
    "template_id": "sales_crm",
    "message": "template applied successfully"
  }
}
```

**Error responses:**

| HTTP Code | Condition |
|-----------|-----------|
| 404 | Unknown template ID |
| 409 | Objects already exist in the database |

### 9.4 Limitations

- **One-time only:** A template can only be applied once to an empty database.
- **All-or-nothing:** You cannot select a subset of objects from a template.
- **No undo:** There is no "unapply" operation. To start over, reset the database.
- **No customization before apply:** Templates are applied as-is. Customization (adding fields, changing labels) is done after application through the standard metadata management API.

---

## 10. Record Management (Generic CRUD)

The platform provides a universal set of REST endpoints for working with records of **any object**. There is no per-object code — the same endpoints serve Account, Contact, Opportunity, or any custom object. The frontend renders forms and tables dynamically from metadata.

### 10.1 Describe API

The Describe API provides metadata introspection — information about available objects and their fields.

#### List Accessible Objects

```
GET /api/v1/describe
```

Returns all objects the current user can read, filtered by OLS permissions. This endpoint powers the sidebar navigation.

**Response:**
```json
{
  "data": [
    {
      "api_name": "Account",
      "label": "Account",
      "plural_label": "Accounts",
      "is_createable": true,
      "is_queryable": true
    },
    {
      "api_name": "Contact",
      "label": "Contact",
      "plural_label": "Contacts",
      "is_createable": true,
      "is_queryable": true
    }
  ]
}
```

Only objects where the user has OLS Read permission appear in the list.

#### Describe a Specific Object

```
GET /api/v1/describe/{objectName}
```

Returns complete object metadata including all fields the user can see (filtered by FLS). Also includes a resolved `form` property from Object Views (see [section 13.7](#137-describe-api-extension)).

**Response:**
```json
{
  "data": {
    "api_name": "Account",
    "label": "Account",
    "plural_label": "Accounts",
    "is_createable": true,
    "is_updateable": true,
    "is_deleteable": true,
    "fields": [
      {
        "api_name": "Id",
        "label": "ID",
        "field_type": "text",
        "field_subtype": null,
        "is_required": false,
        "is_read_only": true,
        "is_system_field": true,
        "sort_order": -6,
        "config": {}
      },
      {
        "api_name": "Name",
        "label": "Name",
        "field_type": "text",
        "field_subtype": null,
        "is_required": true,
        "is_read_only": false,
        "is_system_field": false,
        "sort_order": 1,
        "config": {
          "max_length": 255,
          "default_value": null,
          "values": []
        }
      }
    ]
  }
}
```

**Field properties:**

| Property | Description |
|----------|-------------|
| `api_name` | Field identifier (e.g., "Name", "Email") |
| `label` | Display name |
| `field_type` | Storage type: text, number, boolean, datetime, reference, picklist |
| `field_subtype` | Semantic subtype: email, phone, url, currency, etc. (or null) |
| `is_required` | Mandatory on create/update |
| `is_read_only` | Cannot be modified by the user |
| `is_system_field` | System-managed field (Id, CreatedAt, OwnerId, etc.) |
| `sort_order` | Display order (system fields have negative values) |
| `config` | Field-specific configuration (max_length, precision, scale, default_value, picklist values) |

**Error responses:**

| HTTP Code | Condition |
|-----------|-----------|
| 401 | Not authenticated |
| 403 | User lacks OLS Read permission for this object |
| 404 | Object does not exist |

### 10.2 Record CRUD

All record operations use the same URL pattern: `/api/v1/records/{objectName}`.

#### List Records

```
GET /api/v1/records/{objectName}?page=1&per_page=20
```

Returns paginated records for the specified object. Security is enforced automatically — RLS filters out records the user cannot see, FLS filters out fields the user cannot read.

**Response:**
```json
{
  "data": [
    {
      "Id": "550e8400-e29b-41d4-a716-446655440000",
      "OwnerId": "550e8400-e29b-41d4-a716-446655440001",
      "CreatedAt": "2026-02-15T10:30:00Z",
      "UpdatedAt": "2026-02-15T14:45:00Z",
      "Name": "Acme Corp",
      "Industry": "Technology"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 45,
    "total_pages": 3
  }
}
```

#### Get a Single Record

```
GET /api/v1/records/{objectName}/{recordId}
```

Returns a single record by UUID. Returns `404` if the record doesn't exist or the user cannot see it (RLS).

**Response:**
```json
{
  "data": {
    "Id": "550e8400-e29b-41d4-a716-446655440000",
    "OwnerId": "550e8400-e29b-41d4-a716-446655440001",
    "CreatedAt": "2026-02-15T10:30:00Z",
    "UpdatedAt": "2026-02-15T14:45:00Z",
    "Name": "Acme Corp",
    "Industry": "Technology",
    "Revenue": 1000000
  }
}
```

#### Create a Record

```
POST /api/v1/records/{objectName}
Content-Type: application/json

{
  "Name": "New Corp",
  "Industry": "Finance",
  "Revenue": 5000000
}
```

System fields are injected automatically (see [10.3 System Fields](#103-system-fields)). Validation rules and dynamic defaults are applied before the record is saved.

**Response:**
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

#### Update a Record

```
PUT /api/v1/records/{objectName}/{recordId}
Content-Type: application/json

{
  "Name": "Updated Corp",
  "Revenue": 6000000
}
```

Partial update — only send fields that need to change. The `UpdatedById` field is set automatically to the current user.

**Response:**
```json
{
  "data": {
    "success": true
  }
}
```

#### Delete a Record

```
DELETE /api/v1/records/{objectName}/{recordId}
```

Hard delete (no soft delete). Returns `204 No Content` on success.

**Error responses (all CRUD operations):**

| HTTP Code | Condition |
|-----------|-----------|
| 400 | Invalid request body, UUID format, or type mismatch |
| 401 | Not authenticated |
| 403 | OLS/FLS/RLS permission denied |
| 404 | Object or record not found |

### 10.3 System Fields

Every record automatically includes 6 system fields. These fields are always present in Describe responses (with negative `sort_order`) and in record data.

| Field | Type | Writable | Description |
|-------|------|----------|-------------|
| `Id` | UUID | No | Unique record identifier, auto-generated |
| `OwnerId` | Reference | Yes | Record owner (defaults to current user on create) |
| `CreatedAt` | Datetime | No | Timestamp of record creation |
| `UpdatedAt` | Datetime | No | Timestamp of last update |
| `CreatedById` | Reference | No | User who created the record |
| `UpdatedById` | Reference | No | User who last updated the record |

**On create:** `OwnerId` is set to the current user if not provided. `CreatedById` and `UpdatedById` are always set to the current user (override attempts are ignored).

**On update:** `UpdatedById` is always set to the current user. `Id`, `CreatedAt`, and `CreatedById` cannot be changed.

### 10.4 Pagination

List endpoints support offset-based pagination:

| Parameter | Default | Min | Max | Description |
|-----------|---------|-----|-----|-------------|
| `page` | 1 | 1 | — | 1-based page number |
| `per_page` | 20 | 1 | 100 | Records per page |

The response includes a `pagination` object:

```json
{
  "pagination": {
    "page": 2,
    "per_page": 10,
    "total": 45,
    "total_pages": 5
  }
}
```

### 10.5 Security in Record Operations

Every record operation enforces all 3 security layers:

| Operation | OLS Check | FLS Check | RLS Check |
|-----------|-----------|-----------|-----------|
| List | Read | Read (fields filtered) | Records filtered |
| Get | Read | Read (fields filtered) | Record must be visible |
| Create | Create | Write (on each field) | Owner set automatically |
| Update | Update | Write (on each field) | Record must be accessible |
| Delete | Delete | — | Record must be accessible |

If any security check fails, the operation returns `403 Forbidden` without revealing which specific check failed.

---

## 11. Validation Rules & Dynamic Defaults

The platform supports declarative business logic through CEL (Common Expression Language) — no code required. Administrators can define validation rules that check data on every save, and dynamic defaults that compute field values automatically.

### 11.1 Validation Rules

A validation rule is a CEL expression that must evaluate to `true` for a record to be saved. If the expression evaluates to `false`, the save is blocked and the user sees the configured error message.

**Key properties:**

| Property | Description |
|----------|-------------|
| `api_name` | Unique identifier within the object (e.g., `name_required`) |
| `label` | Display name for the admin UI |
| `expression` | CEL expression that must return `true` for the record to be valid |
| `error_message` | User-facing message shown when the rule fails |
| `error_code` | Machine-readable error code (default: `validation_failed`) |
| `severity` | `error` (blocks save) or `warning` (reported but doesn't block) |
| `when_expression` | Optional CEL gate — the rule is only evaluated if this expression returns `true` |
| `applies_to` | When the rule is checked: `create`, `update`, or `create,update` (default) |
| `is_active` | Enable/disable the rule without deleting it |

**Example rules:**

| Rule | Expression | Error Message |
|------|-----------|---------------|
| Name required | `size(record.Name) > 0` | "Name field cannot be empty" |
| Amount positive | `record.Amount > 0` | "Amount must be greater than zero" |
| Email format | `record.Email.contains("@")` | "Invalid email format" |
| Status change only | `record.Status != old.Status` | "Status must change" (with `when_expression`: `has(old)`) |

**Validation semantics:** All active rules for the object are evaluated (AND semantics). If **any** rule with severity `error` fails, the entire save operation is blocked. Rules with severity `warning` are reported but do not prevent the save.

### 11.2 CEL Expression Language

CEL (Common Expression Language) is used for validation rules, dynamic defaults, conditional gates, and custom functions. Expressions are pure computations with no side effects.

#### Available Variables

| Variable | Type | Description | Available in |
|----------|------|-------------|-------------|
| `record` | map | Current record data (field values) | All contexts |
| `old` | map | Previous record data (before update) | Validation rules (UPDATE only) |
| `user` | map | Current user context | All contexts |
| `now` | timestamp | Current UTC timestamp | All contexts |

#### User Object Properties

```
user.id          — UUID of the current user
user.profile_id  — UUID of the user's profile
user.role_id     — UUID of the user's role (empty string if none)
```

#### Built-in Functions

| Function | Description | Example |
|----------|-------------|---------|
| `size(x)` | Length of string, list, or map | `size(record.Name) > 0` |
| `has(x)` | Check if field/variable exists | `has(old) && old.Status != record.Status` |
| `contains(s)` | String contains substring | `record.Email.contains("@")` |
| `startsWith(s)` | String starts with prefix | `record.Code.startsWith("ACC-")` |
| `endsWith(s)` | String ends with suffix | `record.File.endsWith(".pdf")` |
| `matches(re)` | Regex match | `record.Phone.matches("^\\+[0-9]+$")` |
| `int(x)` | Convert to integer | `int(record.Quantity)` |
| `double(x)` | Convert to float | `double(record.Price)` |
| `string(x)` | Convert to string | `string(now.year)` |
| `bool(x)` | Convert to boolean | `bool(record.IsActive)` |
| `timestamp(s)` | Parse ISO timestamp | `timestamp("2026-01-01T00:00:00Z")` |
| `duration(s)` | Parse duration | `duration("24h")` |

#### Expression Examples

```cel
// Numeric range
record.Amount > 0 && record.Amount <= 1000000

// Conditional validation
record.Type == "Premium" ? record.Amount >= 5000 : true

// Date comparison
now > timestamp("2026-01-01T00:00:00Z")

// Compare with previous value (UPDATE)
has(old) && record.Stage != old.Stage

// User-based logic
user.role_id != ""

// String validation
size(record.Name) >= 3 && size(record.Name) <= 100
```

### 11.3 Dynamic Defaults

Dynamic defaults automatically compute field values when a record is created or updated. They are configured as CEL expressions in the field definition (`default_expr` property).

**How defaults work:**

1. The system examines each field definition for the object.
2. If the field has a `default_expr` (CEL expression) or `default_value` (static value), and the field is not provided in the request, the default is applied.
3. Static defaults (`default_value`) are converted to the field's type (int, float, bool, datetime, etc.).
4. Dynamic defaults (`default_expr`) are evaluated as CEL expressions with access to `record`, `user`, and `now` variables.

**The `default_on` property** controls when defaults are applied:

| Value | Behavior |
|-------|----------|
| `create` | Default is applied only on INSERT (default) |
| `update` | Default is applied only on UPDATE |
| `create,update` | Default is applied on both |

**Default examples:**

| Field | default_expr | Description |
|-------|-------------|-------------|
| OwnerId | `user.id` | Set record owner to current user |
| CreatedAt | `now` | Set creation timestamp |
| Code | `"ACC-" + string(now.year)` | Generate code prefix |
| Priority | (default_value: `"Medium"`) | Static default value |

**Processing order:** Defaults are applied **before** validation rules. This means validation rules can check values set by defaults.

### 11.4 DML Pipeline

Every data write (INSERT, UPDATE, DELETE, UPSERT) passes through a multi-stage pipeline:

```
Parse → Resolve → Defaults → Validate → Compile → Execute
```

| Stage | Description |
|-------|-------------|
| **Parse** | Parse request JSON, extract field names and values |
| **Resolve** | Look up field types from metadata, verify object exists |
| **Defaults** | Apply static defaults (`default_value`) and dynamic defaults (`default_expr` via CEL) |
| **Validate** | Check required fields, type constraints, OLS/FLS permissions, then evaluate CEL validation rules |
| **Compile** | Generate parameterized SQL statement |
| **Execute** | Execute SQL with RLS enforcement, return results |

If any stage fails, the pipeline stops and returns an error. Validation rules that fail with severity `error` block the pipeline at the Validate stage.

### 11.5 Admin UI

Validation rules are managed through the admin panel at:

```
/admin/metadata/objects/{objectId}/validation-rules
```

**List view** — Shows all rules for the selected object with columns: api_name, label, severity badge, active/inactive badge.

**Create view** — Form with fields: api_name, label, expression (with Expression Builder), error_message, error_code, severity dropdown, when_expression (optional, with Expression Builder), applies_to, description.

**Detail view** — Same as create, with api_name disabled (immutable). Includes Save, Cancel, and Delete buttons.

The **Expression Builder** component provides a CodeMirror editor with CEL syntax highlighting, real-time expression validation, and error display.

### 11.6 CEL Validation Endpoint

Administrators can validate CEL expressions before saving rules:

```
POST /api/v1/admin/cel/validate
```

**Request:**
```json
{
  "expression": "size(record.Name) > 0",
  "context": "validation_rule",
  "object_api_name": "Account"
}
```

**Success response:**
```json
{
  "valid": true,
  "return_type": "bool"
}
```

**Error response:**
```json
{
  "valid": false,
  "errors": [
    {
      "message": "undeclared reference to 'recordd'",
      "line": 1,
      "column": 10
    }
  ]
}
```

Context values:

| Context | Variables | Purpose |
|---------|-----------|---------|
| `validation_rule` | record, old, user, now | Validate/when expressions in validation rules |
| `when_expression` | record, old, user, now | Conditional gate expressions |
| `default_expr` | record, user, now | Dynamic default expressions |
| `function_body` | function parameters only | Custom function body expressions |

---

## 12. Custom Functions

### 12.1 Overview

Custom Functions are named, reusable CEL expressions that can be called from any CEL context — validation rules, dynamic defaults, when-expressions, and other functions. They provide a way to encapsulate common business logic without duplicating expressions.

Every custom function:
- Has a unique name and is called via the `fn.*` namespace (e.g., `fn.discount(amount)`)
- Accepts typed parameters and returns a typed result
- Is a **pure computation** — no side effects, no database access, no external calls
- Works on both backend (cel-go) and frontend (cel-js) — **dual-stack**

### 12.2 Creating Functions

Functions are created through the admin panel at `/admin/functions` or via the REST API.

**Example: Discount calculator**

| Property | Value |
|----------|-------|
| Name | `discount` |
| Description | Calculates discount by amount |
| Parameters | `amount` (number) |
| Return type | `number` |
| Body | `amount > 1000 ? amount * 0.1 : 0.0` |

**Example: Premium check**

| Property | Value |
|----------|-------|
| Name | `is_premium` |
| Description | Checks premium customer status |
| Parameters | `total` (number), `count` (number) |
| Return type | `boolean` |
| Body | `total > 10000 && count > 5` |

### 12.3 fn.* Namespace

All custom functions are accessible via the `fn.` prefix. This prevents name collisions with built-in CEL functions.

**Usage in validation rules:**
```cel
fn.is_premium(record.TotalSales, record.OrderCount) && record.Status == "active"
```

**Usage in dynamic defaults:**
```cel
fn.discount(record.Amount)
```

**Nested calls (max 3 levels):**
```cel
fn.calculate_tier(fn.discount(amount))
```

Function name resolution: the system looks up the function name in the metadata cache and evaluates the body expression with the provided arguments bound to the parameter names.

### 12.4 Parameters & Return Types

**Parameter types:**

| Type | Description | Example |
|------|-------------|---------|
| `string` | Text value | `"hello"` |
| `number` | Integer or float | `42`, `3.14` |
| `boolean` | True/false | `true`, `false` |
| `list` | Array of values | `[1, 2, 3]` |
| `map` | Key-value object | `{"key": "value"}` |
| `any` | Any type (dynamic) | — |

**Return types:** Same as parameter types. The `any` type means the function can return different types depending on input.

Each function can have up to **10 parameters**. Parameters have a name, type, and optional description.

### 12.5 Dependency Management

Functions can call other functions, which creates dependencies. The platform enforces safety constraints:

- **No cycles:** `fn_a → fn_b → fn_a` is rejected. The system uses DFS cycle detection.
- **Max nesting depth: 3 levels.** `fn_a → fn_b → fn_c` is allowed, but adding a 4th level is rejected.
- **Usage tracking:** Before deleting a function, the system checks if it's referenced in:
  - Other function bodies (`metadata.functions.body`)
  - Validation rule expressions (`metadata.validation_rules.expression` and `when_expression`)
  - Field default expressions (`metadata.field_definitions.default_expr`)

If a function is referenced anywhere, deletion returns `409 Conflict` with details about where it's used.

### 12.6 Expression Builder

The Expression Builder is a shared UI component used across all CEL expression contexts (validation rules, defaults, functions). It provides:

- **CodeMirror editor** with CEL syntax highlighting, bracket matching, and undo/redo
- **Real-time validation** — expressions are validated via the `/api/v1/admin/cel/validate` endpoint as you type
- **Field picker** — suggests object field names (e.g., `record.Name`, `record.Amount`) when editing in an object context
- **Function picker** — lists all custom `fn.*` functions grouped by category, along with built-in string, type conversion, and time functions
- **Error display** — shows compilation errors with line/column location
- **Return type inference** — displays the detected return type of the expression

### 12.7 API

#### List Functions

```
GET /api/v1/admin/functions
```

**Response:**
```json
{
  "data": [
    {
      "id": "fn111111-1111-1111-1111-111111111111",
      "name": "discount",
      "description": "Calculates discount by amount",
      "params": [
        {"name": "amount", "type": "number", "description": "Amount in currency"}
      ],
      "return_type": "number",
      "body": "amount > 1000 ? amount * 0.1 : 0.0",
      "created_at": "2026-02-15T10:00:00Z",
      "updated_at": "2026-02-15T10:00:00Z"
    }
  ]
}
```

#### Create Function

```
POST /api/v1/admin/functions

{
  "name": "discount",
  "description": "Calculates discount by amount",
  "params": [
    {"name": "amount", "type": "number", "description": "Amount in currency"}
  ],
  "return_type": "number",
  "body": "amount > 1000 ? amount * 0.1 : 0.0"
}
```

Returns `201 Created` with the full function object.

#### Get Function

```
GET /api/v1/admin/functions/{functionId}
```

#### Update Function

```
PUT /api/v1/admin/functions/{functionId}

{
  "description": "Updated discount calculator",
  "body": "amount > 2000 ? amount * 0.15 : 0.0"
}
```

Note: the `name` field cannot be changed after creation.

#### Delete Function

```
DELETE /api/v1/admin/functions/{functionId}
```

Returns `204 No Content` on success. Returns `409 Conflict` if the function is referenced by other functions, validation rules, or field defaults.

**Error responses:**

| HTTP Code | Condition |
|-----------|-----------|
| 400 | Invalid name format, empty body, too many parameters |
| 404 | Function not found |
| 409 | Name already taken (create) or function is in use (delete) |

### 12.8 Limits

| Parameter | Limit | Description |
|-----------|-------|-------------|
| Function body size | 4 KB | Maximum CEL expression length |
| Parameters per function | 10 | Maximum number of input parameters |
| Nesting depth | 3 levels | Maximum fn.a → fn.b → fn.c chain |
| Total functions | 200 | Maximum number of functions per system |
| Name format | `^[a-z][a-z0-9_]*$` | Lowercase letters, digits, underscores |
| Execution timeout | 100 ms | Maximum evaluation time per function call |

---

## 13. Object Views

### 13.1 Overview

Object Views allow administrators to configure **role-based UI** for each object. Different profiles (Sales, Support, Management) can see different field sets and action buttons — all without code changes. Object Views also serve as a **bounded context adapter** (ADR-0022), encapsulating data contract logic (queries, computed fields, mutations, validation, defaults) alongside the presentation config. Related lists are deferred to the Layout layer (ADR-0027).

Every Object View is stored as a JSONB config in the `metadata.object_views` table with a unique `api_name`. Object Views are **not bound to a specific object** — they are standalone entities that can be used as page views (via navigation `page` items) or referenced by api_name. Optionally, an OV can be linked to a specific profile.

**Why profiles, not roles?** Object Views are bound to **profiles** rather than roles because profiles and roles serve fundamentally different purposes in the security model (ADR-0009):

- **Profile** defines *what a user can do* — OLS (CRUD on objects) and FLS (read/write on fields). It represents the user's **functional role**: Sales Rep, Support Agent, Manager. Since Object Views configure *which fields to display and how*, this directly aligns with FLS — the profile already determines which fields a user can access, so binding the UI configuration to the same entity ensures consistency. The resolved Object View config is intersected with the profile's FLS permissions (see [section 13.6](#136-fls-intersection)).
- **Role** defines *what a user can see* — the position in the organizational hierarchy used for RLS (record visibility via role hierarchy, sharing rules). A "Sales Manager" and "Sales Director" may share the same UI but see different sets of records. Binding Object Views to roles would conflate presentation with data visibility.

This follows the Salesforce pattern: page layouts (the analog of Object Views) are assigned per profile, not per role.

Key capabilities:

**Read (`read`):**
- **Fields** — flat list of field `api_name` values to include in this view (WHAT to show). Sections and highlight fields are auto-generated from this list in the computed form.
- **Actions** — custom buttons with CEL visibility expressions (e.g., show "Send Proposal" only when `record.Status == 'draft'`)
- **Queries** — named SOQL queries scoped to this Object View context
- **Computed** — computed fields derived from CEL expressions, scoped to this view (display-only, not persisted)

**Write (`write`, optional):**
- **Validation** — view-scoped validation rules (additive with metadata-level rules)
- **Defaults** — view-scoped default expressions (replace metadata-level defaults)
- **Computed** — fields whose values are computed from CEL expressions on save
- **Mutations** — DML operations scoped to this Object View context

### 13.2 Creating an Object View

Object Views are created through the admin panel at `/admin/metadata/object-views` or via the REST API.

**Admin UI:**
1. Navigate to **Object Views** (`/admin/metadata/object-views`).
2. Click the **"+"** button.
3. Fill in the required fields:
   - **API Name:** `account_sales_view` (lowercase with underscores, unique)
   - **Label:** "Account Sales View"
   - **Profile:** Optionally select a profile (e.g., Sales). Leave empty for a global view.
   - **Is Default:** Check if this should be the default view.
4. Click **"Create"**. You are redirected to the visual constructor.

### 13.3 Config Structure

The Object View config is a JSON object with two top-level sections: `read` (presentation and read-time data contract) and `write` (write-time data contract):

```json
{
  "read": {
    "fields": ["Name", "Industry", "Phone", "Revenue", "AnnualBudget"],
    "actions": [
      {
        "key": "send_proposal",
        "label": "Send Proposal",
        "type": "primary",
        "icon": "mail",
        "visibility_expr": "record.Status == 'draft'"
      },
      {
        "key": "mark_urgent",
        "label": "Mark Urgent",
        "type": "danger",
        "icon": "alert-triangle",
        "visibility_expr": "record.Priority != 'high'"
      }
    ],
    "queries": [
      {
        "name": "recent_activities",
        "soql": "SELECT Id, Subject, Type FROM Activity WHERE AccountId = :recordId ORDER BY CreatedAt DESC LIMIT 5",
        "when": "record.status == 'active'"
      }
    ],
    "computed": [
      {
        "name": "total_with_tax",
        "type": "float",
        "expr": "record.amount * 1.2"
      }
    ]
  },
  "write": {
    "validation": [
      {
        "expr": "record.amount > 0",
        "message": "Amount must be positive",
        "code": "invalid_amount",
        "severity": "error"
      }
    ],
    "defaults": [
      {
        "field": "status",
        "expr": "'draft'",
        "on": "create"
      }
    ],
    "computed": [
      {
        "field": "total_with_tax",
        "expr": "record.amount * (1 + record.tax_rate / 100)"
      }
    ],
    "mutations": [
      {
        "dml": "UPDATE Account SET last_contacted_at = now() WHERE id = :recordId",
        "when": "record.status == 'active'"
      }
    ]
  }
}
```

**Property reference:**

**Read properties (`read.*`):**

| Property | Type | Description |
|----------|------|-------------|
| `read.fields` | string[] | Field `api_name` values included in this view. Order matters — first 3 are used as highlights in the computed form. |
| `read.actions` | array | Custom action buttons |
| `read.actions[].key` | string | Unique action identifier |
| `read.actions[].label` | string | Button text |
| `read.actions[].type` | string | `primary`, `secondary`, or `danger` |
| `read.actions[].icon` | string | Lucide icon name (e.g., `mail`, `check`, `alert-triangle`) |
| `read.actions[].visibility_expr` | string | CEL expression evaluated against the current record |
| `read.queries` | array | Named SOQL queries scoped to this Object View context |
| `read.queries[].name` | string | Query identifier (e.g., `recent_activities`) |
| `read.queries[].soql` | string | SOQL query with `:recordId` parameter binding |
| `read.queries[].when` | string | Optional CEL condition for when query executes |
| `read.computed` | array | Computed fields (read) — derived from CEL expressions, display-only |
| `read.computed[].name` | string | Computed field name |
| `read.computed[].type` | string | `string`, `int`, `float`, `bool`, or `timestamp` |
| `read.computed[].expr` | string | CEL expression computing the value |
| `read.computed[].when` | string | Optional CEL condition for when field applies |

**Write properties (`write.*`):**

| Property | Type | Description |
|----------|------|-------------|
| `write.validation` | array | View-scoped validation rules (additive with metadata-level rules) |
| `write.validation[].expr` | string | CEL expression that must evaluate to `true` |
| `write.validation[].message` | string | Error message shown when validation fails |
| `write.validation[].code` | string | Optional error code identifier |
| `write.validation[].severity` | string | `error` (blocks save) or `warning` (advisory) |
| `write.validation[].when` | string | Optional CEL condition for when rule applies |
| `write.defaults` | array | View-scoped default expressions (replace metadata-level defaults) |
| `write.defaults[].field` | string | Target field `api_name` |
| `write.defaults[].expr` | string | CEL expression computing the default value |
| `write.defaults[].on` | string | `create`, `update`, or `create,update` |
| `write.defaults[].when` | string | Optional CEL condition for when default applies |
| `write.computed` | array | Computed fields (write) — values computed from CEL expressions on save |
| `write.computed[].field` | string | Target field `api_name` |
| `write.computed[].expr` | string | CEL expression computing the value |
| `write.mutations` | array | DML operations scoped to this Object View |
| `write.mutations[].dml` | string | DML statement with `:recordId` parameter binding |
| `write.mutations[].foreach` | string | Optional CEL expression for iteration (e.g., `queries.line_items`) |
| `write.mutations[].sync` | object | Optional sync mapping (`key`, `value` fields) |
| `write.mutations[].when` | string | Optional CEL condition for when mutation executes |

### 13.4 Visual Constructor

The Object View detail page provides a tab-based visual constructor for editing the config without writing JSON. Tabs are organized as follows:

**Read tabs:**

1. **General** — edit label, description, is_default flag. Read-only: api_name, profile.
2. **Fields** — add/remove field api_names to include in this view. Order matters — first 3 become highlights in the computed form.
3. **Actions** — add action buttons with key, label, type, icon, and a CEL visibility expression (uses the Expression Builder from Phase 8).
4. **Queries** — define named SOQL queries scoped to this view (name, SOQL statement, optional `when` condition). Each query's SOQL is edited via the **SOQL Editor** — a rich CodeMirror-based editor with syntax highlighting, context-aware autocomplete (objects, fields, keywords, functions, date literals), server-side validation (`POST /admin/soql/validate`), and test query execution (preview first 5 records).
5. **Computed (Read)** — define computed fields from CEL expressions (name, type, expression, optional `when` condition). These are display-only and not persisted.

**Write tabs:**

6. **Validation** — define view-scoped validation rules (CEL expression, error message, optional code, severity: error/warning, optional `when` condition). These are additive with metadata-level validation rules.
7. **Defaults** — define view-scoped default expressions (field, CEL expression, trigger: create/update/create,update, optional `when` condition). These replace metadata-level defaults.
8. **Computed (Write)** — define fields whose values are computed from CEL expressions on save (field, CEL expression).
9. **Mutations** — define DML operations scoped to this view (DML statement, optional `foreach` iteration, optional `sync` mapping, optional `when` condition).

Click **"Save"** to persist changes. All changes take effect immediately.

### 13.5 View Resolution

Object Views are accessed directly by `api_name` via the View endpoint:

```
GET /api/v1/view/:ovApiName
```

This returns the OV config with FLS intersection applied. Navigation items of type `page` reference an OV by `ov_api_name` — clicking such an item loads the OV as a standalone page.

The Describe API (`GET /api/v1/describe/:objectName`) **always returns a fallback form** — auto-generated from all FLS-accessible fields (one "Details" section, first 3 fields as highlights, first 5 as list columns). The Describe API no longer resolves Object Views.

This allows gradual adoption — the system works without any Object Views configured, and administrators can add views incrementally via navigation page items.

### 13.6 FLS Intersection

The resolved Object View config is intersected with the current user's Field-Level Security (FLS) permissions:

- **Fields** — fields the user cannot read are removed from the flat field list (and consequently from the auto-generated sections and highlights)
- **List fields** — inaccessible columns are removed
- **Related lists** — related objects the user cannot read (OLS) are excluded

This ensures that even if an administrator includes a field in the Object View, users without FLS access will never see it.

### 13.7 Describe API (Fallback Form)

The Describe API always returns a **fallback form** — auto-generated from all FLS-accessible fields. It no longer resolves Object Views.

```
GET /api/v1/describe/{objectName}
```

**Response (always fallback):**
```json
{
  "data": {
    "api_name": "Account",
    "label": "Account",
    "plural_label": "Accounts",
    "is_createable": true,
    "is_updateable": true,
    "is_deleteable": true,
    "fields": [...],
    "form": {
      "sections": [
        {
          "key": "details",
          "label": "Details",
          "columns": 2,
          "collapsed": false,
          "fields": ["Name", "Industry", "Phone", "Revenue"]
        }
      ],
      "highlight_fields": ["Name", "Industry", "Phone"],
      "list_fields": ["Name", "Industry", "Phone", "Revenue", "CreatedAt"],
      "list_default_sort": "created_at DESC"
    }
  }
}
```

The `fields` array remains for backward compatibility. The `form` property is always present and always auto-generated from FLS-accessible fields.

To get a customized view configuration, use the View endpoint: `GET /api/v1/view/:ovApiName`.

### 13.8 CRM UI Rendering

When the CRM frontend (`/app/*`) receives a `form` in the Describe response, it renders:

- **Record detail page:**
  - Highlight fields at the top in a compact card
  - Action buttons (filtered by `visibility_expr` evaluated via cel-js)
  - Collapsible sections with fields arranged in the specified column layout
- **Record create page:**
  - Sections with fields (without highlights or actions)

- **Record list page:**
  - Columns from `list_fields` with `list_default_sort` as the default sort order

If no `form` is present (backward compatibility), the UI falls back to the original single-card layout with all fields.

### 13.9 API

#### List Object Views

```
GET /api/v1/admin/object-views
```

Returns all Object Views.

**Response:**
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "profile_id": null,
      "api_name": "account_default",
      "label": "Account Default View",
      "description": "Default view for all users",
      "is_default": true,
      "config": { "read": { "fields": [...], ... }, "write": { ... } },
      "created_at": "2026-02-17T10:00:00Z",
      "updated_at": "2026-02-17T10:00:00Z"
    }
  ]
}
```

#### Get a Single Object View

```
GET /api/v1/admin/object-views/{viewId}
```

#### Create an Object View

```
POST /api/v1/admin/object-views

{
  "profile_id": null,
  "api_name": "account_default",
  "label": "Account Default View",
  "description": "Default view for all users",
  "is_default": true,
  "config": {
    "read": {
      "fields": [],
      "actions": [],
      "queries": [],
      "computed": []
    },
    "write": {
      "validation": [],
      "defaults": [],
      "computed": [],
      "mutations": []
    }
  }
}
```

Returns `201 Created` with the created Object View.

#### Update an Object View

```
PUT /api/v1/admin/object-views/{viewId}

{
  "label": "Updated Account View",
  "description": "Updated description",
  "is_default": true,
  "config": { ... }
}
```

Note: `api_name` and `profile_id` cannot be changed after creation.

#### Delete an Object View

```
DELETE /api/v1/admin/object-views/{viewId}
```

Returns `204 No Content` on success.

**Error responses:**

| HTTP Code | Condition |
|-----------|-----------|
| 400 | Invalid api_name format, missing required fields |
| 404 | Object View not found |
| 409 | Duplicate api_name |

---

## 14. Procedures

### 14.1 Overview

Procedures are named business logic sequences described in JSON. They allow administrators to automate multi-step operations — creating records, performing validations, calling external APIs, branching on conditions — all without writing Go code.

Key features:
- **Visual Constructor UI** — build procedures via forms and dropdowns, no raw JSON editing required.
- **Versioning** — each procedure has a draft and optionally a published version. Changes are made to the draft; publishing promotes it to live.
- **Dry Run** — test a procedure without side effects before publishing.
- **Command types** — record operations (`record.create`, `record.update`, `record.delete`, `record.get`, `record.query`), computations (`compute.transform`, `compute.validate`, `compute.fail`), flow control (`flow.if`, `flow.match`, `flow.call`, `flow.try`), HTTP integrations (`integration.http`), and stubs for future notification/wait commands.
- **Retry** — любая команда может быть сконфигурирована с автоматическим retry (до 5 попыток, задержка с экспоненциальным backoff). Критично для нестабильных внешних API.
- **Try/Catch** — `flow.try` позволяет перехватить ошибку, выполнить recovery-логику и продолжить процедуру. Переменная `$.error` доступна в catch-блоке.
- **Saga rollback** — if a command fails, previously completed commands with rollback definitions are undone in LIFO order.
- **Security** — record commands go through the standard SOQL/DML security layers (OLS, FLS, RLS). HTTP integrations use Named Credentials.

### 14.2 Creating a Procedure

**Admin UI:**
1. Navigate to **Procedures** (`/admin/metadata/procedures`).
2. Click the **"+"** button.
3. Fill in:
   - **Code:** `create_account_workflow` (lowercase with underscores, must start with a letter)
   - **Name:** "Create Account Workflow"
   - **Description:** (optional) "Creates account and sends welcome notification"
4. Click **"Create"**. You are redirected to the detail page with an empty draft (v1).

**API:**
```
POST /api/v1/admin/procedures

{
  "code": "create_account_workflow",
  "name": "Create Account Workflow",
  "description": "Creates account and sends welcome notification"
}
```

Returns `201 Created` with the procedure and its draft version.

### 14.3 Definition & Commands

A procedure definition is a JSON object with a list of commands and an optional result mapping:

```json
{
  "commands": [
    {
      "type": "record.create",
      "as": "account",
      "object": "Account",
      "data": {
        "Name": "$.input.name",
        "Industry": "$.input.industry"
      }
    },
    {
      "type": "compute.validate",
      "as": "check_name",
      "condition": "$.input.name != ''",
      "code": "name_required",
      "message": "Name is required"
    },
    {
      "type": "flow.if",
      "as": "branch",
      "condition": "$.input.sendWelcome == true",
      "then": [
        {
          "type": "integration.http",
          "as": "welcome_call",
          "credential": "slack_webhook",
          "method": "POST",
          "path": "/api/notify",
          "body": "{\"text\": \"New account: $.account.id\"}"
        }
      ],
      "else": []
    }
  ],
  "result": {
    "accountId": "$.account.id"
  }
}
```

**Command reference:**

| Type | Description | Key Fields |
|------|-------------|------------|
| `record.create` | Insert a new record | `object`, `data` (field→expression map) |
| `record.update` | Update an existing record | `object`, `id` (expression), `data` |
| `record.delete` | Delete a record | `object`, `id` (expression) |
| `record.get` | Fetch a single record by ID | `object`, `id` (expression) |
| `record.query` | Execute a SOQL query | `query` (SOQL string) |
| `compute.transform` | Map/compute values | `value` (key→expression map) |
| `compute.validate` | Assert a condition | `condition` (CEL), `code`, `message` |
| `compute.fail` | Raise an error immediately | `code`, `message` |
| `flow.if` | Conditional branching | `condition` (CEL), `then` (commands), `else` (commands) |
| `flow.match` | Switch/case branching | `expression` (CEL), `cases` (key→commands map) |
| `flow.call` | Call another procedure | `procedure` (code), `input` (expression map) |
| `flow.try` | Try/Catch error handling | `try` (commands), `catch` (commands) |
| `integration.http` | HTTP request via Named Credential | `credential`, `method`, `path`, `headers`, `body` |

**Common command fields:**

| Field | Type | Description |
|-------|------|-------------|
| `as` | string | Variable name for the step result (accessible as `$.<as>`) |
| `when` | string | Optional CEL condition — skip if evaluates to `false` |
| `optional` | bool | If `true`, errors are captured as warnings instead of aborting |
| `rollback` | array | List of compensating commands, executed in order if a later step fails |
| `retry` | object | Retry config: `max_attempts` (1–5), `delay_ms` (100–60000), `backoff_mult` (multiplier, default 1) |

**Saga Rollback:**

Each command can define a `rollback` — a list of compensating commands that run if a **later** step fails. Rollbacks execute in LIFO order (last registered → first executed), following the Saga pattern. Each rollback entry can contain multiple commands:

```json
{
  "commands": [
    {
      "type": "record.create",
      "as": "order",
      "object": "Order",
      "data": { "Status": "'pending'", "Amount": "$.input.amount" },
      "rollback": [
        {
          "type": "record.delete",
          "object": "Order",
          "id": "$.order.id"
        }
      ]
    },
    {
      "type": "integration.http",
      "as": "payment",
      "credential": "stripe_api",
      "method": "POST",
      "path": "/v1/charges",
      "body": "{\"amount\": \"$.input.amount\"}",
      "rollback": [
        {
          "type": "integration.http",
          "credential": "stripe_api",
          "method": "POST",
          "path": "/v1/refunds",
          "body": "{\"charge\": \"$.payment.id\"}"
        },
        {
          "type": "compute.transform",
          "as": "log_refund",
          "value": { "refunded": "true" }
        }
      ]
    },
    {
      "type": "integration.http",
      "credential": "email_service",
      "method": "POST",
      "path": "/send",
      "body": "{\"to\": \"$.input.email\", \"template\": \"order_confirmation\"}"
    }
  ]
}
```

If the email step fails:
1. Payment rollback runs (refund via Stripe).
2. Order rollback runs (delete the order record).

Rollback is only registered for commands that **succeeded**. If a command fails before completing, its rollback is not added to the stack.

**Retry:**

Любая команда может быть сконфигурирована с автоматическим retry. Это особенно полезно для `integration.http` команд, вызывающих нестабильные внешние API.

```json
{
  "type": "integration.http",
  "as": "payment",
  "credential": "stripe_api",
  "method": "POST",
  "path": "/v1/charges",
  "body": "{\"amount\": \"$.input.amount\"}",
  "retry": {
    "max_attempts": 3,
    "delay_ms": 1000,
    "backoff_mult": 2
  }
}
```

Параметры retry:
- `max_attempts` (1–5) — максимальное количество попыток (включая первую).
- `delay_ms` (100–60000) — задержка перед повторной попыткой (мс).
- `backoff_mult` (default 1) — множитель задержки после каждой неудачной попытки. При `backoff_mult: 2` и `delay_ms: 1000` задержки будут: 1s, 2s, 4s…

Поведение:
- Если команда успешна — retry не нужен, результат возвращается сразу.
- Если все попытки неудачны — последняя ошибка пробрасывается.
- Каждая повторная попытка добавляет запись в trace со статусом `"retry"`.
- Retry учитывает deadline выполнения: если задержка превысит оставшееся время, retry прекращается с ошибкой.

**Try/Catch (`flow.try`):**

`flow.try` позволяет перехватить ошибку команды и выполнить recovery-логику вместо прерывания всей процедуры. Это промежуточный вариант между `optional: true` (игнорировать ошибку) и стандартным поведением (прервать процедуру).

```json
{
  "type": "flow.try",
  "as": "safe_call",
  "try": [
    {
      "type": "integration.http",
      "as": "api_call",
      "credential": "external_api",
      "method": "POST",
      "path": "/process"
    }
  ],
  "catch": [
    {
      "type": "compute.transform",
      "as": "fallback",
      "value": {
        "failed": "'true'",
        "error_msg": "$.error.message"
      }
    }
  ]
}
```

Семантика:
1. Выполняются команды из `try` блока.
2. Если `try` успешен — `catch` не выполняется. Результат: `{"caught": false}`.
3. Если `try` падает — ошибка сохраняется в `$.error` (объект с полями `code` и `message`), затем выполняются команды из `catch` блока.
4. Если `catch` успешен — процедура продолжается. Результат: `{"caught": true, "error_code": "...", "error_message": "..."}`.
5. Если `catch` тоже падает — ошибка пробрасывается выше (как если бы `flow.try` не было).

Переменная `$.error` доступна только внутри `catch`-блока:
- `$.error.code` — код ошибки (из `ExecutionError` или `AppError`, иначе `"unknown"`).
- `$.error.message` — текст ошибки.

Переменные, установленные в `try`-блоке (через `as`), остаются доступны после `flow.try`, даже если произошла ошибка.

**Expression resolution:**

All string values starting with `$.` are resolved at runtime:
- `$.input.<field>` — input parameter value
- `$.user.<field>` — current user context
- `$.now` — current timestamp
- `$.<step_name>` — result of a previous command (by `as` name)
- `$.error` — объект ошибки (только внутри `catch`-блока `flow.try`): `$.error.code`, `$.error.message`

**Пример: record.query и работа с результатами**

`record.query` возвращает массив записей (`[]map`). Каждый элемент — одна запись, ключи — имена полей из SOQL-запроса. Результат доступен через `as` как `$.<step_name>`:

```json
{
  "commands": [
    {
      "type": "record.query",
      "as": "deals",
      "query": "SELECT Id, Name, Amount FROM Opportunity WHERE StageName = 'Closed Won' AND OwnerId = '$.input.user_id'"
    },
    {
      "type": "compute.validate",
      "condition": "size(deals) > 0",
      "code": "no_deals",
      "message": "No closed deals found for this user"
    },
    {
      "type": "compute.transform",
      "as": "summary",
      "value": {
        "count": "size(deals)",
        "first_deal": "deals[0].Name",
        "first_amount": "deals[0].Amount"
      }
    },
    {
      "type": "record.update",
      "object": "Contact",
      "id": "$.input.contact_id",
      "data": {
        "LastDealName": "$.deals[0].Name",
        "DealCount": "$.summary.count"
      }
    }
  ],
  "result": {
    "dealCount": "$.summary.count",
    "deals": "$.deals"
  }
}
```

Доступ к результатам `record.query`:
- `$.deals` — весь массив записей, например `[{"Id": "...", "Name": "Acme", "Amount": 50000}, ...]`
- `$.deals[0].Name` — поле первой записи (в `data` record-команд, `$.`-синтаксис)
- `deals[0].Name` — то же в CEL-выражениях (в `value`, `condition` — без `$.`)
- `size(deals)` — количество записей (CEL-функция)

### 14.4 Versioning

Procedures use a draft/published versioning model (ADR-0029):

- **Draft** — editable work-in-progress. Save anytime without affecting live execution.
- **Published** — the live version used when executing the procedure.
- **Superseded** — archived previous published versions (up to 10 kept).

**Workflow:**

1. Create a procedure → draft v1 is created automatically.
2. Edit the definition → save draft (can save multiple times).
3. Publish → draft becomes published, previous published becomes superseded.
4. Continue editing → create a new draft from the current published version.
5. Rollback → revert to the previous published version.

**API:**

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/admin/procedures/:id/draft` | PUT | Save draft definition |
| `/api/v1/admin/procedures/:id/draft` | DELETE | Discard draft |
| `/api/v1/admin/procedures/:id/publish` | POST | Publish draft |
| `/api/v1/admin/procedures/:id/rollback` | POST | Rollback to previous published |
| `/api/v1/admin/procedures/:id/versions` | GET | List version history |

### 14.5 Dry Run & Execution

**Dry run** tests a procedure without side effects. Record mutations return fake UUIDs, HTTP calls are skipped.

```
POST /api/v1/admin/procedures/:id/dry-run

{
  "input": {
    "name": "Acme Corp",
    "industry": "Technology",
    "sendWelcome": true
  }
}
```

**Response:**
```json
{
  "data": {
    "success": true,
    "result": {
      "accountId": "00000000-0000-0000-0000-000000000000"
    },
    "warnings": [],
    "trace": [
      {"command": "record.create", "as": "account", "duration_ms": 0, "status": "ok"},
      {"command": "flow.if", "as": "branch", "duration_ms": 0, "status": "ok"}
    ]
  }
}
```

**Execute** runs the published version with real side effects:

```
POST /api/v1/admin/procedures/:id/execute

{
  "input": {
    "name": "Acme Corp",
    "industry": "Technology"
  }
}
```

### 14.6 Constructor UI

The procedure detail page provides a visual Constructor UI for building definitions:

1. **Command List** — ordered list of command cards. Each card shows the command type (color-coded badge), key fields, and action buttons (move up/down, remove).
2. **Command Picker** — the "+" button opens a categorized dropdown: Record, Compute, Flow, Integration, Notification, Wait. Select a command type to append it.
3. **Command Editor** — each card expands to show type-specific form fields:
   - `record.query`: SOQL Query textarea only (Object is inferred from the query itself).
   - `record.create`: Object picker, data mapping table.
   - `record.update`/`record.delete`/`record.get`: Object picker, Record ID expression.
   - Compute: CEL expression editors (condition, value mappings).
   - Flow: nested command lists for branches (then/else, cases). `flow.try` shows try/catch info.
   - Integration: credential picker, HTTP method, path, headers, body.
4. **Common fields** — each command has optional fields: `as` (variable name), `when` (condition), `optional` toggle.
5. **Retry config** — each command can enable retry with configurable max attempts (1–5), delay (ms), and backoff multiplier.
6. **Tabs** — Definition (Constructor), Versions (history), Settings (metadata), Dry Run (test panel).

### 14.7 API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/admin/procedures` | POST | Create procedure (with draft v1) |
| `/api/v1/admin/procedures` | GET | List all procedures |
| `/api/v1/admin/procedures/:id` | GET | Get procedure with versions |
| `/api/v1/admin/procedures/:id` | PUT | Update metadata (name, description) |
| `/api/v1/admin/procedures/:id` | DELETE | Delete procedure |
| `/api/v1/admin/procedures/:id/draft` | PUT | Save draft definition |
| `/api/v1/admin/procedures/:id/draft` | DELETE | Discard draft |
| `/api/v1/admin/procedures/:id/publish` | POST | Publish draft |
| `/api/v1/admin/procedures/:id/rollback` | POST | Rollback to previous published |
| `/api/v1/admin/procedures/:id/versions` | GET | Version history |
| `/api/v1/admin/procedures/:id/execute` | POST | Execute published version |
| `/api/v1/admin/procedures/:id/dry-run` | POST | Dry-run (draft if exists, else published) |

**Error responses:**

| HTTP Code | Condition |
|-----------|-----------|
| 400 | Invalid code format, definition too large, unknown command type |
| 404 | Procedure not found |
| 409 | Duplicate code, no draft to publish, no version to rollback |

### 14.8 Limits

| Parameter | Limit |
|-----------|-------|
| Execution timeout | 30 seconds |
| Max commands per execution | 50 |
| Max call depth (`flow.call`) | 3 |
| Max if/match/try nesting | 5 |
| Max definition JSON size | 64 KB |
| Max input size | 1 MB |
| Max HTTP calls per execution | 10 |
| Max notifications per execution | 10 |
| Retry: max attempts | 5 |
| Retry: delay range | 100–60000 ms |

---

## 15. Named Credentials

### 15.1 Overview

Named Credentials provide secure, encrypted storage for authentication secrets used in HTTP integrations (`integration.http` commands in Procedures). They allow administrators to configure API access without exposing secrets in procedure definitions.

Key features:
- **AES-256-GCM encryption** — all secrets are encrypted at rest with a unique nonce per record.
- **Three auth types** — API Key, Basic Auth, OAuth2 Client Credentials.
- **SSRF protection** — all HTTP requests are validated against the credential's base URL (HTTPS only, no internal IPs).
- **Test connection** — verify that a credential works before using it in procedures.
- **Usage audit log** — every HTTP request made through a credential is logged with URL, status, duration, and the calling procedure.

### 15.2 Credential Types

| Type | Auth Mechanism | Fields |
|------|---------------|--------|
| `api_key` | Custom header with API key | `header` (header name), `value` (key value) |
| `basic` | HTTP Basic Authentication | `username`, `password` |
| `oauth2_client` | OAuth2 Client Credentials Grant | `client_id`, `client_secret`, `token_url`, `scope` |

For `api_key`, the system sends the configured header (e.g., `X-API-Key: sk-abc123`).
For `basic`, the system sends `Authorization: Basic <base64(username:password)>`.
For `oauth2_client`, the system obtains an access token via the client credentials flow and sends `Authorization: Bearer <token>`. Tokens are cached and auto-refreshed.

### 15.3 Creating a Credential

**Admin UI:**
1. Navigate to **Credentials** (`/admin/metadata/credentials`).
2. Click the **"+"** button.
3. Fill in:
   - **Code:** `stripe_api` (lowercase with underscores, must start with a letter)
   - **Name:** "Stripe API"
   - **Base URL:** `https://api.stripe.com` (must be HTTPS)
   - **Type:** Select `api_key`, `basic`, or `oauth2_client`
4. Fill in type-dependent auth fields:
   - For `api_key`: Header = `Authorization`, Value = `Bearer sk_live_abc123`
   - For `basic`: Username = `api`, Password = `sk_live_abc123`
   - For `oauth2_client`: Client ID, Client Secret, Token URL, Scope
5. Click **"Create"**.

**API:**
```
POST /api/v1/admin/credentials

{
  "code": "stripe_api",
  "name": "Stripe API",
  "base_url": "https://api.stripe.com",
  "type": "api_key",
  "auth_data": {
    "header": "Authorization",
    "value": "Bearer sk_live_abc123"
  }
}
```

Returns `201 Created`. Note: secrets are encrypted on write and never returned in plaintext — API responses show masked values.

### 15.4 Test Connection

The Test Connection feature sends a GET request to the credential's base URL with the configured authentication and reports success or failure.

**Admin UI:**
1. Open a credential's detail page.
2. Click the **Test Connection** button (Wifi icon).
3. The result shows: HTTP status, response time, and success/failure indication.

**API:**
```
POST /api/v1/admin/credentials/:id/test
```

Returns `200 OK` with test result including `success`, `status_code`, and `duration_ms`.

### 15.5 Usage Log

Every HTTP request made through a credential is recorded in the usage log. The log captures:

| Field | Description |
|-------|-------------|
| `procedure_code` | Which procedure made the request |
| `request_url` | Full URL of the HTTP request |
| `response_status` | HTTP response status code |
| `success` | Whether the request succeeded |
| `error_message` | Error details (if failed) |
| `duration_ms` | Request duration in milliseconds |
| `user_id` | User who triggered the execution |
| `created_at` | Timestamp |

**Admin UI:**
1. Open a credential's detail page.
2. Switch to the **Usage** tab.
3. View the log entries with timestamps, URLs, status codes, and duration.

**API:**
```
GET /api/v1/admin/credentials/:id/usage
```

### 15.6 API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/admin/credentials` | POST | Create credential (secrets encrypted) |
| `/api/v1/admin/credentials` | GET | List all (secrets masked) |
| `/api/v1/admin/credentials/:id` | GET | Get credential (secrets masked) |
| `/api/v1/admin/credentials/:id` | PUT | Update credential |
| `/api/v1/admin/credentials/:id` | DELETE | Delete credential (409 if used by procedure) |
| `/api/v1/admin/credentials/:id/test` | POST | Test connection |
| `/api/v1/admin/credentials/:id/usage` | GET | Usage audit log |
| `/api/v1/admin/credentials/:id/deactivate` | POST | Deactivate credential |
| `/api/v1/admin/credentials/:id/activate` | POST | Activate credential |

**Error responses:**

| HTTP Code | Condition |
|-----------|-----------|
| 400 | Invalid code format, non-HTTPS base_url, unknown type |
| 404 | Credential not found |
| 409 | Duplicate code, credential in use (on delete) |

**Security considerations:**
- The `CREDENTIAL_ENCRYPTION_KEY` environment variable (32 bytes) is required for encryption.
- Secrets are never returned in API responses — only masked values (e.g., `***`).
- Base URL must use HTTPS. Internal IPs (127.0.0.0/8, 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, ::1) are blocked.

---

## 16. Profile Navigation

### 16.1 Overview

Profile Navigation (ADR-0032) allows administrators to configure per-profile sidebar navigation. Each profile can have its own set of grouped navigation items instead of the default flat alphabetical list.

Key features:
- **Grouped navigation** — items are organized into collapsible groups (e.g., "Sales", "Support").
- **Four item types** — `object` (links to an object's record list), `link` (external or internal URL), `page` (renders an Object View as a standalone page via `ov_api_name`), `divider` (visual separator).
- **OLS intersection** — object items are filtered by the user's read permissions. Objects the user cannot read are hidden.
- **Fallback** — if no navigation config exists for the user's profile, the sidebar shows an OLS-filtered alphabetical list of all queryable objects (current default behavior).
- **One config per profile** — `UNIQUE(profile_id)` constraint. `ON DELETE CASCADE` when the profile is deleted.

### 16.2 Navigation Config

The navigation config is stored as JSONB in `metadata.profile_navigation`. Structure:

```json
{
  "groups": [
    {
      "key": "sales",
      "label": "Sales",
      "icon": "trending-up",
      "items": [
        { "type": "object", "object_api_name": "Account" },
        { "type": "object", "object_api_name": "Opportunity" },
        { "type": "divider" },
        { "type": "page", "label": "Sales Dashboard", "ov_api_name": "sales_dashboard", "icon": "layout-dashboard" },
        { "type": "link", "label": "Reports", "url": "/app/reports", "icon": "bar-chart" }
      ]
    }
  ]
}
```

**Item types:**

| Type | Required fields | Description |
|------|----------------|-------------|
| `object` | `object_api_name` | Links to object record list (`/app/{objectApiName}`) |
| `link` | `label`, `url` | External or internal URL |
| `page` | `label`, `ov_api_name` | Renders Object View as a standalone page (`/app/page/{ovApiName}`) |
| `divider` | — | Visual separator |

**Validation rules:**
- Max 20 groups
- Max 50 items per group
- Group keys must be unique
- Item types must be `object`, `link`, `page`, or `divider`
- URLs in link items: no `javascript:` scheme allowed
- `ov_api_name` in page items must reference an existing Object View

### 16.3 Resolution Logic

`GET /api/v1/navigation` resolves the sidebar for the authenticated user:

1. Look up the user's `profile_id` from the JWT-derived `UserContext`.
2. Query `metadata.profile_navigation` for a config matching this `profile_id`.
3. If found: iterate groups → items, filter object items by OLS (remove objects the user cannot read), resolve labels from metadata.
4. If not found: build a fallback — single group with `key="_default"`, empty `label`, containing all OLS-accessible objects alphabetically.

### 16.4 Admin UI

Admin views for managing navigation configs:

| View | Route | Description |
|------|-------|-------------|
| List | `/admin/metadata/navigation` | Shows all configs with profile ID and group count |
| Create | `/admin/metadata/navigation/new` | Form: profile ID + JSON config textarea |
| Detail | `/admin/metadata/navigation/:id` | Read-only profile ID + editable JSON config + Save/Delete |

### 16.5 API

**Admin CRUD** (requires admin authentication):

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/admin/profile-navigation` | Create navigation config |
| `GET` | `/api/v1/admin/profile-navigation` | List all navigation configs |
| `GET` | `/api/v1/admin/profile-navigation/:id` | Get navigation config by ID |
| `PUT` | `/api/v1/admin/profile-navigation/:id` | Update navigation config |
| `DELETE` | `/api/v1/admin/profile-navigation/:id` | Delete navigation config |

**Resolution** (requires user authentication):

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/navigation` | Resolve navigation for current user's profile |
| `GET` | `/api/v1/view/:ovApiName` | Get Object View config by api_name (with FLS intersection) |

---

## 17. Automation Rules

### 17.1 Overview

Automation Rules (ADR-0031) allow administrators to define reactive triggers on DML events. When a record is inserted, updated, or deleted, matching rules evaluate CEL conditions and execute published procedures.

Key features:
- **6 event types**: `before_insert`, `after_insert`, `before_update`, `after_update`, `before_delete`, `after_delete`
- **CEL conditions**: expressions like `new.Status != old.Status` evaluated per record
- **Procedure execution**: each rule references a `procedure_code` — a published procedure
- **Execution modes**: `per_record` (one procedure call per record) or `per_batch` (one call for all records)
- **Execution order**: `sort_order` field controls rule evaluation order per object per event
- **Recursion depth limit**: configurable (default 3) to prevent infinite trigger chains

### 17.2 API

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/admin/automation-rules` | Create an automation rule |
| `GET` | `/api/v1/admin/automation-rules` | List automation rules (filter by `object_id`) |
| `GET` | `/api/v1/admin/automation-rules/:id` | Get automation rule by ID |
| `PUT` | `/api/v1/admin/automation-rules/:id` | Update an automation rule |
| `DELETE` | `/api/v1/admin/automation-rules/:id` | Delete an automation rule |

---

## 18. Layouts

### 18.1 Overview

Layouts (ADR-0027 revised) control **how** records are displayed per Object View, form factor, and mode. While an Object View defines **what** data is available (fields, actions, queries), a Layout defines **how** that data is presented on screen: section grids, field sizing, UI component types, and list column configuration.

Each Layout is scoped to a unique combination of:
- **object_view_id** — the Object View this layout belongs to
- **form_factor** — `desktop`, `tablet`, or `mobile`
- **mode** — `edit` or `view`

This means you can have different layouts for the same Object View on different devices and for different interaction modes (viewing a record vs. editing it).

### 18.2 Layout Config

The Layout config (stored as JSONB) can contain the following sections:

**section_config** — overrides for OV sections:
- `columns` (1-4) — number of grid columns in the section
- `collapsed` (boolean) — whether the section starts collapsed
- `visibility_expr` (CEL string) — condition for showing/hiding the section

**field_config** — per-field presentation overrides:
- `col_span` (1-4) — how many grid columns the field occupies
- `ui_kind` (string) — UI component type (e.g., `auto`, `text`, `textarea`, `badge`, `lookup`, `rating`, `slider`, `toggle`)
- `required_expr` (CEL string) — dynamic required condition
- `readonly_expr` (CEL string) — dynamic read-only condition
- `reference_config` — for reference fields: display_fields, search_fields, target, filter

**list_config** — for list/table views:
- `columns` — column definitions (api_name, width, align, sortable)
- `sort_by` — default sort configuration
- `search` — search field configuration
- `row_actions` — per-row action buttons

### 18.3 Form Merge & Fallback

When a client requests record metadata via the Describe API, the server performs a **form merge**:

1. Resolves the Object View for the current user's profile
2. Finds the matching Layout based on `X-Form-Factor` and `X-Form-Mode` request headers
3. Merges OV config + Layout config into a computed **Form**
4. The frontend works exclusively with the Form — it never sees OV or Layout separately

**Fallback chain** (if no exact match is found):
1. Requested form_factor + requested mode
2. Same form_factor + any mode
3. Desktop + same mode
4. Desktop + edit
5. Auto-generate from metadata (all FLS-accessible fields)

**Request headers:**
- `X-Form-Factor`: `desktop` (default), `tablet`, or `mobile`
- `X-Form-Mode`: `edit` (default) or `view`

### 18.4 Admin UI — Visual Layout Constructor

The Layout admin UI is available at `/admin/metadata/layouts`. The detail page features a **Visual Layout Constructor** with three tabs:

**Form Layout tab** (default):
- **Canvas** (left panel, ~65%): displays OV sections as cards with field chips inside. Each section card shows the column count and collapsible icon. Fields show their name, type icon, and col_span badge. Clicking a section or field selects it for editing.
- **Properties panel** (right panel, ~35%): context-sensitive editor that shows section properties (columns, collapsible, collapsed by default, visibility expression) when a section is selected, or field properties (col_span, shared layout ref, required/readonly/visibility expressions, reference config) when a field is selected.

**List Config tab**:
- **Available fields** (left): all OV fields as clickable items. Click to add a field as a column.
- **Active columns** (right): drag-and-drop reorderable list (via vue-draggable-plus). Each column has inline settings: width, alignment, label override, sortable toggle, sort direction.
- **Search config**: configurable search fields and placeholder text.

**JSON tab** (power-user fallback):
- Raw JSON textarea for direct editing of the full LayoutConfig. Bidirectional sync with visual tabs — changes in visual tabs update JSON and vice versa. Shows parse errors for invalid JSON.

Other pages:
- **List page**: all layouts with Object View name, form factor badge, mode badge, OV filter dropdown
- **Create page**: select Object View, form factor, mode

### 18.5 API

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/admin/layouts` | Create a layout |
| `GET` | `/api/v1/admin/layouts` | List layouts (filter by `object_view_id`) |
| `GET` | `/api/v1/admin/layouts/:id` | Get layout by ID |
| `PUT` | `/api/v1/admin/layouts/:id` | Update a layout |
| `DELETE` | `/api/v1/admin/layouts/:id` | Delete a layout |

---

## 19. Shared Layouts

### 19.1 Overview

Shared Layouts are reusable configuration snippets stored in `metadata.shared_layouts`. Instead of duplicating the same field config, section config, or list config across multiple layouts, you can define it once as a Shared Layout and reference it from any layout via `layout_ref`.

### 19.2 Types

Each Shared Layout has a `type` that determines what kind of configuration it holds:

| Type | Description | Use Case |
|------|-------------|----------|
| `field` | Field presentation config | Reusable field_config (col_span, ui_kind, etc.) across layouts |
| `section` | Section presentation config | Shared section_config (columns, collapsed, visibility) |
| `list` | List/table presentation config | Common list_config (columns, sort, search, row_actions) |

Each Shared Layout has a globally unique `api_name` for easy identification and referencing.

### 19.3 layout_ref & Overrides

To reference a Shared Layout from within a Layout config, use the `layout_ref` field:

```json
{
  "field_config": {
    "some_field": {
      "layout_ref": "shared_address_field",
      "col_span": 2
    }
  }
}
```

**Merge rule**: inline properties (like `col_span: 2` above) **override** matching properties from the referenced Shared Layout. This means you can use a shared base and customize per-layout.

**Delete protection**: Shared Layouts referenced by any Layout cannot be deleted (RESTRICT). You must remove all references before deleting a Shared Layout.

### 19.4 API

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/admin/shared-layouts` | Create a shared layout |
| `GET` | `/api/v1/admin/shared-layouts` | List shared layouts (filter by `type`) |
| `GET` | `/api/v1/admin/shared-layouts/:id` | Get shared layout by ID |
| `PUT` | `/api/v1/admin/shared-layouts/:id` | Update a shared layout |
| `DELETE` | `/api/v1/admin/shared-layouts/:id` | Delete a shared layout (RESTRICT if referenced) |

---

## 20. Common Scenarios

### Scenario 1: First Login

1. Ensure that the `ADMIN_INITIAL_PASSWORD` environment variable is set when starting the server.
2. Navigate to the login page (`/login`).
3. Enter:
   - Username: `admin`
   - Password: value from `ADMIN_INITIAL_PASSWORD`
4. Click **"Login"**.
5. After logging in, it is recommended to change the admin password via the API:
   ```
   PUT /api/v1/admin/security/users/<admin-uuid>/password
   {"password": "new-secure-password"}
   ```

### Scenario 2: Recover a Forgotten Password

1. On the login page (`/login`), click **"Forgot password?"**.
2. Enter the email associated with the account.
3. Click **"Send"**.
4. Open the email and follow the password reset link.
5. Enter the new password and confirm it.
6. Click **"Reset password"**.
7. Log in with the new password.

### Scenario 3: Create a New Object with Fields

1. Go to **Objects** (`/admin/metadata/objects`).
2. Click **"Create Object"**.
3. Fill in:
   - API Name: `Invoice__c`
   - Label: `Invoice`
   - Plural Label: `Invoices`
   - Object Type: `custom`
   - Enable the necessary flags (at minimum: Record creation, Record updates, Queries).
4. Click **"Create"**.
5. On the object detail page, switch to the **"Fields"** tab.
6. Click **"Add Field"** and create fields:
   - `number__c` — type `text`, subtype `plain`, max length: 50, required, unique.
   - `amount__c` — type `number`, subtype `currency`, precision: 18, scale: 2.
   - `invoice_date__c` — type `datetime`, subtype `date`.
   - `account_id__c` — type `reference`, subtype `association`, reference object: `Account`, on delete: `set_null`.

### Scenario 4: Configure Object Access

1. Create or open a **permission set** (`/admin/security/permission-sets`).
   - When creating: API Name: `invoice_access`, Label: "Invoice Access", Type: `Grant`.
2. Go to the **"Object Permissions"** tab.
3. Find the `Invoice__c` row and check the needed checkboxes: Read, Create, Update.
4. Go to the **"Field Permissions"** tab.
5. Select the `Invoice__c` object from the dropdown.
6. For each field, check: Read and Write.

### Scenario 5: Create a User with Full Security Setup

1. **Create a role** (if none exists):
   - Go to **Roles** → **"Create Role"**.
   - API Name: `sales_manager`, Label: "Sales Manager".
   - Parent Role: select if needed.

2. **Create a profile** (if none exists):
   - Go to **Profiles** → **"Create Profile"**.
   - API Name: `sales_profile`, Label: "Sales Profile".

3. **Configure profile OLS/FLS:**
   - On the profile detail page, click the **"Open Base Permission Set"** link.
   - On the **"Object Permissions"** tab — configure CRUD permissions.
   - On the **"Field Permissions"** tab — configure field access.

4. **Create a user:**
   - Go to **Users** → **"Create User"**.
   - Fill in: username, email, first name, last name.
   - Select profile: "Sales Profile".
   - Select role: "Sales Manager".

5. **Assign additional permission sets** (if needed):
   - On the user detail page, go to the **"Permission Sets"** tab.
   - Click **"Assign Permission Set"** and select the desired one.

### Scenario 6: Restrict User Access with Deny

1. Create a **deny permission set:**
   - Go to **Permission Sets** → **"Create Permission Set"**.
   - API Name: `no_delete_invoices`, Label: "No Invoice Deletion", Type: `Deny`.
2. On the **"Object Permissions"** tab — check the **Delete** checkbox for the `Invoice__c` object.
3. Assign this set to the user:
   - Open the user detail page → **"Permission Sets"** tab → **"Assign Permission Set"** → select "No Invoice Deletion".

Now, even if the user's profile allows invoice deletion, the deny set will block it.

### Scenario 7: Change a User's Role

1. Go to **Users** → open the desired user.
2. On the **"General"** tab, change the value in the **Role** field.
3. Click **"Save"**.

Membership in role groups is recalculated automatically.

### Scenario 8: Configure Object Visibility

1. Go to **Objects** → open the desired object.
2. In the **"Visibility (OWD)"** field, select the desired value:
   - `private` — only the owner can see records (default).
   - `public_read` — everyone can read, updates — only by owner.
   - `public_read_write` — full access for everyone (no share table created).
   - `controlled_by_parent` — access is determined by the parent object.
3. Click **"Save"**.

> **Warning:** When switching from `public_read_write` to any other mode, a share table is created. When switching to `public_read_write` — the share table is deleted along with all sharing entries.

### Scenario 9: Share Records with the Sales Department

1. **Create a sharing rule:**
   - Object: select the desired one (e.g., `Invoice__c`).
   - Rule type: `owner_based`.
   - Source group: `role_sales_manager` (record owners — managers).
   - Target group: `role_and_sub_sales_director` (access — director and all subordinates).
   - Access level: `read`.

Now all records owned by sales managers will be visible to the sales director and their subordinates.

### Scenario 10: Grant Manual Access to a Specific Record

1. Call the manual sharing API:
   - Table: `obj_invoice` (object table name).
   - Record ID: UUID of the specific invoice.
   - Group: UUID of the user's personal group (or a public group).
   - Access level: `read_write`.

The entry will appear in the `obj_invoice__share` share table with reason `manual`.

### Scenario 11: Create a Public Group for a Project Team

1. Create a group of type `public`:
   - API Name: `project_alpha_team`
   - Label: "Project Alpha Team"
2. Add members — individual users or entire groups.
3. Use this group in sharing rules or for manual sharing.

---

### Scenario 12: Execute a SOQL Query

1. Send a GET request:
   ```
   GET /api/v1/query?q=SELECT Id, Name, Email FROM Contact WHERE Status = 'Active' ORDER BY Name LIMIT 20
   ```
2. Or a POST request for long queries:
   ```json
   POST /api/v1/query
   {
     "query": "SELECT Name, SUM(Amount) FROM Deal GROUP BY Name HAVING SUM(Amount) > 100000",
     "pageSize": 50
   }
   ```
3. The response contains a `records` array with the fields requested in SELECT. All fields inaccessible to the user by FLS will be excluded. Records invisible by RLS will not appear in the results.

### Scenario 13: Create Records via DML

1. Send a POST request:
   ```json
   POST /api/v1/data
   {
     "statement": "INSERT INTO Contact (FirstName, LastName, Email) VALUES ('John', 'Smith', 'john@example.com')"
   }
   ```
2. The response contains `inserted_ids` — a list of UUIDs of the created records.
3. For batch inserts, pass multiple VALUES rows separated by commas (up to 10,000 rows).

### Scenario 14: Update Records via DML

1. Send a POST request:
   ```json
   POST /api/v1/data
   {
     "statement": "UPDATE Contact SET Status = 'Inactive' WHERE LastLoginDate < 2025-01-01"
   }
   ```
2. The response contains `updated_ids`. RLS guarantees that only records visible to the current user will be updated.

### Scenario 15: Set Up a Territory Model (Enterprise)

1. **Create a model:** POST `/api/v1/admin/territory/models` with `api_name`, `label`.
2. **Create territories:** POST `/api/v1/admin/territory/territories` — root and child territories (specifying `parent_id`).
3. **Configure access:** POST `/api/v1/admin/territory/territories/:id/object-defaults` — specify `object_id` and `access_level` (`read` or `read_write`) for each object.
4. **Assign users:** POST `/api/v1/admin/territory/territories/:id/users` — add users to territories.
5. **Activate the model:** POST `/api/v1/admin/territory/models/:id/activate` — territories will begin affecting record visibility.

### Scenario 16: Apply an App Template

1. Ensure the database is empty (no objects created yet).
2. Navigate to `/admin/templates`.
3. Choose a template (e.g., "Sales CRM") and click "Apply".
4. The system creates 4 objects (Account, Contact, Opportunity, Task) with 36 fields, grants OLS permissions to the System Administrator profile, and rebuilds the metadata cache.
5. Navigate to `/app` — the sidebar now shows the created objects. You can immediately start creating records.

### Scenario 17: Create and Edit Records via Generic CRUD

1. Navigate to `/app` — the sidebar shows all objects accessible to the current user.
2. Click on an object (e.g., "Accounts") to see the record list.
3. Click "Create" — the form renders all editable fields from the object metadata.
4. Fill in the required fields (e.g., Name) and click "Create".
5. The system injects system fields (OwnerId, CreatedById), applies dynamic defaults, runs validation rules, and saves the record.
6. To edit, click on a record in the list, modify fields, and click "Save".
7. To delete, open a record and click "Delete" with confirmation.

### Scenario 18: Configure a Validation Rule

1. Navigate to `/admin/metadata/objects` and click on the target object.
2. Go to the "Validation Rules" tab and click "Create".
3. Fill in the form:
   - **api_name:** `amount_positive`
   - **label:** "Amount must be positive"
   - **expression:** `record.Amount > 0` (use the Expression Builder for autocompletion)
   - **error_message:** "Amount must be greater than zero"
   - **severity:** `error`
   - **applies_to:** `create,update`
4. Click "Create". The rule is immediately active.
5. Test: Try to create a record with Amount = 0 — the system returns an error with the message "Amount must be greater than zero".

### Scenario 19: Set Up Dynamic Defaults for a Field

1. Navigate to `/admin/metadata/objects/{objectId}/fields` and edit the target field.
2. Set the `default_expr` property to a CEL expression. Examples:
   - `user.id` — auto-fill with the current user's ID.
   - `now` — set the current timestamp.
   - `"ACC-" + string(now.year)` — generate a code prefix.
3. Set `default_on` to control when the default applies: `create`, `update`, or `create,update`.
4. Save the field definition.
5. Test: Create a record without providing a value for this field — the default expression is evaluated and the result is saved.

### Scenario 20: Create a Custom Function

1. Navigate to `/admin/functions` and click "Create".
2. Fill in:
   - **name:** `discount` (lowercase with underscores)
   - **description:** "Calculates discount by amount"
   - **parameters:** Add a parameter `amount` of type `number`
   - **return_type:** `number`
   - **body:** `amount > 1000 ? amount * 0.1 : 0.0`
3. The Expression Builder validates the expression in real time.
4. Click "Create".
5. Now use this function in any CEL context:
   - Validation rule: `fn.discount(record.Amount) < record.MaxDiscount`
   - Dynamic default: `fn.discount(record.TotalSales)`

### Scenario 21: Use the Expression Builder

1. When editing a validation rule, default expression, or function body, the Expression Builder provides:
   - **CodeMirror editor** with syntax highlighting for CEL.
   - **Field picker** (tab "Fields") — click a field name to insert `record.FieldName` at the cursor.
   - **Function picker** (tab "Functions") — click a function name to insert `fn.function_name()` at the cursor.
   - **Real-time validation** — the expression is checked as you type; errors are shown with line and column numbers.
   - **Return type display** — shows the inferred return type of the expression.
2. The same Expression Builder component is used across all CEL contexts: validation rules, when-expressions, default expressions, function bodies, and action visibility expressions.

### Scenario 22: Create an Object View for Sales

1. Navigate to **Object Views** (`/admin/metadata/object-views`).
2. Click **"+"** to create a new view.
3. Fill in:
   - **API Name:** `account_sales`
   - **Label:** "Account Sales View"
   - **Object:** Account
   - **Profile:** Sales Manager (or leave empty for a global view)
   - **Is Default:** leave unchecked (unless this should be the fallback)
4. Click **"Create"** — you are redirected to the visual constructor.
5. Switch to the **Fields** tab:
   - Add fields: Name, Industry, Phone, Revenue, AnnualBudget, Employees.
   - Order matters — first 3 (Name, Industry, Phone) become highlights in the computed form.
6. Switch to the **Actions** tab:
   - Add an action: key `send_proposal`, label "Send Proposal", type `primary`, icon `mail`.
   - Set visibility expression: `record.Status == 'draft'` (uses the Expression Builder).
7. Click **"Save"**.
9. Log in as a user with the Sales Manager profile. Navigate to Accounts — the record detail page now shows fields, highlights, and the "Send Proposal" button (when Status is 'draft').

### Scenario 23: Set Up a Default Object View for All Users

1. Create an Object View with **Is Default** checked and **Profile** left empty.
2. Configure fields in the visual constructor.
3. All users who don't have a profile-specific view will see this layout.
4. To override for a specific profile — create another Object View for the same object with that profile selected.

### Scenario 24: Create and Execute a Procedure

1. Navigate to **Procedures** (`/admin/metadata/procedures`).
2. Click **"+"** to create a new procedure:
   - **Code:** `create_account_with_contact`
   - **Name:** "Create Account with Contact"
3. On the detail page, the **Definition** tab shows the Constructor UI.
4. Click **"+"** → **Record** → `record.create`:
   - Set **Object:** `Account`
   - Set **As:** `account`
   - Map **Data:** `Name` → `$.input.accountName`, `Industry` → `$.input.industry`
5. Click **"+"** → **Record** → `record.create`:
   - Set **Object:** `Contact`
   - Set **As:** `contact`
   - Map **Data:** `FirstName` → `$.input.contactName`, `AccountId` → `$.account.id`
6. Click **"Save Draft"**.
7. Switch to the **Dry Run** tab:
   - Enter input: `{"accountName": "Acme", "industry": "Tech", "contactName": "John"}`
   - Click **"Run"** — verify the trace shows both commands as successful.
8. Click **"Publish"** to make the procedure live.
9. Execute via API:
   ```
   POST /api/v1/admin/procedures/:id/execute
   {"input": {"accountName": "Acme", "industry": "Tech", "contactName": "John"}}
   ```

### Scenario 25: Configure a Named Credential for an External API

1. Navigate to **Credentials** (`/admin/metadata/credentials`).
2. Click **"+"** to create:
   - **Code:** `stripe_api`
   - **Name:** "Stripe API"
   - **Base URL:** `https://api.stripe.com`
   - **Type:** `api_key`
   - **Header:** `Authorization`
   - **Value:** `Bearer sk_live_abc123`
3. Click **"Create"**.
4. Click the **Test Connection** button (Wifi icon) — verify it returns a successful response.
5. Now use this credential in a procedure's `integration.http` command:
   - **Credential:** `stripe_api`
   - **Method:** `POST`
   - **Path:** `/v1/charges`
   - **Body:** `{"amount": "$.input.amount", "currency": "usd"}`

### Scenario 26: Use Saga Rollback for Distributed Operations

1. Create a procedure `order_with_payment`.
2. Add a `record.create` command for Order, set **As:** `order`.
3. Expand rollback for this command — add `record.delete` with **Object:** `Order`, **ID:** `$.order.id`.
4. Add an `integration.http` command for payment (e.g., Stripe charge), set **As:** `payment`.
5. Expand rollback — add another `integration.http` for refund using `$.payment.id`.
6. Add a third command (e.g., email notification) without rollback.
7. **Test:** if the email command fails:
   - Payment rollback runs first (refund).
   - Then order rollback runs (delete record).
   - LIFO order ensures consistent compensation.
8. Use **Dry Run** to verify the trace — rollback commands appear when a step fails.

### Scenario 27: Rollback a Procedure to a Previous Version

1. Open a published procedure's detail page.
2. Switch to the **Versions** tab to view version history.
3. Make changes to the definition and click **"Save Draft"** → **"Publish"** (this creates v2).
4. If v2 has issues, click **"Rollback"**:
   - The current published version (v2) becomes superseded.
   - The previous version (v1) is restored as published.
5. Verify by switching to the **Versions** tab — v1 should show as "published" again.

---

### Scenario 28: Configure Profile Navigation

1. Navigate to **Navigation** (`/admin/metadata/navigation`).
2. Click **"+"** to create a new navigation config.
3. Enter the **Profile ID** of the target profile (UUID).
4. Enter the config JSON:
   ```json
   {
     "groups": [
       {
         "key": "sales",
         "label": "Sales",
         "items": [
           { "type": "object", "object_api_name": "Account" },
           { "type": "object", "object_api_name": "Opportunity" },
           { "type": "page", "label": "Sales Dashboard", "ov_api_name": "sales_dashboard", "icon": "layout-dashboard" }
         ]
       },
       {
         "key": "support",
         "label": "Support",
         "items": [
           { "type": "link", "label": "Help Center", "url": "https://help.example.com", "icon": "life-buoy" }
         ]
       }
     ]
   }
   ```
5. Click **"Create"**.
6. Log in as a user with the target profile — the sidebar now shows grouped navigation instead of the flat object list.

---

*Document created for CRM Platform. Current for Phase 0–10b (Scaffolding, Metadata engine, Security engine, SOQL, DML, Auth, App Templates, Generic CRUD, CEL engine, Validation Rules, Dynamic Defaults, Custom Functions, Object Views, Profile Navigation, Procedure Engine with Saga Rollback, Named Credentials, Automation Rules) + Territory Management (Enterprise).*
