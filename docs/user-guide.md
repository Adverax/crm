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
14. [Common Scenarios](#14-common-scenarios)

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

Every Object View is stored as a JSONB config in the `metadata.object_views` table and is linked to a specific object. Optionally, it can be linked to a specific profile (profile-specific view) or left global (accessible as a default fallback).

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
   - **API Name:** `account_sales_view` (lowercase with underscores)
   - **Label:** "Account Sales View"
   - **Object:** Select the target object (e.g., Account)
   - **Profile:** Optionally select a profile (e.g., Sales). Leave empty for a global view.
   - **Is Default:** Check if this should be the default view for the object.
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

1. **General** — edit label, description, is_default flag. Read-only: api_name, object, profile.
2. **Fields** — add/remove field api_names to include in this view. Order matters — first 3 become highlights in the computed form.
3. **Actions** — add action buttons with key, label, type, icon, and a CEL visibility expression (uses the Expression Builder from Phase 8).
4. **Queries** — define named SOQL queries scoped to this view (name, SOQL statement, optional `when` condition).
5. **Computed (Read)** — define computed fields from CEL expressions (name, type, expression, optional `when` condition). These are display-only and not persisted.

**Write tabs:**

6. **Validation** — define view-scoped validation rules (CEL expression, error message, optional code, severity: error/warning, optional `when` condition). These are additive with metadata-level validation rules.
7. **Defaults** — define view-scoped default expressions (field, CEL expression, trigger: create/update/create,update, optional `when` condition). These replace metadata-level defaults.
8. **Computed (Write)** — define fields whose values are computed from CEL expressions on save (field, CEL expression).
9. **Mutations** — define DML operations scoped to this view (DML statement, optional `foreach` iteration, optional `sync` mapping, optional `when` condition).

Click **"Save"** to persist changes. All changes take effect immediately.

### 13.5 Resolution Logic

When the CRM UI loads a record page, the Describe API resolves the Object View using a 3-step cascade:

1. **Profile-specific view** — look for an Object View where `object_id` matches AND `profile_id` matches the current user's profile.
2. **Default view** — if no profile-specific view exists, look for an Object View where `is_default = true` for this object.
3. **Fallback** — if no Object View exists at all, the system auto-generates a form: one "Details" section with all FLS-accessible fields, first 3 fields as highlights, first 5 fields as list columns.

This allows gradual adoption — the system works without any Object Views configured, and administrators can add views incrementally.

### 13.6 FLS Intersection

The resolved Object View config is intersected with the current user's Field-Level Security (FLS) permissions:

- **Fields** — fields the user cannot read are removed from the flat field list (and consequently from the auto-generated sections and highlights)
- **List fields** — inaccessible columns are removed
- **Related lists** — related objects the user cannot read (OLS) are excluded

This ensures that even if an administrator includes a field in the Object View, users without FLS access will never see it.

### 13.7 Describe API Extension

The Describe API response now includes an optional `form` property alongside the existing `fields` array:

```
GET /api/v1/describe/{objectName}
```

**Response with Object View:**
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
          "key": "client_info",
          "label": "Client Information",
          "columns": 2,
          "collapsed": false,
          "fields": ["Name", "Industry", "Phone"]
        }
      ],
      "highlight_fields": ["Name", "Industry"],
      "actions": [
        {
          "key": "send_proposal",
          "label": "Send Proposal",
          "type": "primary",
          "icon": "mail",
          "visibility_expr": "record.Status == 'draft'"
        }
      ],
      "list_fields": ["Name", "Industry", "Phone"],
      "list_default_sort": "created_at DESC"
    }
  }
}
```

The `fields` array remains for backward compatibility. The `form` property is always present — either resolved from an Object View or auto-generated as a fallback.

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
GET /api/v1/admin/object-views?object_id={objectId}
```

Returns all Object Views, optionally filtered by object.

**Response:**
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "object_id": "660e8400-e29b-41d4-a716-446655440000",
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
  "object_id": "660e8400-e29b-41d4-a716-446655440000",
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

Note: `api_name`, `object_id`, and `profile_id` cannot be changed after creation.

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
| 409 | Duplicate (object_id, profile_id) pair |

---

## 14. Common Scenarios

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

---

*Document created for CRM Platform. Current for Phase 0–9a (Scaffolding, Metadata engine, Security engine, SOQL, DML, Auth, App Templates, Generic CRUD, CEL engine, Validation Rules, Dynamic Defaults, Custom Functions, Object Views with Read/Write config) + Territory Management (Enterprise).*
