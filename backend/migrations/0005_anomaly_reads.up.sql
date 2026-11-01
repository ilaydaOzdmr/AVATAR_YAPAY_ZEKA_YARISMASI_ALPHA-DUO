-- Persist which rule-based anomalies the operator has already read.
CREATE TABLE anomaly_reads (
  alert_key VARCHAR(191) PRIMARY KEY,
  read_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
