"""Prediction service: artifact loading, validation, uncertainty and warnings."""

from __future__ import annotations

import json
from functools import lru_cache
from pathlib import Path
from typing import Any

import pandas as pd

from .config import ARTIFACTS_DIR, DATA_ORIGIN
from .constraints import butchery_prediction, production_prediction
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
            deployed = bundle.metadata["deployed_metrics"][target]
            return {"model_mae": deployed["test"]["mae"], "model_rmse": deployed["test"]["rmse"],
                "model_r2": deployed["test"]["r2"],
                "baseline_mae": deployed["proportional_baseline_test"]["mae"],
                "baseline_rmse": deployed["proportional_baseline_test"]["rmse"],
                "baseline_kind": "train-only proportional physical baseline"}

        return {
            "yield": legacy_metrics(yield_bundle, "yield_pct"),
            "butchery": legacy_metrics(lamb, "total_saleable_weight"),
            "butchery_lamb": legacy_metrics(lamb, "total_saleable_weight"),
            "butchery_cattle": legacy_metrics(cattle, "total_saleable_weight"),
            "note": WARNING_SYNTHETIC,
            "model_version": yield_bundle.model_version,
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
        if "input_ranges" in bundle.metadata:
            return bundle.metadata["input_ranges"]
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
        deployed = bundle.metadata.get("deployed_metrics", {}).get(target, {}).get("test", {})
        rmse = deployed.get("rmse")
        if rmse is None:
            return None
        rmse = float(rmse)
        return {
            "method": "illustrative ±1.96 × deployed held-out RMSE; synthetic; coverage has not been calibrated",
            "illustrative_error_band": {"lower": round(-1.96 * rmse, 4), "upper": round(1.96 * rmse, 4)},
            "coverage_validated": False,
            "test_rmse": rmse,
            "test_mae": deployed.get("mae"),
        }

    # -- production yield ------------------------------------------------
    def _yield_costs(self, request: YieldV2Request, finished_qty: float) -> dict[str, float]:
        material_meat = request.raw_material_kg * request.material_unit_price_meat
        material_spice = request.spice_kg * request.material_unit_price_spice
        material_casing = request.casing_kg * request.material_unit_price_casing
        material_other = request.other_material_kg * request.material_unit_price_other
        service_cost = (finished_qty * request.service_unit_cost
                        if request.service_unit_cost is not None else request.service_cost)
        total = (
            material_meat
            + material_spice
            + material_casing
            + material_other
            + service_cost
            + request.labor_cost
        )
        return {
            "material_meat": round(material_meat, 2),
            "material_spice": round(material_spice, 2),
            "material_casing": round(material_casing, 2),
            "material_other": round(material_other, 2),
            "service_cost": round(service_cost, 2),
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
        physical = production_prediction(request.raw_material_kg, {
            "yield_pct": float(bundle.predict("yield_pct", row)[0]),
        })
        finished_qty, yield_pct = physical["finished_qty"], physical["yield_pct"]
        if finished_qty <= 0:
            raise ValueError("predicted finished quantity is zero; cannot estimate EKP")

        costs = self._yield_costs(request, finished_qty)
        estimated_ekp = costs["total_cost"] / finished_qty if finished_qty > 0 else 0.0
        warnings = [WARNING_SYNTHETIC]
        defaults = sorted(set(YieldV2Request.model_fields) - request.model_fields_set - {"actual_finished_qty", "service_unit_cost"})
        if defaults:
            warnings.append(f"Unspecified fields use documented demo defaults, not observed values: {', '.join(defaults)}.")
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
            actual_costs = self._yield_costs(request, request.actual_finished_qty)
            response["accounting_cost_breakdown"] = actual_costs
            response["accounting_ekp"] = round(actual_costs["total_cost"] / request.actual_finished_qty, 4)
            error = finished_qty - request.actual_finished_qty
            response["comparison"] = {"actual": request.actual_finished_qty, "predicted": finished_qty,
                "signed_error": round(error, 3), "absolute_error": round(abs(error), 3),
                "percentage_error": round(abs(error) / request.actual_finished_qty * 100, 3)}
            response["ekp_kind"] = "actual_accounting"
        return response

    def predict_yield_legacy(self, request: LegacyYieldRequest) -> dict[str, Any]:
        finished = self.predict_yield_v2(
            YieldV2Request(
                raw_material_kg=request.raw_material_kg,
                service_cost=request.service_cost if request.service_cost is not None else 0,
                service_unit_cost=5.50 if request.service_cost is None else None,
            )
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
        if request.breed is None or request.age_months is None or "sex" not in request.model_fields_set:
            warnings.append("Missing breed/age/sex use synthetic demo defaults; provide measured attributes for a specific animal.")
        features = {
            "breed": request.breed or DEFAULT_BREED[request.animal_type],
            "sex": request.sex,
            "age_months": request.age_months
            if request.age_months is not None
            else DEFAULT_AGE[request.animal_type],
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
        raw_predictions = {
            target: float(bundle.predict(target, row)[0])
            for target in bundle.metadata["target_names"]
        }

        carcass = float(features["carcass_weight_kg"])
        predictions = butchery_prediction(carcass, raw_predictions)
        warnings.append("Part weights reconciled to measured carcass; saleable total and yield derived from those parts.")
        parts_sum = sum(max(0.0, predictions.get(target, 0.0)) for target in PART_TARGETS)
        saleable = sum(max(0.0, predictions.get(target, 0.0)) for target in SALEABLE_TARGETS)
        physical_check = {
            "carcass_weight_kg": round(carcass, 3),
            "parts_sum_kg": round(parts_sum, 3),
            "saleable_sum_kg": round(saleable, 3),
            "parts_over_carcass": parts_sum > carcass + 0.000501,
            "reconciliation": "nonnegative proportional allocation; largest remainder at 0.001 kg",
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
                carcass_weight_kg=request.input_weight,
            )
        )
        parts = [
            {"product_code": PRODUCT_MAP[target], "weight": result["prediction"][target]}
            for target in PRODUCT_MAP
        ]
        return {**result, "parts": parts}
