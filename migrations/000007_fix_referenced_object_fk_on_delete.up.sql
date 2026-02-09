-- Меняем FK referenced_object_id: RESTRICT → SET NULL
-- При удалении объекта, ссылающиеся поля других объектов получают referenced_object_id = NULL
ALTER TABLE metadata.field_definitions
    DROP CONSTRAINT field_definitions_referenced_object_id_fkey;

ALTER TABLE metadata.field_definitions
    ADD CONSTRAINT field_definitions_referenced_object_id_fkey
    FOREIGN KEY (referenced_object_id) REFERENCES metadata.object_definitions(id)
    ON DELETE SET NULL;
