BEGIN;
SELECT plan(3);

SELECT has_column('iam', 'users', 'password_hash', 'has password_hash');
SELECT col_type_is('iam', 'users', 'password_hash', 'character varying(255)', 'password_hash is varchar(255)');
SELECT col_not_null('iam', 'users', 'password_hash', 'password_hash is NOT NULL');

SELECT finish();
ROLLBACK;
