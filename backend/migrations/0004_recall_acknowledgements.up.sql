-- Persist recall recipient confirmations per affected source lot. A confirmation
-- survives page reloads and applies to every recall event for the same lot.
CREATE TABLE recall_acknowledgements (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  source_lot_id BIGINT NOT NULL,
  market_code VARCHAR(64) NOT NULL,
  confirmed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_ra_source_lot FOREIGN KEY (source_lot_id) REFERENCES lots(id),
  UNIQUE KEY uq_ra_source_market (source_lot_id, market_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
