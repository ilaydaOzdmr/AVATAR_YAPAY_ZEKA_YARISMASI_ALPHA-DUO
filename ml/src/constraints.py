"""Physical identities shared by API inference and held-out evaluation."""
from __future__ import annotations

import numpy as np

from .data import PART_COLUMNS, SALEABLE_COLUMNS


def production_prediction(raw_kg: float, predictions: dict[str, float]) -> dict[str, float]:
    value = float(predictions["yield_pct"])
    if not np.isfinite(value):
        raise ValueError("model returned a non-finite yield")
    ratio = float(np.clip(value, 0.0, 1.0))
    # This model's denominator is meat input, not meat+spice+casing.
    finished = round(raw_kg * ratio, 3)
    return {"finished_qty": finished, "yield_pct": finished / raw_kg}


def butchery_prediction(carcass_kg: float, predictions: dict[str, float]) -> dict[str, float]:
    weights = np.array([predictions[part] for part in PART_COLUMNS], dtype=float)
    if not np.isfinite(weights).all():
        raise ValueError("model returned non-finite part weights")
    weights = np.maximum(weights, 0)
    if weights.sum() <= 0:
        raise ValueError("model returned no positive part weights")
    # Reconcile all eight parts to measured carcass mass. Largest remainder
    # allocation makes milligram rounding exact without negative waste.
    total_units = round(carcass_kg * 1000)
    exact = total_units * weights / weights.sum()
    units = np.floor(exact).astype(int)
    remaining = total_units - int(units.sum())
    order = np.argsort(-(exact - units), kind="stable")
    units[order[:remaining]] += 1
    result = {part: int(unit) / 1000 for part, unit in zip(PART_COLUMNS, units)}
    result["total_saleable_weight"] = round(sum(result[p] for p in SALEABLE_COLUMNS), 3)
    result["cutting_yield_pct"] = 100 * result["total_saleable_weight"] / (total_units / 1000)
    return result
