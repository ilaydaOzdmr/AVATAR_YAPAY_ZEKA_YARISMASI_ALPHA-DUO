"""Compute reproducible data-quality checks for the processed datasets."""

from __future__ import annotations

import json
import sys
from pathlib import Path

import pandas as pd

ML_ROOT = Path(__file__).resolve().parents[1]
PROCESSED = ML_ROOT / "data" / "processed"
REPORTS = ML_ROOT / "reports"

PART_COLUMNS = [
    "leg_weight",
    "loin_weight",
    "shoulder_weight",
    "rib_weight",
    "mince_weight",
    "diced_meat_weight",
    "bone_weight",
    "waste_weight",
]
SALEABLE_COLUMNS = [
    "leg_weight",
    "loin_weight",
    "shoulder_weight",
    "rib_weight",
    "mince_weight",
    "diced_meat_weight",
]


def describe(frame: pd.DataFrame) -> dict:
    numeric = frame.select_dtypes(include="number")
    return {
        "rows": int(len(frame)),
        "columns": int(len(frame.columns)),
        "null_cells": int(frame.isna().sum().sum()),
        "numeric": {
            column: {
                "min": float(numeric[column].min()),
                "max": float(numeric[column].max()),
                "mean": float(numeric[column].mean()),
            }
            for column in numeric.columns
        },
    }


def main() -> None:
    reports: dict = {"datasets": {}, "checks": {}}

    production_path = PROCESSED / "production_orders_synthetic.csv"
    if production_path.exists():
        production = pd.read_csv(production_path)
        reports["datasets"]["production_orders_synthetic"] = describe(production)
        ekp_error = (
            (production["actual_total_cost"] / production["finished_qty"])
            - production["actual_ekp"]
        ).abs()
        reports["checks"]["production"] = {
            "data_origin": sorted(production["data_origin"].unique().tolist()),
            "batches": int(production["production_batch"].nunique()),
            "yield_pct_min": float(production["yield_pct"].min()),
            "yield_pct_max": float(production["yield_pct"].max()),
            "finished_qty_min": float(production["finished_qty"].min()),
            "max_abs_ekp_identity_error": float(ekp_error.max()),
            "usda_price_index_min": float(production["usda_price_index"].min()),
            "usda_price_index_max": float(production["usda_price_index"].max()),
        }

    for animal_type in ("lamb", "cattle"):
        path = PROCESSED / f"butchery_{animal_type}_synthetic.csv"
        if not path.exists():
            continue
        frame = pd.read_csv(path)
        reports["datasets"][f"butchery_{animal_type}_synthetic"] = describe(frame)
        part_sum = frame[PART_COLUMNS].sum(axis=1)
        saleable = frame[SALEABLE_COLUMNS].sum(axis=1)
        reports["checks"][f"butchery_{animal_type}"] = {
            "data_origin": sorted(frame["data_origin"].unique().tolist()),
            "farms": int(frame["farm_id"].nunique()),
            "max_abs_parts_minus_carcass": float((part_sum - frame["carcass_weight_kg"]).abs().max()),
            "max_abs_saleable_identity_error": float(
                (saleable - frame["total_saleable_weight"]).abs().max()
            ),
            "saleable_over_carcass": int((saleable > frame["carcass_weight_kg"] + 1e-9).sum()),
            "cutting_yield_pct_min": float(frame["cutting_yield_pct"].min()),
            "cutting_yield_pct_max": float(frame["cutting_yield_pct"].max()),
            "carcass_weight_mean": float(frame["carcass_weight_kg"].mean()),
        }

    REPORTS.mkdir(parents=True, exist_ok=True)
    target = REPORTS / "data_quality.json"
    target.write_text(json.dumps(reports, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    print(json.dumps(reports, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:  # noqa: BLE001
        print(f"data quality failed: {exc}", file=sys.stderr)
        raise
