DROP INDEX IF EXISTS ix_refresh_tokens_replaced_by;

ALTER TABLE refresh_tokens
    DROP COLUMN IF EXISTS replaced_by,
    DROP COLUMN IF EXISTS revoked_at;
