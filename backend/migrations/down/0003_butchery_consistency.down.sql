-- Revert 0003_butchery_consistency.up.sql.
-- Kept in a subdirectory so the MySQL init entrypoint (which runs only files
-- directly in /docker-entrypoint-initdb.d) never executes it on a fresh volume.
ALTER TABLE butchery_batches
  DROP COLUMN weight_check_status,
  DROP COLUMN cost_check_status,
  DROP COLUMN cost_difference,
  DROP COLUMN expected_cost;
