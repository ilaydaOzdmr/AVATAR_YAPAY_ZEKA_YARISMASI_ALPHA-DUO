"""Prediction service: artifact loading, validation, uncertainty and warnings."""

from __future__ import annotations

import json
from functools import lru_cache
from pathlib import Path
from typing import Any

import pandas as pd

from .config import ARTIFACTS_DIR, DATA_ORIGIN
from .data import butchery_frame, production_frame
from .registry import ModelBundle, load_all
from .schemas import (
    ButcheryV2Request,
    LegacyButcheryRequest,
    LegacyYieldRequest,
    YieldV2Request,
)

PRODUCT_MAP = {
    "leg_weight": "PRD-101",
    "loin_weight": "PRD-102",
    "shoulder_weight": "PRD-103",
    "rib_weight": "PRD-104",
    "mince_weight": "PRD-105",
    "diced_meat_weight": "PRD-106",
    "bone_weight": "PRD-107",
}
PART_TARGETS = [
    "leg_weight",
    "loin_weight",
    "shoulder_weight",
    "rib_weight",
    "mince_weight",
    "diced_meat_weight",
    "bone_weight",
    "waste_weight",
]
SALEABLE_TARGETS = [
    "leg_weight",
    "loin_weight",
    "shoulder_weight",
    "rib_weight",
    "mince_weight",
    "diced_meat_weight",
]
DRESSING = {"lamb": 0.47, "cattle": 0.58}
DEFAULT_AGE = {"lamb": 9.0, "cattle": 24.0}
DEFAULT_BREED = {"lamb": "Crossbred", "cattle": "Holstein"}
WARNING_SYNTHETIC = (
    "Model is trained on synthetic_calibrated data and is not validated on real "
    "business records."
)


class ModelUnavailableError(RuntimeError):
    """Raised when an artifact is missing; the API maps this to HTTP 503."""


class MLService:
    def __init__(self, artifacts_dir: Path = ARTIFACTS_DIR) -> None:
        self.artifacts_dir = artifacts_dir
        self.bundles = load_all(artifacts_dir)

    def require(self, name: str) -> ModelBundle:
        bundle = self.bundles.get(name)
        if bundle is None:
            raise ModelUnavailableError(
                f"model artifact '{name}' is not available; run `python ml/train.py`"
            )
        return bundle

    # -- metadata ---------------------------------------------------------
    def models(self) -> list[dict[str, Any]]:
        out = []
        for name, bundle in self.bundles.items():
            metadata = bundle.metadata
            out.append(
                {
                    "name": name,
                    "model_version": bundle.model_version,
                    "model_status": bundle.status,
                    "dataset_version": metadata.get("dataset_version", "unknown"),
                    "targets": bundle.targets,
                    "training_sample_count": metadata.get("training_sample_count", 0),
                    "test_sample_count": metadata.get("test_sample_count", 0),
                    "warnings": metadata.get("warnings", []),
                }
            )
        return out

    def health(self) -> dict[str, Any]:
        return {
            "status": "ok",
            "models_loaded": len(self.bundles),
            "models": self.models(),
        }

    def report(self) -> dict[str, Any]:
        yield_bundle = self.require("yield")
        lamb = self.require("butchery_lamb")
        cattle = self.require("butchery_cattle")

        def legacy_metrics(bundle: ModelBundle, target: str) -> dict[str, Any]:
            report = bundle.metadata["targets"][target]
            selected = report["selected"]
            baseline = next(
                (
                    row
                    for row in report["candidates"]
                    if row.get("ok") and row["model"] == "dummy_mean"
                ),
                None,
            )
            return {
                "model_mae": selected["test_mae"],
                "model_rmse": selected["test_rmse"],
                "model_r2": selected.get("test_r2"),
                "baseline_mae": baseline["test_mae"] if baseline else None,
                "baseline_rmse": baseline["test_rmse"] if baseline else None,
            }

        return {
            "yield": legacy_metrics(yield_bundle, "yield_pct"),
            "butchery": legacy_metrics(lamb, "total_saleable_weight"),
            "butchery_lamb": legacy_metrics(lamb, "total_saleable_weight"),
            "butchery_cattle": legacy_metrics(cattle, "total_saleable_weight"),
            "note": WARNING_SYNTHETIC,
            "model_status": "synthetic_only",
            "data_origin": DATA_ORIGIN,
            "groups": {
                "yield": yield_bundle.metadata["targets"],
                "butchery_lamb": lamb.metadata["targets"],
                "butchery_cattle": cattle.metadata["targets"],
            },
        }

    # -- input ranges -----------------------------------------------------
    @lru_cache(maxsize=3)
    def _frame(self, dataset_name: str) -> pd.DataFrame:
        if dataset_name == "yield":
            return production_frame()
        if dataset_name == "butchery_lamb":
            return butchery_frame("lamb")
        if dataset_name == "butchery_cattle":
            return butchery_frame("cattle")
        raise KeyError(dataset_name)

    def _training_frame(self, bundle: ModelBundle) -> pd.DataFrame:
        metadata = bundle.metadata
        frame = self._frame(metadata["dataset_name"])
        split_path = bundle.path / "split.json"
        if split_path.exists():
            split = json.loads(split_path.read_text(encoding="utf-8"))
            return frame.iloc[split["train_indices"]]
        return frame

    def input_ranges(self, bundle: ModelBundle) -> dict[str, Any]:
        frame = self._training_frame(bundle)
        ranges: dict[str, Any] = {}
        for feature in bundle.metadata["feature_names"]:
            series = frame[feature]
            if pd.api.types.is_numeric_dtype(series):
                ranges[feature] = {
                    "min": round(float(series.min()), 4),
                    "max": round(float(series.max()), 4),
                    "mean": round(float(series.mean()), 4),
                }
            else:
                ranges[feature] = {"allowed": sorted(series.astype(str).unique().tolist())}
        return ranges

    def _range_warnings(
        self, bundle: ModelBundle, features: dict[str, Any]
    ) -> list[str]:
        warnings: list[str] = []
        ranges = self.input_ranges(bundle)
        for name, value in features.items():
            spec = ranges.get(name)
            if not spec:
                continue
            if "allowed" in spec and str(value) not in spec["allowed"]:
                warnings.append(
                    f"{name}={value!r} is outside the training categories "
                    f"{spec['allowed']}; prediction is extrapolated."
                )
            if "min" in spec and isinstance(value, (int, float)):
                if value < spec["min"] or value > spec["max"]:
                    warnings.append(
                        f"{name}={value} is outside the training range "
                        f"[{spec['min']}, {spec['max']}]; prediction is extrapolated."
                    )
        return warnings

    def _uncertainty(self, bundle: ModelBundle, target: str) -> dict[str, Any] | None:
        report = bundle.metadata.get("targets", {}).get(target)
        if not report:
            return None
        selected = report.get("selected", {})
        rmse = selected.get("test_rmse")
        if rmse is None:
            return None
        rmse = float(rmse)
        return {
            "method": "±1.96 × held-out test RMSE (synthetic)",
            "roughly_95pct": {"lower": round(-1.96 * rmse, 4), "upper": round(1.96 * rmse, 4)},
            "test_rmse": rmse,
            "test_mae": selected.get("test_mae"),
        }

    # -- production yield ------------------------------------------------
    def _yield_costs(self, request: YieldV2Request, finished_qty: float) -> dict[str, float]:
        material_meat = request.raw_material_kg * request.material_unit_price_meat
        material_spice = request.spice_kg * request.material_unit_price_spice
        material_casing = request.casing_kg * request.material_unit_price_casing
        material_other = request.other_material_kg * request.material_unit_price_other
        total = (
            material_meat
            + material_spice
            + material_casing
            + material_other
            + request.service_cost
            + request.labor_cost
        )
        return {
            "material_meat": round(material_meat, 2),
            "material_spice": round(material_spice, 2),
            "material_casing": round(material_casing, 2),
            "material_other": round(material_other, 2),
            "service_cost": round(request.service_cost, 2),
            "labor_cost": round(request.labor_cost, 2),
            "total_cost": round(total, 2),
        }

    def predict_yield_v2(self, request: YieldV2Request) -> dict[str, Any]:
        bundle = self.require("yield")
        features = {
            "raw_material_kg": request.raw_material_kg,
            "spice_kg": request.spice_kg,
            "casing_kg": request.casing_kg,
            "other_material_kg": request.other_material_kg,
            "material_unit_price_meat": request.material_unit_price_meat,
            "material_unit_price_spice": request.material_unit_price_spice,
            "material_unit_price_casing": request.material_unit_price_casing,
            "material_unit_price_other": request.material_unit_price_other,
            "service_cost": request.service_cost,
            "labor_cost": request.labor_cost,
            "product_type": request.product_type,
        }
        row = pd.DataFrame([features])[bundle.metadata["feature_names"]]
        finished_qty = float(bundle.predict("finished_qty", row)[0])
        yield_pct = float(bundle.predict("yield_pct", row)[0])

        costs = self._yield_costs(request, finished_qty)
        estimated_ekp = costs["total_cost"] / finished_qty if finished_qty > 0 else 0.0
        warnings = [WARNING_SYNTHETIC]
        warnings.extend(self._range_warnings(bundle, features))
        if not (0.5 <= yield_pct <= 1.0):
            warnings.append(
                f"Predicted yield_pct={yield_pct:.3f} is outside the physical 0.5–1.0 range."
            )
        if finished_qty <= 0:
            warnings.append("Predicted finished_qty is not positive.")

        response: dict[str, Any] = {
            "prediction": {
                "finished_qty": round(finished_qty, 3),
                "yield_pct": round(yield_pct, 4),
                "estimated_ekp": round(estimated_ekp, 4),
            },
            "model_version": bundle.model_version,
            "model_status": bundle.status,
            "data_origin": DATA_ORIGIN,
            "applicable_input_range": self.input_ranges(bundle),
            "warnings": warnings,
            "uncertainty": {
                "finished_qty": self._uncertainty(bundle, "finished_qty"),
                "yield_pct": self._uncertainty(bundle, "yield_pct"),
            },
            "cost_breakdown": costs,
            "ekp_kind": "estimated_pre_production",
        }
        if request.actual_finished_qty is not None:
            response["accounting_ekp"] = round(
                costs["total_cost"] / request.actual_finished_qty, 4
            )
            response["ekp_kind"] = "actual_accounting"
        return response

    def predict_yield_legacy(self, request: LegacyYieldRequest) -> dict[str, Any]:
        finished = self.predict_yield_v2(
            YieldV2Request(
                raw_material_kg=request.raw_material_kg,
                service_cost=request.service_cost or 0,
            )
        )
        predicted_qty = finished["prediction"]["finished_qty"]
        if request.service_cost is None:
            service = round(5.50 * predicted_qty, 2)
            finished = self.predict_yield_v2(
                YieldV2Request(raw_material_kg=request.raw_material_kg, service_cost=service)
            )
        return {
            **finished,
            "predicted_qty": finished["prediction"]["finished_qty"],
            "predicted_unit_cost": finished["prediction"]["estimated_ekp"],
        }

    # -- butchery --------------------------------------------------------
    def _butchery_features(self, request: ButcheryV2Request) -> tuple[dict[str, Any], list[str]]:
        warnings: list[str] = []
        carcass = request.carcass_weight_kg
        if carcass is None:
            carcass = request.live_weight_kg * DRESSING[request.animal_type]
            warnings.append(
                f"carcass_weight_kg was not provided; estimated as {carcass:.3f} kg "
                f"using dressing ratio {DRESSING[request.animal_type]}."
            )
        features = {
            "breed": request.breed or DEFAULT_BREED[request.animal_type],
            "sex": request.sex,
            "age_months": request.age_months
            if request.age_months is not None
            else DEFAULT_AGE[request.animal_type],
            "live_weight_kg": request.live_weight_kg,
            "carcass_weight_kg": carcass,
            "conformation_score": request.conformation_score,
            "fat_score": request.fat_score,
        }
        return features, warnings

    def predict_butchery_v2(self, request: ButcheryV2Request) -> dict[str, Any]:
        bundle = self.require(f"butchery_{request.animal_type}")
        features, warnings = self._butchery_features(request)
        warnings.insert(0, WARNING_SYNTHETIC)
        warnings.extend(self._range_warnings(bundle, features))
        row = pd.DataFrame([features])[bundle.metadata["feature_names"]]
        predictions = {
            target: round(float(bundle.predict(target, row)[0]), 3)
            for target in bundle.metadata["target_names"]
        }

        carcass = float(features["carcass_weight_kg"])
        parts_sum = sum(max(0.0, predictions.get(target, 0.0)) for target in PART_TARGETS)
        saleable = sum(max(0.0, predictions.get(target, 0.0)) for target in SALEABLE_TARGETS)
        physical_check = {
            "carcass_weight_kg": round(carcass, 3),
            "parts_sum_kg": round(parts_sum, 3),
            "saleable_sum_kg": round(saleable, 3),
            "parts_over_carcass": parts_sum > carcass * 1.05,
        }
        if physical_check["parts_over_carcass"]:
            warnings.append(
                "Predicted part weights exceed the carcass by more than 5%; "
                "the values were not silently adjusted."
            )
        if any(value < 0 for value in predictions.values()):
            warnings.append("One or more predicted weights are negative; do not use as-is.")

        return {
            "prediction": predictions,
            "model_version": bundle.model_version,
            "model_status": bundle.status,
            "data_origin": DATA_ORIGIN,
            "applicable_input_range": self.input_ranges(bundle),
            "warnings": warnings,
            "uncertainty": {
                target: self._uncertainty(bundle, target)
                for target in ("total_saleable_weight", "cutting_yield_pct")
            },
            "physical_check": physical_check,
        }

    def predict_butchery_legacy(self, request: LegacyButcheryRequest) -> dict[str, Any]:
        aliases = {"kuzu": "lamb", "dana": "cattle", "lamb": "lamb", "cattle": "cattle"}
        animal_type = aliases.get(request.animal_type.strip().lower())
        if animal_type not in ("lamb", "cattle"):
            raise ValueError(f"unknown animal type: {request.animal_type}")
        result = self.predict_butchery_v2(
            ButcheryV2Request(
                animal_type=animal_type,  # type: ignore[arg-type]
                live_weight_kg=request.input_weight,
            )
        )
        parts = [
            {"product_code": PRODUCT_MAP[target], "weight": result["prediction"][target]}
            for target in PRODUCT_MAP
        ]
        return {**result, "parts": parts}
