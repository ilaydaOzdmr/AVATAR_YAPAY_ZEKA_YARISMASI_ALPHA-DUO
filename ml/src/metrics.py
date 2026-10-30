"""Regression metrics with safe MAPE handling."""

from __future__ import annotations

import math

import numpy as np
from sklearn.metrics import (
    mean_absolute_error,
    mean_squared_error,
    median_absolute_error,
    r2_score,
)


def regression_metrics(y_true, y_pred) -> dict[str, float]:
    y_true = np.asarray(y_true, dtype=float)
    y_pred = np.asarray(y_pred, dtype=float)
    mask = np.abs(y_true) > 1e-9
    mape = float(np.mean(np.abs((y_true[mask] - y_pred[mask]) / y_true[mask])) * 100) if mask.any() else float("nan")
    return {
        "mae": float(mean_absolute_error(y_true, y_pred)),
        "rmse": float(math.sqrt(mean_squared_error(y_true, y_pred))),
        "r2": float(r2_score(y_true, y_pred)) if len(y_true) > 1 else float("nan"),
        "medae": float(median_absolute_error(y_true, y_pred)),
        "mape": mape,
    }


def rounded(metrics: dict[str, float], digits: int = 4) -> dict[str, float | None]:
    out: dict[str, float | None] = {}
    for key, value in metrics.items():
        out[key] = None if math.isnan(value) else round(value, digits)
    return out
