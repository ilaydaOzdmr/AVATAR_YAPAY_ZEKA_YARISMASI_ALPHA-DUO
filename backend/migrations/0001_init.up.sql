-- 0001_init.up.sql — AVATAR ALPHA-DUO schema
-- MySQL 8, InnoDB, utf8mb4
SET NAMES utf8mb4;

CREATE TABLE materials (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  code VARCHAR(64) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL,
  unit VARCHAR(32) NOT NULL,
  is_raw BOOL NOT NULL DEFAULT TRUE,
  default_purchase_price DECIMAL(12,2) NOT NULL DEFAULT 0.00
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE products (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  code VARCHAR(64) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL,
  unit VARCHAR(32) NOT NULL,
  list_price DECIMAL(12,2) NOT NULL DEFAULT 0.00,
  category VARCHAR(64) NOT NULL DEFAULT 'general',
  is_finished BOOL NOT NULL DEFAULT TRUE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE suppliers (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  code VARCHAR(64) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL,
  contact VARCHAR(255) NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE markets (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  code VARCHAR(64) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL,
  contact VARCHAR(255) NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE lots (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  code VARCHAR(64) NOT NULL UNIQUE,
  item_type ENUM('material','product') NOT NULL,
  item_id BIGINT NOT NULL,
  qty DECIMAL(12,3) NOT NULL DEFAULT 0.000,
  unit VARCHAR(32) NOT NULL,
  supplier_id BIGINT NULL,
  purchase_price DECIMAL(12,2) NOT NULL DEFAULT 0.00,
  unit_cost DECIMAL(12,2) NOT NULL DEFAULT 0.00,
  received_at DATETIME NULL,
  expiry_date DATE NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_lots_supplier FOREIGN KEY (supplier_id) REFERENCES suppliers(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE production_orders (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  code VARCHAR(64) NOT NULL UNIQUE,
  status VARCHAR(32) NOT NULL DEFAULT 'draft',
  started_at DATETIME NULL,
  finished_at DATETIME NULL,
  notes TEXT NULL,
  service_name VARCHAR(255) NULL,
  service_unit_cost DECIMAL(12,2) NOT NULL DEFAULT 0.00,
  service_qty DECIMAL(12,3) NOT NULL DEFAULT 0.000,
  expected_qty DECIMAL(12,3) NOT NULL DEFAULT 0.000
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE production_order_inputs (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  production_order_id BIGINT NOT NULL,
  material_lot_id BIGINT NOT NULL,
  qty DECIMAL(12,3) NOT NULL,
  CONSTRAINT fk_poi_order FOREIGN KEY (production_order_id) REFERENCES production_orders(id),
  CONSTRAINT fk_poi_lot FOREIGN KEY (material_lot_id) REFERENCES lots(id),
  INDEX idx_poi_order (production_order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE production_order_outputs (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  production_order_id BIGINT NOT NULL,
  product_lot_id BIGINT NOT NULL,
  qty DECIMAL(12,3) NOT NULL,
  CONSTRAINT fk_poo_order FOREIGN KEY (production_order_id) REFERENCES production_orders(id),
  CONSTRAINT fk_poo_lot FOREIGN KEY (product_lot_id) REFERENCES lots(id),
  INDEX idx_poo_order (production_order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE butchery_batches (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  code VARCHAR(64) NOT NULL UNIQUE,
  animal_type VARCHAR(64) NOT NULL,
  input_lot_id BIGINT NULL,
  input_weight DECIMAL(12,3) NOT NULL,
  unit_cost DECIMAL(12,2) NOT NULL,
  total_cost DECIMAL(12,2) NOT NULL,
  batch_date DATE NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'draft',
  CONSTRAINT fk_bb_input_lot FOREIGN KEY (input_lot_id) REFERENCES lots(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE butchery_items (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  batch_id BIGINT NOT NULL,
  product_id BIGINT NOT NULL,
  product_lot_id BIGINT NULL,
  weight DECIMAL(12,3) NOT NULL,
  quantity INT NOT NULL DEFAULT 1,
  is_waste BOOL NOT NULL DEFAULT FALSE,
  allocated_cost DECIMAL(12,2) NOT NULL DEFAULT 0.00,
  CONSTRAINT fk_bi_batch FOREIGN KEY (batch_id) REFERENCES butchery_batches(id),
  CONSTRAINT fk_bi_product FOREIGN KEY (product_id) REFERENCES products(id),
  CONSTRAINT fk_bi_lot FOREIGN KEY (product_lot_id) REFERENCES lots(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE lot_links (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  parent_lot_id BIGINT NOT NULL,
  child_lot_id BIGINT NOT NULL,
  qty DECIMAL(12,3) NOT NULL,
  link_type VARCHAR(32) NOT NULL DEFAULT 'production',
  CONSTRAINT fk_ll_parent FOREIGN KEY (parent_lot_id) REFERENCES lots(id),
  CONSTRAINT fk_ll_child FOREIGN KEY (child_lot_id) REFERENCES lots(id),
  INDEX idx_ll_parent (parent_lot_id),
  INDEX idx_ll_child (child_lot_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE stock_movements (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  lot_id BIGINT NOT NULL,
  movement_type VARCHAR(32) NOT NULL,
  qty DECIMAL(12,3) NOT NULL,
  unit_cost DECIMAL(12,2) NOT NULL DEFAULT 0.00,
  total_cost DECIMAL(12,2) NOT NULL DEFAULT 0.00,
  ref_type VARCHAR(32) NULL,
  ref_id BIGINT NULL,
  occurred_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_sm_lot FOREIGN KEY (lot_id) REFERENCES lots(id),
  INDEX idx_sm_lot (lot_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE sales (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  code VARCHAR(64) NOT NULL UNIQUE,
  market_id BIGINT NOT NULL,
  sold_at DATETIME NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'completed',
  client_id VARCHAR(64) NULL,
  CONSTRAINT fk_sales_market FOREIGN KEY (market_id) REFERENCES markets(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE sale_lines (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  sale_id BIGINT NOT NULL,
  lot_id BIGINT NOT NULL,
  qty DECIMAL(12,3) NOT NULL,
  unit_price DECIMAL(12,2) NOT NULL,
  total_price DECIMAL(12,2) NOT NULL,
  stock_movement_id BIGINT NULL,
  CONSTRAINT fk_sl_sale FOREIGN KEY (sale_id) REFERENCES sales(id),
  CONSTRAINT fk_sl_lot FOREIGN KEY (lot_id) REFERENCES lots(id),
  CONSTRAINT fk_sl_movement FOREIGN KEY (stock_movement_id) REFERENCES stock_movements(id),
  INDEX idx_sl_lot (lot_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE recall_events (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  code VARCHAR(64) NOT NULL UNIQUE,
  reason TEXT NULL,
  source_lot_id BIGINT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_re_source FOREIGN KEY (source_lot_id) REFERENCES lots(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE recall_blocks (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  recall_event_id BIGINT NOT NULL,
  lot_id BIGINT NOT NULL,
  blocked_at DATETIME NOT NULL,
  CONSTRAINT fk_rb_event FOREIGN KEY (recall_event_id) REFERENCES recall_events(id),
  CONSTRAINT fk_rb_lot FOREIGN KEY (lot_id) REFERENCES lots(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE market_redistributions (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  market_id BIGINT NOT NULL,
  lot_id BIGINT NOT NULL,
  sub_buyer VARCHAR(255) NOT NULL,
  contact VARCHAR(255) NULL,
  qty DECIMAL(12,3) NOT NULL,
  CONSTRAINT fk_mr_market FOREIGN KEY (market_id) REFERENCES markets(id),
  CONSTRAINT fk_mr_lot FOREIGN KEY (lot_id) REFERENCES lots(id),
  INDEX idx_mr_market (market_id),
  INDEX idx_mr_lot (lot_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE sync_outbox (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  client_id VARCHAR(64) NOT NULL,
  entity VARCHAR(64) NOT NULL,
  entity_id BIGINT NOT NULL,
  operation VARCHAR(32) NOT NULL,
  payload JSON NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  synced_at DATETIME NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
