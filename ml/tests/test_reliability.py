import json

import numpy as np
import pandas as pd
import pytest

from ml.src.config import ARTIFACTS_DIR, PRODUCTION_FEATURES
from ml.src.constraints import butchery_prediction, production_prediction
from ml.src.data import PART_COLUMNS, SALEABLE_COLUMNS, production_frame, validate_production
from ml.src.schemas import ButcheryV2Request, YieldV2Request
from ml.src.service import MLService
from ml.src.models import train_target


@pytest.mark.parametrize("kind,carcass", [("lamb", 21.123), ("cattle", 325.357)])
def test_reconciliation_for_negative_and_excessive_parts(kind, carcass):
    raw = {part: (1000 if i % 2 else -50) for i, part in enumerate(PART_COLUMNS)}
    result = butchery_prediction(carcass, raw)
    assert all(result[part] >= 0 for part in PART_COLUMNS)
    assert sum(result[part] for part in PART_COLUMNS) == pytest.approx(carcass)
    assert result["total_saleable_weight"] == pytest.approx(sum(result[part] for part in SALEABLE_COLUMNS))
    assert result["cutting_yield_pct"] == pytest.approx(100 * result["total_saleable_weight"] / carcass)


def test_constraints_reject_nonfinite_outputs():
    with pytest.raises(ValueError, match="non-finite"):
        production_prediction(100, {"yield_pct": float("nan")})
    with pytest.raises(ValueError, match="non-finite"):
        butchery_prediction(25, dict.fromkeys(PART_COLUMNS, float("inf")))


def test_prediction_cost_independence_and_pdf_accounting():
    service = MLService()
    request = dict(raw_material_kg=100, spice_kg=2, casing_kg=5)
    first = service.predict_yield_v2(YieldV2Request(**request, service_cost=0))
    second = service.predict_yield_v2(YieldV2Request(**request, service_cost=900, labor_cost=700))
    assert first["prediction"]["finished_qty"] == second["prediction"]["finished_qty"]
    assert first == service.predict_yield_v2(YieldV2Request(**request, service_cost=0))
    assert first["prediction"]["finished_qty"] == pytest.approx(100 * first["prediction"]["yield_pct"], abs=.006)
    pdf = service.predict_yield_v2(YieldV2Request(**request, service_unit_cost=5.50, actual_finished_qty=85))
    assert pdf["accounting_cost_breakdown"]["total_cost"] == 1543.50
    assert round(pdf["accounting_ekp"], 2) == 18.16
    assert pdf["comparison"]["absolute_error"] == pytest.approx(abs(pdf["prediction"]["finished_qty"] - 85), abs=.001)
    assert service.predict_yield_v2(YieldV2Request(raw_material_kg=140))["prediction"]["finished_qty"] > service.predict_yield_v2(YieldV2Request(raw_material_kg=70))["prediction"]["finished_qty"]


@pytest.mark.parametrize("animal,carcass,live", [("lamb", 25, 53), ("cattle", 300, 520)])
def test_legacy_carcass_equals_v2_without_live_conversion(animal, carcass, live):
    from ml.src.schemas import LegacyButcheryRequest
    service = MLService()
    modern = service.predict_butchery_v2(ButcheryV2Request(animal_type=animal, carcass_weight_kg=carcass))
    with_live = service.predict_butchery_v2(ButcheryV2Request(animal_type=animal, carcass_weight_kg=carcass, live_weight_kg=live))
    legacy = service.predict_butchery_legacy(LegacyButcheryRequest(animal_type=animal, input_weight=carcass))
    assert modern["prediction"] == with_live["prediction"] == legacy["prediction"]
    assert modern == service.predict_butchery_v2(ButcheryV2Request(animal_type=animal, carcass_weight_kg=carcass))
    assert modern["physical_check"]["parts_sum_kg"] == carcass
    assert modern["physical_check"]["saleable_sum_kg"] <= carcass


@pytest.mark.parametrize("payload", [{}, {"raw_material_kg": 0}, {"raw_material_kg": -1},
    {"raw_material_kg": float("nan")}, {"raw_material_kg": 100, "labor_cost": float("inf")},
    {"raw_material_kg": 100, "spice_kg": None}])
def test_missing_negative_and_nonfinite_inputs(payload):
    with pytest.raises(ValueError):
        YieldV2Request(**payload)


def test_live_only_request_requires_measured_carcass():
    with pytest.raises(ValueError):
        ButcheryV2Request(animal_type="lamb", live_weight_kg=45)


def test_all_recorded_splits_have_disjoint_groups_and_ids():
    from ml.src.data import butchery_frame
    for name, frame in [("yield", production_frame()), ("butchery_lamb", butchery_frame("lamb")), ("butchery_cattle", butchery_frame("cattle"))]:
        split = json.loads((ARTIFACTS_DIR / f"{name}_model" / "split.json").read_text())
        train, test = frame.iloc[split["train_indices"]], frame.iloc[split["test_indices"]]
        group = split["group_column"]
        identity = "order_id" if name == "yield" else "animal_id"
        assert not set(train[group]) & set(test[group])
        assert not set(train[identity]) & set(test[identity])


def test_post_production_variables_excluded_and_bad_row_rejected():
    assert not set(PRODUCTION_FEATURES) & {"actual_ekp", "actual_total_cost", "finished_qty", "yield_pct", "expected_qty", "service_cost", "labor_cost"}
    frame = production_frame().copy()
    frame.loc[0, "finished_qty"] = -1
    with pytest.raises(ValueError):
        validate_production(frame)


def test_changing_outcome_noise_cannot_change_any_planned_input():
    from ml.src.synthetic import generate_production
    first = generate_production(n_orders=80, seed=15, outcome_noise_scale=0)
    changed = generate_production(n_orders=80, seed=15, outcome_noise_scale=.03)
    pd.testing.assert_frame_equal(first[PRODUCTION_FEATURES + ["service_cost", "labor_cost", "expected_qty"]],
                                  changed[PRODUCTION_FEATURES + ["service_cost", "labor_cost", "expected_qty"]])
    assert not first["finished_qty"].equals(changed["finished_qty"])


def test_test_labels_never_affect_selection_or_preprocessor():
    x = pd.DataFrame({"x": np.arange(100, dtype=float)})
    groups = pd.Series(np.repeat(np.arange(10), 10))
    y = pd.Series(2 * x.x + 1)
    kwargs = dict(X_train=x.iloc[:80], y_train=y.iloc[:80], X_test=x.iloc[80:], groups=groups.iloc[:80], quick=True)
    first = train_target(**kwargs, y_test=y.iloc[80:])
    changed = train_target(**kwargs, y_test=pd.Series(np.repeat(-10000, 20)))
    assert first["selected_model"] == changed["selected_model"]
    assert first["selected"]["cv_mae_mean"] == changed["selected"]["cv_mae_mean"]
    scale = first["pipeline"].named_steps["prep"].named_transformers_["num"].named_steps["scale"]
    assert scale.mean_[0] == pytest.approx(x.iloc[:80].x.mean())
