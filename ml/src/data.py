"""Dataset loading and validation."""

from __future__ import annotations

import hashlib
from pathlib import Path

import pandas as pd
import numpy as np

from .config import (
    BUTCHERY_FEATURES,
    BUTCHERY_GROUP,
    BUTCHERY_TARGETS,
    DATASET_FILES,
    PRODUCTION_FEATURES,
    PRODUCTION_GROUP,
    PRODUCTION_TARGETS,
)

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


def file_sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def load_dataset(name: str) -> pd.DataFrame:
    path = DATASET_FILES[name]
    if not path.exists():
        raise FileNotFoundError(
            f"{path} missing; run `python ml/scripts/build_datasets.py` first"
        )
    return pd.read_csv(path)


def validate_production(frame: pd.DataFrame) -> None:
    required = set(PRODUCTION_FEATURES + PRODUCTION_TARGETS + [PRODUCTION_GROUP])
    missing = required - set(frame.columns)
    if missing:
        raise ValueError(f"production dataset missing columns: {sorted(missing)}")
    if frame[list(required)].isna().any().any():
        raise ValueError("production dataset contains nulls")
    if not frame["yield_pct"].between(0.5, 1.0).all():
        raise ValueError("production yield outside physical range")
    if (frame["finished_qty"] <= 0).any():
        raise ValueError("production finished_qty must be positive")
    if (frame["raw_material_kg"] <= 0).any():
        raise ValueError("production raw_material_kg must be positive")
    numeric = frame[list(required)].select_dtypes(include="number")
    if not np.isfinite(numeric).all().all() or numeric.lt(0).any().any():
        raise ValueError("production quantities must be finite and nonnegative")
    if not np.allclose(frame["finished_qty"], frame["raw_material_kg"] * frame["yield_pct"], atol=0.009):
        raise ValueError("finished quantity and meat-based yield disagree")


def validate_butchery(frame: pd.DataFrame, animal_type: str) -> None:
    required = set(BUTCHERY_FEATURES + BUTCHERY_TARGETS + [BUTCHERY_GROUP, "animal_type"])
    missing = required - set(frame.columns)
    if missing:
        raise ValueError(f"butchery dataset missing columns: {sorted(missing)}")
    if frame[list(required)].isna().any().any():
        raise ValueError("butchery dataset contains nulls")
    if set(frame["animal_type"]) != {animal_type}:
        raise ValueError(f"expected only {animal_type} rows")
    carcass = frame["carcass_weight_kg"]
    parts = frame[PART_COLUMNS].sum(axis=1)
    if ((parts - carcass).abs() > 1e-6).any():
        raise ValueError("butchery part weights do not sum to carcass weight")
    saleable = frame[SALEABLE_COLUMNS].sum(axis=1)
    if (saleable > carcass + 1e-9).any():
        raise ValueError("butchery saleable weight exceeds carcass")
    if frame[BUTCHERY_TARGETS].lt(0).any().any():
        raise ValueError("butchery targets must be non-negative")
    if not np.isfinite(frame[list(required)].select_dtypes(include="number")).all().all():
        raise ValueError("butchery quantities must be finite")
    if carcass.le(0).any() or (carcass > frame["live_weight_kg"]).any():
        raise ValueError("carcass must be positive and no larger than live weight")
    if not np.allclose(saleable, frame["total_saleable_weight"], atol=1e-6):
        raise ValueError("saleable total does not match parts")
    if not np.allclose(100 * saleable / carcass, frame["cutting_yield_pct"], atol=0.001):
        raise ValueError("cutting yield does not match parts")


def production_frame() -> pd.DataFrame:
    frame = load_dataset("yield")
    validate_production(frame)
    return frame


def butchery_frame(animal_type: str) -> pd.DataFrame:
    name = f"butchery_{animal_type}"
    frame = load_dataset(name)
    validate_butchery(frame, animal_type)
    return frame


def dataset_version(name: str) -> str:
    return file_sha256(DATASET_FILES[name])[:16]
