BEGIN;
SELECT plan(14);

-- Table exists
SELECT has_table('metadata', 'functions', 'has metadata.functions table');

-- Columns
SELECT has_column('metadata', 'functions', 'id', 'has id column');
SELECT has_column('metadata', 'functions', 'name', 'has name column');
SELECT has_column('metadata', 'functions', 'description', 'has description column');
SELECT has_column('metadata', 'functions', 'params', 'has params column');
SELECT has_column('metadata', 'functions', 'return_type', 'has return_type column');
SELECT has_column('metadata', 'functions', 'body', 'has body column');
SELECT has_column('metadata', 'functions', 'created_at', 'has created_at column');
SELECT has_column('metadata', 'functions', 'updated_at', 'has updated_at column');

-- Types
SELECT col_type_is('metadata', 'functions', 'id', 'uuid', 'id is uuid');
SELECT col_type_is('metadata', 'functions', 'params', 'jsonb', 'params is jsonb');
SELECT col_type_is('metadata', 'functions', 'body', 'text', 'body is text');

-- Constraints
SELECT has_check('metadata', 'functions', 'has check constraints');

-- Index
SELECT has_index('metadata', 'functions', 'idx_functions_name', 'has name index');

SELECT * FROM finish();
ROLLBACK;
