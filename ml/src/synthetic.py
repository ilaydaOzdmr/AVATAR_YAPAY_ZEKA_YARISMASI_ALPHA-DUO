"""Consistent, documented synthetic datasets for gaps without public raw data.

Nothing here is a real business record. Every generated row carries
``data_origin = synthetic_calibrated`` and the generator version. Public USDA
price series are used only as an exogenous price index; they are not production
orders or carcass measurements.
"""

from __future__ import annotations

import csv
from pathlib import Path

import numpy as np
import pandas as pd

GENERATOR_VERSION = "synthetic-v2"

PRODUCTION_BASE_YIELD = {"sucuk": 0.84, "salam": 0.87, "sausage": 0.90}
PRODUCTION_SERVICE_PER_KG = {"sucuk": 5.50, "salam": 5.00, "sausage": 4.50}
PRODUCTION_PRICES = {
    "meat": 9.70,
    "spice": 24.00,
    "casing": 11.60,
    "other": 3.50,
}

BUTCHERY_SPECS = {
    "lamb": {
        "live_mean": 45.0,
        "live_sd": 6.0,
        "live_min": 30.0,
        "live_max": 65.0,
        "dressing_mean": 0.47,
        "dressing_sd": 0.015,
        "breeds": ["Romney", "Merino", "Crossbred", "Suffolk"],
        "ratios": {
            "leg_weight": 0.18,
            "loin_weight": 0.12,
            "shoulder_weight": 0.14,
            "rib_weight": 0.08,
            "mince_weight": 0.12,
            "diced_meat_weight": 0.07,
            "bone_weight": 0.13,
            "waste_weight": 0.16,
        },
    },
    "cattle": {
        "live_mean": 560.0,
        "live_sd": 55.0,
        "live_min": 420.0,
        "live_max": 720.0,
        "dressing_mean": 0.58,
        "dressing_sd": 0.02,
        "breeds": ["Angus", "Hereford", "Holstein", "Limousin"],
        "ratios": {
            "leg_weight": 0.16,
            "loin_weight": 0.13,
            "shoulder_weight": 0.13,
            "rib_weight": 0.09,
            "mince_weight": 0.12,
            "diced_meat_weight": 0.08,
            "bone_weight": 0.12,
            "waste_weight": 0.17,
        },
    },
}

TARGET_PARTS = [
    "leg_weight",
    "loin_weight",
    "shoulder_weight",
    "rib_weight",
    "mince_weight",
    "diced_meat_weight",
    "bone_weight",
    "waste_weight",
]
SALEABLE_PARTS = [
    "leg_weight",
    "loin_weight",
    "shoulder_weight",
    "rib_weight",
    "mince_weight",
    "diced_meat_weight",
]


def load_usda_beef_price_index(path: Path) -> pd.Series:
    """Return a mean-1.0 monthly boxed-beef price index from USDA ERS data.

    Falls back to a flat index (1.0) when the public file is absent, so dataset
    generation remains reproducible without network access.
    """
    if not path.exists():
        return pd.Series([1.0], name="price_index")

    rows: list[tuple[str, float]] = []
    with path.open("r", encoding="utf-8-sig", newline="") as handle:
        for row in csv.DictReader(handle):
            if row.get("commodity_desc") != "Beef":
                continue
            attribute = (row.get("attribute_desc") or "").lower()
            if "boxed beef cutout choice" not in attribute:
                continue
            try:
                rows.append(
                    (
                        f"{int(row['year_id']):04d}-{int(row['month_id']):02d}",
                        float(row["amount"]),
                    )
                )
            except (KeyError, TypeError, ValueError):
                continue
    if not rows:
        return pd.Series([1.0], name="price_index")

    series = pd.Series(dict(rows), name="price_index").sort_index()
    return series / series.mean()


def generate_production(
    n_orders: int = 480,
    n_batches: int = 60,
    seed: int = 42,
    price_index: pd.Series | None = None,
    outcome_noise_scale: float = 0.012,
) -> pd.DataFrame:
    """Generate synthetic production orders with a documented mechanism.

    The target ``finished_qty`` depends on raw material, product type, spice
    load, a batch effect and measurement noise. ``actual_ekp`` is the accounting
    identity ``total_cost / finished_qty``; it is not a model prediction.
    """
    rng = np.random.default_rng(seed)
    batch_ids = [f"BATCH-{i + 1:03d}" for i in range(n_batches)]
    batch_effect = rng.normal(0.0, 0.012, n_batches)
    product_types = list(PRODUCTION_BASE_YIELD)

    if price_index is None or price_index.empty:
        price_index = pd.Series([1.0], name="price_index")
    index_values = price_index.to_numpy(dtype=float)

    records: list[dict[str, object]] = []
    for i in range(n_orders):
        batch_idx = int(rng.integers(0, n_batches))
        product_type = str(rng.choice(product_types, p=[0.65, 0.20, 0.15]))
        raw_kg = float(rng.uniform(60.0, 160.0))
        spice_ratio = float(rng.uniform(0.010, 0.030))
        other_ratio = float(rng.uniform(0.000, 0.020))
        spice_kg = raw_kg * spice_ratio
        other_kg = raw_kg * other_ratio

        yield_pct = (
            PRODUCTION_BASE_YIELD[product_type]
            + batch_effect[batch_idx]
            - 0.35 * (spice_ratio - 0.020)
            - 0.20 * (other_ratio - 0.010)
            + float(rng.normal(0.0, outcome_noise_scale))
        )
        yield_pct = float(np.clip(yield_pct, 0.72, 0.94))
        finished_qty = raw_kg * yield_pct

        casing_ratio = float(rng.uniform(0.020, 0.050))
        casing_kg = raw_kg * casing_ratio
        planned_qty = raw_kg * PRODUCTION_BASE_YIELD[product_type]
        service_cost = planned_qty * PRODUCTION_SERVICE_PER_KG[product_type] * float(
            rng.normal(1.0, 0.08)
        )
        labor_cost = raw_kg * float(rng.uniform(0.80, 1.60))

        price_index_value = float(index_values[int(rng.integers(0, len(index_values)))])
        meat_price = PRODUCTION_PRICES["meat"] * price_index_value * float(rng.normal(1.0, 0.03))
        spice_price = PRODUCTION_PRICES["spice"] * float(rng.normal(1.0, 0.04))
        casing_price = PRODUCTION_PRICES["casing"] * float(rng.normal(1.0, 0.04))
        other_price = PRODUCTION_PRICES["other"] * float(rng.normal(1.0, 0.05))

        total_cost = (
            raw_kg * meat_price
            + spice_kg * spice_price
            + casing_kg * casing_price
            + other_kg * other_price
            + service_cost
            + labor_cost
        )
        actual_ekp = total_cost / finished_qty if finished_qty > 0 else 0.0

        records.append(
            {
                "order_id": f"SYN-URT-{i + 1:04d}",
                "production_batch": batch_ids[batch_idx],
                "product_type": product_type,
                "raw_material_kg": round(raw_kg, 3),
                "spice_kg": round(spice_kg, 3),
                "casing_kg": round(casing_kg, 3),
                "other_material_kg": round(other_kg, 3),
                "material_unit_price_meat": round(meat_price, 2),
                "material_unit_price_spice": round(spice_price, 2),
                "material_unit_price_casing": round(casing_price, 2),
                "material_unit_price_other": round(other_price, 2),
                "service_cost": round(service_cost, 2),
                "labor_cost": round(labor_cost, 2),
                "usda_price_index": round(price_index_value, 4),
                "expected_qty": round(planned_qty, 3),
                "finished_qty": round(finished_qty, 3),
                "yield_pct": round(yield_pct, 4),
                "actual_total_cost": round(total_cost, 2),
                "actual_ekp": round(actual_ekp, 2),
                "data_origin": "synthetic_calibrated",
                "generator_version": GENERATOR_VERSION,
            }
        )
    return pd.DataFrame.from_records(records)


def generate_butchery(
    animal_type: str,
    n_animals: int = 800,
    n_farms: int = 12,
    seed: int = 42,
) -> pd.DataFrame:
    """Generate synthetic animal-level cut weights with physical constraints.

    Cut ratios are calibrated to published ranges for lamb/cattle primals, not
    copied from any real record. Part weights are normalised to the generated
    carcass weight so that saleable + bone + waste equals the carcass exactly.
    """
    if animal_type not in BUTCHERY_SPECS:
        raise ValueError(f"unknown animal_type: {animal_type}")
    spec = BUTCHERY_SPECS[animal_type]
    rng = np.random.default_rng(seed)
    farms = [f"FARM-{i + 1:02d}" for i in range(n_farms)]
    farm_effect = rng.normal(0.0, 0.01, n_farms)

    records: list[dict[str, object]] = []
    for i in range(n_animals):
        farm_idx = int(rng.integers(0, n_farms))
        sex = str(rng.choice(["male", "female"], p=[0.55, 0.45]))
        breed = str(rng.choice(spec["breeds"]))
        conformation = float(rng.integers(1, 6))  # 1 (poor) .. 5 (excellent)
        fat_score = float(rng.integers(1, 6))  # 1 (lean) .. 5 (fat)
        conf_norm = (conformation - 3.0) / 2.0
        fat_norm = (fat_score - 3.0) / 2.0

        live_weight = float(
            np.clip(
                rng.normal(spec["live_mean"], spec["live_sd"]),
                spec["live_min"],
                spec["live_max"],
            )
        )
        dressing = (
            spec["dressing_mean"]
            + 0.010 * conf_norm
            - 0.006 * fat_norm
            + farm_effect[farm_idx]
            + float(rng.normal(0.0, spec["dressing_sd"]))
        )
        dressing = float(np.clip(dressing, 0.40, 0.65))
        carcass_weight = live_weight * dressing
        age_months = float(
            np.clip(
                rng.normal(9.0, 1.5) if animal_type == "lamb" else rng.normal(24.0, 3.0),
                4.0,
                36.0,
            )
        )

        ratios: dict[str, float] = {}
        for part, base in spec["ratios"].items():
            adjustment = 0.0
            if part in ("bone_weight", "waste_weight"):
                adjustment = -0.010 * conf_norm + 0.008 * fat_norm
            else:
                adjustment = 0.008 * conf_norm - 0.004 * fat_norm
            ratios[part] = max(0.001, base + adjustment + float(rng.normal(0.0, 0.006)))

        ratio_sum = sum(ratios.values())
        weights = {part: carcass_weight * value / ratio_sum for part, value in ratios.items()}
        carcass_rounded = round(carcass_weight, 3)
        rounded_weights = {
            part: round(weight, 3) for part, weight in weights.items() if part != "waste_weight"
        }
        # Make the physical identity exact after rounding: waste absorbs the
        # remaining milligrams. It is still positive for every generated animal.
        rounded_weights["waste_weight"] = round(
            carcass_rounded - sum(rounded_weights.values()), 3
        )
        if rounded_weights["waste_weight"] < 0:
            raise ValueError("generated cut weights exceed carcass weight")
        total_saleable = sum(rounded_weights[part] for part in SALEABLE_PARTS)
        cutting_yield_pct = 100.0 * total_saleable / carcass_rounded

        record: dict[str, object] = {
            "animal_id": f"{animal_type.upper()}-{i + 1:05d}",
            "farm_id": farms[farm_idx],
            "animal_type": animal_type,
            "breed": breed,
            "sex": sex,
            "age_months": round(age_months, 1),
            "live_weight_kg": round(live_weight, 3),
            "carcass_weight_kg": carcass_rounded,
            "conformation_score": conformation,
            "fat_score": fat_score,
            **{part: rounded_weights[part] for part in TARGET_PARTS},
            "total_saleable_weight": round(total_saleable, 3),
            "cutting_yield_pct": round(cutting_yield_pct, 3),
            "data_origin": "synthetic_calibrated",
            "generator_version": GENERATOR_VERSION,
        }
        records.append(record)
    return pd.DataFrame.from_records(records)


def build_all(out_dir: Path, usda_prices_path: Path, seed: int = 42) -> dict[str, int]:
    """Generate and write all processed datasets. Returns row counts."""
    out_dir.mkdir(parents=True, exist_ok=True)
    price_index = load_usda_beef_price_index(usda_prices_path)

    production = generate_production(seed=seed, price_index=price_index)
    lamb = generate_butchery("lamb", seed=seed + 1)
    cattle = generate_butchery("cattle", seed=seed + 2)

    production.to_csv(out_dir / "production_orders_synthetic.csv", index=False)
    lamb.to_csv(out_dir / "butchery_lamb_synthetic.csv", index=False)
    cattle.to_csv(out_dir / "butchery_cattle_synthetic.csv", index=False)
    return {
        "production_orders_synthetic": len(production),
        "butchery_lamb_synthetic": len(lamb),
        "butchery_cattle_synthetic": len(cattle),
    }
