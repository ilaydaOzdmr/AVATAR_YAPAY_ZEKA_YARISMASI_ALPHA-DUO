"""Central paths and constants for the ML service."""

from __future__ import annotations

from pathlib import Path

ML_ROOT = Path(__file__).resolve().parents[1]
DATA_DIR = ML_ROOT / "data"
PROCESSED_DIR = DATA_DIR / "processed"
ARTIFACTS_DIR = ML_ROOT / "artifacts"
REPORTS_DIR = ML_ROOT / "reports"
PLOTS_DIR = REPORTS_DIR / "plots"

RANDOM_STATE = 42
TEST_SIZE = 0.20
CV_FOLDS = 5
MODEL_VERSION = "1.0.0"

PRODUCTION_FEATURES = [
    "raw_material_kg",
    "spice_kg",
    "casing_kg",
    "other_material_kg",
    "material_unit_price_meat",
    "material_unit_price_spice",
    "material_unit_price_casing",
    "material_unit_price_other",
    "service_cost",
    "labor_cost",
    "product_type",
]
PRODUCTION_TARGETS = ["finished_qty", "yield_pct"]
PRODUCTION_GROUP = "production_batch"

BUTCHERY_FEATURES = [
    "breed",
    "sex",
    "age_months",
    "live_weight_kg",
    "carcass_weight_kg",
    "conformation_score",
    "fat_score",
]
BUTCHERY_TARGETS = [
    "leg_weight",
    "loin_weight",
    "shoulder_weight",
    "rib_weight",
    "mince_weight",
    "diced_meat_weight",
    "bone_weight",
    "waste_weight",
    "total_saleable_weight",
    "cutting_yield_pct",
]
BUTCHERY_GROUP = "farm_id"

DATASET_FILES = {
    "yield": PROCESSED_DIR / "production_orders_synthetic.csv",
    "butchery_lamb": PROCESSED_DIR / "butchery_lamb_synthetic.csv",
    "butchery_cattle": PROCESSED_DIR / "butchery_cattle_synthetic.csv",
}

DATA_ORIGIN = "synthetic_calibrated"
MODEL_STATUS = "synthetic_only"
