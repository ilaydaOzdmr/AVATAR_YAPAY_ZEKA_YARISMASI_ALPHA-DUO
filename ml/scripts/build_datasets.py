"""Build the processed datasets used by training and reporting.

Run from the repository root:

    python ml/scripts/build_datasets.py

The production and butchery files are synthetic (no public animal-level raw
data exists); the USDA price files are real public data.
"""

from __future__ import annotations

import json
import sys
from pathlib import Path

ML_ROOT = Path(__file__).resolve().parents[1]
if str(ML_ROOT) not in sys.path:
    sys.path.insert(0, str(ML_ROOT))

from src.synthetic import build_all  # noqa: E402

RAW_DIR = ML_ROOT / "data" / "raw" / "usda"
PROCESSED_DIR = ML_ROOT / "data" / "processed"
REPORTS_DIR = ML_ROOT / "reports"


def main() -> None:
    prices = (
        RAW_DIR
        / "livestock_domestic_data"
        / "livestock-data-machine-readable-files"
        / "WholesalePrices.csv"
    )
    counts = build_all(PROCESSED_DIR, prices)
    REPORTS_DIR.mkdir(parents=True, exist_ok=True)
    summary = {
        "processed_rows": counts,
        "usda_price_file_used": str(prices.relative_to(ML_ROOT)) if prices.exists() else None,
        "price_index_available": prices.exists(),
        "data_origin": {
            "production_orders_synthetic": "synthetic_calibrated",
            "butchery_lamb_synthetic": "synthetic_calibrated",
            "butchery_cattle_synthetic": "synthetic_calibrated",
        },
        "warning": (
            "Synthetic rows are not real business records and must not be "
            "reported as real validation data."
        ),
    }
    (REPORTS_DIR / "dataset_inventory.json").write_text(
        json.dumps(summary, indent=2, ensure_ascii=False) + "\n", encoding="utf-8"
    )
    print(json.dumps(summary, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    main()
