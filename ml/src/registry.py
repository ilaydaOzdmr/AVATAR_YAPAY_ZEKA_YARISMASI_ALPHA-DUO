"""Versioned model artifact loading for the API and evaluation scripts."""

from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any

import joblib

from .config import MODEL_VERSION, PRODUCTION_FEATURES, BUTCHERY_FEATURES


@dataclass
class ModelBundle:
    name: str
    path: Path
    model_version: str
    metadata: dict[str, Any]
    pipelines: dict[str, Any]

    @property
    def status(self) -> str:
        return str(self.metadata.get("model_status", "unknown"))

    @property
    def targets(self) -> list[str]:
        return list(self.pipelines)

    def predict(self, target: str, features):
        if target not in self.pipelines:
            raise KeyError(f"{self.name} has no target {target}")
        return self.pipelines[target].predict(features)


def load_bundle(path: Path, name: str) -> ModelBundle | None:
    model_path = path / "model.joblib"
    metadata_path = path / "metadata.json"
    if not model_path.exists() or not metadata_path.exists():
        return None
    pipelines = joblib.load(model_path)
    metadata = json.loads(metadata_path.read_text(encoding="utf-8"))
    expected_features = PRODUCTION_FEATURES if name == "yield" else BUTCHERY_FEATURES
    if metadata.get("model_version") != MODEL_VERSION or metadata.get("feature_names") != expected_features:
        raise ValueError(f"incompatible model schema at {path}; retrain with python ml/train.py")
    return ModelBundle(
        name=name,
        path=path,
        model_version=str(metadata.get("model_version", "unknown")),
        metadata=metadata,
        pipelines=pipelines,
    )


def load_all(artifacts_dir: Path) -> dict[str, ModelBundle]:
    bundles: dict[str, ModelBundle] = {}
    for name in ("yield", "butchery_lamb", "butchery_cattle"):
        bundle = load_bundle(artifacts_dir / f"{name}_model", name)
        if bundle is not None:
            bundles[name] = bundle
    return bundles
