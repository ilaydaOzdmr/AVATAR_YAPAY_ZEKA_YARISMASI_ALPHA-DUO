-- Revert 0005_anomaly_reads.up.sql.
-- Kept in a subdirectory so the MySQL init entrypoint never runs it.
DROP TABLE IF EXISTS anomaly_reads;
