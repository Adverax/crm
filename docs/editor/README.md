# CEL Expression Editor

A full-featured visual editor for [Common Expression Language (CEL)](https://github.com/google/cel-spec) expressions, built with Vue 3 + CodeMirror 6.

## Features

- **Syntax highlighting** — keywords, operators, strings, numbers, built-in functions, context variables (`record`, `old`, `user`, `now`, `args`, `datasets`, `data`), custom function namespace (`fn.*`)
- **Smart autocomplete** — context-aware completions: record fields, old values, user properties, custom functions (`fn.*`), built-in functions, keywords, operators
- **Backend validation** — server-side expression compilation via CEL-Go with line:column error reporting
- **Client-side preview** — live evaluation using `@marcbachmann/cel-js` with debounced (300ms) result display and type inference
- **Field & Function pickers** — visual panels for inserting field references and function calls
- **Custom functions** — `fn.*` namespace, dual-stack evaluation (cel-go backend + cel-js frontend), precompiled function registry
- **Multiple contexts** — 7 evaluation contexts with different available variables:
  - `validation_rule` — `record`, `old`, `user`, `now`
  - `when_expression` — `record`, `old`, `user`, `now`
  - `default_expr` — `record`, `user`, `now`
  - `function_body` — only function parameters
  - `visibility_expr` — field visibility
  - `portal_when` — `args`
  - `gate_when` — `args`, `datasets`, `data`, `user`
- **Error navigation** — clickable errors jump to the exact line:column in the editor
- **Mode switching** — toggle between CodeMirror editor and plain textarea
- **Configurable** — height, placeholder, disabled state, field picker visibility

## Architecture

```
┌──────────────────────────────────────────────────┐
│              ExpressionBuilder.vue                │
│  (orchestrator: toolbar, mode, validation, etc.)  │
├──────────────────────────────────────────────────┤
│                                                    │
│  ┌─────────────────┐  ┌────────────────────────┐ │
│  │ CodeMirrorEditor │  │  Toolbar               │ │
│  │ (CM6 wrapper)    │  │  [Mode][Validate]      │ │
│  │                  │  │  [Preview][Fields&Fns]  │ │
│  └────────┬─────────┘  └───────────┬────────────┘ │
│           │                        │               │
│  ┌────────┴─────────┐  ┌──────────┴────────────┐ │
│  │ cel-language.ts   │  │ Popover (Picker)       │ │
│  │ (syntax parser)   │  │ ┌──────────────────┐  │ │
│  │                   │  │ │  FieldPicker.vue  │  │ │
│  │ cel-autocomplete  │  │ │  FunctionPicker   │  │ │
│  │ (smart complete)  │  │ └──────────────────┘  │ │
│  └───────────────────┘  └───────────────────────┘ │
│                                                    │
│  ┌─────────────────────────────────────────────┐  │
│  │ ExpressionPreview.vue (cel-js live eval)     │  │
│  └─────────────────────────────────────────────┘  │
│  ┌─────────────────────────────────────────────┐  │
│  │ ExpressionErrors.vue (clickable errors)      │  │
│  └─────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────┘
         │                        │
         ▼                        ▼
  ┌──────────────┐     ┌────────────────────┐
  │ cel-js       │     │ Backend (Go)       │
  │ (client-side │     │ POST /cel/validate │
  │  evaluation) │     │ (cel-go compile)   │
  └──────────────┘     └────────────────────┘
```

## File Structure

```
docs/editor/
├── README.md                  ← this file
├── cel-language.ts            ← CodeMirror 6 CEL syntax highlighter (StreamLanguage)
├── cel-autocomplete.ts        ← CodeMirror 6 smart autocomplete extension
├── cel-environment.ts         ← Client-side CEL evaluation (cel-js wrapper)
├── types.ts                   ← TypeScript type definitions
├── cel-handler.go             ← Go backend: HTTP validation endpoint
├── cel-env.go                 ← Go backend: CEL environment builders
├── cel-functions.go           ← Go backend: custom function registry
├── cel-eval.go                ← Go backend: ProgramCache + evaluation helpers
└── cel-errors.go              ← Go backend: typed error types
```

## Frontend Components

### ExpressionBuilder.vue (Main Component)

The orchestrator component that integrates all parts.

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `modelValue` | `string` | required | CEL expression (v-model) |
| `context` | `CelContext` | required | Evaluation context |
| `objectApiName` | `string?` | — | Object for field loading |
| `functionParams` | `FunctionParam[]` | `[]` | Params for function_body context |
| `height` | `string` | `'120px'` | Editor height |
| `placeholder` | `string` | `''` | Placeholder text |
| `disabled` | `boolean` | `false` | Disable editing |
| `showFieldPicker` | `boolean` | `true` | Show fields & functions button |

**Events:**
- `update:modelValue` — expression changed

**Usage:**
```vue
<ExpressionBuilder
  v-model="expression"
  context="validation_rule"
  object-api-name="Contact"
/>
```

### CodeMirrorEditor.vue (Editor Wrapper)

Low-level CodeMirror 6 wrapper with dynamic extension reconfiguration.

**Exposed methods:**
- `insertAtCursor(text: string)` — insert text at cursor position
- `setCursorAt(offset: number)` — move cursor to absolute offset
- `setCursorAtLineCol(line: number, col: number)` — move cursor to line:column
- `getCursorOffset(): number` — get current cursor offset

**Features:**
- Bracket matching, history (undo/redo), line wrapping
- Dynamic `Compartment` for live extension updates (autocomplete changes when fields load)
- Custom theme matching the host app's CSS variables

### FieldPicker.vue

Context-aware variable picker with search.

**Groups by context:**
- `validation_rule` / `when_expression` — Record Fields, Old Values, User, System
- `default_expr` — Record Fields, User, System
- `function_body` — Parameters only
- `portal_when` — Portal Args
- `gate_when` — Gate Variables (args, datasets, data, user)

### FunctionPicker.vue

Function browser with two sections:
1. **Custom (fn.*)** — functions from the registry
2. **Built-in** — String (size, contains, startsWith, endsWith, matches), Type Casting (int, double, string, bool), Time (duration, timestamp), General (has, type)

### ExpressionPreview.vue

Live client-side evaluation with debounced updates (300ms).
- For `function_body`: shows parameter input fields for manual testing
- For other contexts: builds a sample record from field definitions
- Uses `Proxy` for dynamic property access (unknown fields return `0`)

### ExpressionErrors.vue

Displays validation errors from the backend. Clickable errors with line:column jump to the error position in the editor.

## Frontend Libraries

### cel-language.ts — Syntax Highlighting

StreamLanguage parser for CodeMirror 6 that highlights:
- Strings (single/double quoted with escape sequences)
- Line comments (`//`)
- Numbers (integers, floats, scientific notation)
- Operators (`&&`, `||`, `==`, `!=`, `>=`, `<=`, `+`, `-`, `*`, `/`, `%`, `<`, `>`, `!`)
- Keywords (`true`, `false`, `null`, `in`, `has`)
- Namespace (`fn`)
- Context variables (`record`, `old`, `user`, `now`)
- Built-in functions (`size`, `contains`, `startsWith`, `endsWith`, `matches`, `int`, `uint`, `double`, `string`, `bool`, `type`, `duration`, `timestamp`)

### cel-autocomplete.ts — Smart Autocomplete

Context-aware autocompletion:
- After `record.` → show record fields with types
- After `old.` → show fields (validation_rule/when_expression only)
- After `args.` → dynamic hint (portal_when/gate_when)
- After `datasets.` → dynamic hint (gate_when)
- After `data.` → dynamic hint (gate_when)
- After `user.` → show user properties (id, profile_id, role_id)
- After `fn.` → show custom functions with signatures
- General context → variables, fields, functions, keywords, operators

### cel-environment.ts — Client-side Evaluation

Uses `@marcbachmann/cel-js` for browser-side CEL evaluation:
- `createCelEnvironment(functions)` — registers custom `fn.*` functions
- `evaluateCel(env, expression, context)` — evaluates with BigInt→Number conversion
- `evaluateCelSafe(env, expression, context, timeoutMs)` — with timeout protection
- Type inference: null, bool, number, int, string, list, map, unknown

## Backend (Go)

### CEL Handler (cel-handler.go)

HTTP endpoint: `POST /api/v1/admin/cel/validate`

**Request:**
```json
{
  "expression": "record.FirstName != ''",
  "context": "validation_rule",
  "object_api_name": "Contact",
  "params": [{"name": "x", "type": "number"}]
}
```

**Response (valid):**
```json
{
  "valid": true,
  "return_type": "bool"
}
```

**Response (invalid):**
```json
{
  "valid": false,
  "errors": [
    {"message": "undeclared reference 'foo'", "line": 1, "column": 5}
  ]
}
```

### CEL Environments (cel-env.go)

Environment builders per context:

| Context | Variables |
|---------|-----------|
| `StandardEnv` | record (dyn), old (dyn), user (dyn), now (timestamp) + Strings |
| `DefaultEnv` | record (dyn), user (dyn), now (timestamp) + Strings |
| `PortalEnv` | args (dyn) + Strings |
| `GateEnv` | args (dyn), datasets (dyn), data (dyn), user (dyn) + Strings |
| `FunctionBodyEnv` | params as variables + Strings |

All environments include `ext.Strings()` for string extension functions.

### Function Registry (cel-functions.go)

Thread-safe registry for custom `fn.*` functions:
- Precompiles function bodies at initialization
- `EnvOptions()` returns `cel.EnvOption` entries for registration
- Each function registered as `fn.name(dyn, ...): dyn` overload
- `RWMutex` for concurrent read access

### ProgramCache (cel-eval.go)

Thread-safe compiled program cache:
- `GetOrCompile(expr)` — compile once, reuse forever
- `EvaluateBool(expr, vars)` — evaluate expecting boolean
- `EvaluateAny(expr, vars)` — evaluate returning any type
- `Reset(env)` — rebuild when functions change

## Dependencies

### Frontend (npm)
```json
{
  "@codemirror/autocomplete": "^6.x",
  "@codemirror/commands": "^6.x",
  "@codemirror/language": "^6.x",
  "@codemirror/state": "^6.x",
  "@codemirror/view": "^6.x",
  "@lezer/highlight": "^1.x",
  "@marcbachmann/cel-js": "^0.x",
  "vue": "^3.x"
}
```

### Backend (Go)
```
github.com/google/cel-go v0.22+
github.com/google/cel-go/ext (Strings extension)
```

## Integration Guide

### Minimal Frontend Setup

```vue
<script setup lang="ts">
import { ref } from 'vue'
import ExpressionBuilder from './ExpressionBuilder.vue'

const expression = ref('')
</script>

<template>
  <ExpressionBuilder
    v-model="expression"
    context="validation_rule"
    object-api-name="Contact"
  />
</template>
```

### Standalone Usage (without CRM)

To use the editor in another project, you need:

1. **Copy the standalone files** from this directory (`cel-language.ts`, `cel-autocomplete.ts`, `cel-environment.ts`, `types.ts`)
2. **Install npm dependencies** (CodeMirror 6 + cel-js)
3. **Adapt the Vue components** or build your own using the library files
4. **Implement a validation API** endpoint (or use client-side only evaluation)

The library files (`cel-language.ts`, `cel-autocomplete.ts`, `cel-environment.ts`) have no Vue dependencies and can be used with any framework.

### Backend Setup

```go
// Create function registry
registry, err := cel.NewFunctionRegistry([]cel.FunctionDef{
    {Name: "fullName", Params: []cel.ParamDef{{Name: "first"}, {Name: "last"}}, Body: `first + " " + last`},
})

// Create program cache with standard env + functions
env, _ := cel.StandardEnvWithFunctions(registry)
cache := cel.NewProgramCache(env)

// Evaluate
result, err := cache.EvaluateBool(`record.FirstName != ""`, map[string]any{
    "record": map[string]any{"FirstName": "John"},
    "user":   map[string]any{"id": "123"},
    "now":    time.Now(),
})
```

## CEL Quick Reference

```
// Comparison
record.Amount > 1000
record.Status == "active"
record.Email != ""

// Logical
record.Amount > 0 && record.Status == "active"
record.Type == "A" || record.Type == "B"
!(record.Amount < 0)

// String functions
record.Email.contains("@")
record.Name.startsWith("Mr")
record.Name.endsWith("Jr")
size(record.Name) > 0
record.Phone.matches("^\\+\\d{10,15}$")

// Ternary
record.Amount > 0 ? "positive" : "non-positive"

// Membership
record.Status in ["active", "pending"]

// Type casting
int(record.StringAmount)
string(record.NumericCode)
double(record.Price)

// Custom functions
fn.fullName(record.FirstName, record.LastName)
fn.isValidEmail(record.Email)

// Context variables
user.id             // current user ID
user.profile_id     // user's profile
user.role_id        // user's role
now                 // current timestamp
```
