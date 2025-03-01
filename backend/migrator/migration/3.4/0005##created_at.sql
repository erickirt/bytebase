ALTER TABLE task_run_log RENAME COLUMN created_ts TO created_at;
ALTER TABLE revision RENAME COLUMN created_ts TO created_at;
ALTER TABLE revision RENAME COLUMN deleted_ts TO deleted_at;
ALTER INDEX idx_revision_unique_database_id_version_deleted_ts_null RENAME TO idx_revision_unique_database_id_version_deleted_at_null;
ALTER TABLE sync_history RENAME COLUMN created_ts TO created_at;
ALTER INDEX idx_sync_history_database_id_created_ts RENAME TO idx_sync_history_database_id_created_at;
ALTER TABLE changelog RENAME COLUMN created_ts TO created_at;
ALTER TABLE release RENAME COLUMN created_ts TO created_at;
ALTER TABLE policy RENAME COLUMN updated_ts TO updated_at;
