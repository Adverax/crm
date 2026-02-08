DROP TRIGGER IF EXISTS trg_outbox_notify ON security.security_outbox;
DROP FUNCTION IF EXISTS security.notify_outbox();

DROP INDEX IF EXISTS security.idx_security_outbox_unprocessed;
DROP TABLE IF EXISTS security.security_outbox;

DROP INDEX IF EXISTS security.idx_effective_field_lists_user_id;
DROP TABLE IF EXISTS security.effective_field_lists;

DROP INDEX IF EXISTS security.idx_effective_fls_user_id;
DROP TABLE IF EXISTS security.effective_fls;

DROP INDEX IF EXISTS security.idx_effective_ols_user_id;
DROP TABLE IF EXISTS security.effective_ols;
