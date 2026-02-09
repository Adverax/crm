-- Откат: возвращаем RESTRICT (поведение по умолчанию)
ALTER TABLE metadata.field_definitions
    DROP CONSTRAINT field_definitions_referenced_object_id_fkey;

ALTER TABLE metadata.field_definitions
    ADD CONSTRAINT field_definitions_referenced_object_id_fkey
    FOREIGN KEY (referenced_object_id) REFERENCES metadata.object_definitions(id);
