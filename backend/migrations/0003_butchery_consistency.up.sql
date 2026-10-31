-- Persist the automatic fire/cost consistency checks for butchery batches.
-- Values are derived by the backend; they are never supplied by clients.
ALTER TABLE butchery_batches
  ADD COLUMN expected_cost DECIMAL(12,2) NOT NULL DEFAULT 0.00,
  ADD COLUMN cost_difference DECIMAL(12,2) NOT NULL DEFAULT 0.00,
  ADD COLUMN cost_check_status VARCHAR(16) NOT NULL DEFAULT 'ok',
  ADD COLUMN weight_check_status VARCHAR(16) NOT NULL DEFAULT 'ok';

-- Backfill existing batches, including seed records such as KUZ-2026-DEMO.
UPDATE butchery_batches b
SET b.expected_cost = ROUND(b.input_weight * b.unit_cost, 2),
    b.cost_difference = ROUND(b.total_cost - ROUND(b.input_weight * b.unit_cost, 2), 2),
    b.cost_check_status = CASE
        WHEN ROUND(b.total_cost, 2) = ROUND(b.input_weight * b.unit_cost, 2) THEN 'ok'
        ELSE 'mismatch'
    END,
    b.weight_check_status = CASE
        WHEN ROUND(b.input_weight - COALESCE((
            SELECT SUM(i.weight) FROM butchery_items i
            WHERE i.batch_id = b.id AND i.is_waste = FALSE
        ), 0), 3) < 0 THEN 'mismatch'
        ELSE 'ok'
    END;
