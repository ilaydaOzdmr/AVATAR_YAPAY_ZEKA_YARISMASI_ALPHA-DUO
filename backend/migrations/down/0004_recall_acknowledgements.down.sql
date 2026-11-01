-- Revert 0004_recall_acknowledgements.up.sql.
-- Kept in a subdirectory so the MySQL init entrypoint never runs it.
DROP TABLE IF EXISTS recall_acknowledgements;
