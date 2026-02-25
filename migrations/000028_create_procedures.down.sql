ALTER TABLE metadata.procedures DROP CONSTRAINT IF EXISTS fk_procedures_draft_version;
ALTER TABLE metadata.procedures DROP CONSTRAINT IF EXISTS fk_procedures_published_version;

DROP TABLE IF EXISTS metadata.procedure_versions;
DROP TABLE IF EXISTS metadata.procedures;
