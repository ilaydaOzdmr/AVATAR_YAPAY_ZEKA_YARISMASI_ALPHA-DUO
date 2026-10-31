"""Regression checks for deployed identities and train-only baselines."""
import numpy as np
import pandas as pd
import pytest

from ml.src.data import PART_COLUMNS
from ml.src.models import _constrained_predictions
from ml.src import models


class ConstantPrediction:
    def __init__(self, value):
        self.value = value

    def predict(self, frame):
        return np.full(len(frame), self.value)


def test_production_constraints_ignore_inconsistent_finished_target_and_test_labels():
    training = pd.DataFrame({"product_type": ["sucuk", "sucuk"], "yield_pct": [.8, .9]})
    frame = pd.DataFrame({"raw_material_kg": [100., 200.], "product_type": ["sucuk", "unseen"], "yield_pct": [-999., 999.]})
    targets = ["yield_pct", "finished_qty"]
    pipelines = {"yield_pct": ConstantPrediction(.88), "finished_qty": ConstantPrediction(9999)}
    rows, predicted, reference = _constrained_predictions(frame, [0, 1], ["raw_material_kg"], targets, pipelines, training, "yield")
    assert rows.index.tolist() == [0, 1]
    assert predicted.finished_qty.tolist() == [88., 176.]
    assert reference.finished_qty.tolist() == [85., 170.]
    frame["yield_pct"] = 0
    _, _, changed_reference = _constrained_predictions(frame, [0, 1], ["raw_material_kg"], targets, pipelines, training, "yield")
    pd.testing.assert_frame_equal(reference, changed_reference)


def test_butchery_constraints_and_baseline_reconcile_to_measured_carcass():
    frame = pd.DataFrame({"carcass_weight_kg": [20., 40.], **{part: [2.5, 5.] for part in PART_COLUMNS}})
    pipelines = {part: ConstantPrediction(1) for part in PART_COLUMNS}
    _, predicted, baseline = _constrained_predictions(frame, [0, 1], ["carcass_weight_kg"], PART_COLUMNS, pipelines, frame, "butchery_lamb")
    for result in [predicted, baseline]:
        assert result[PART_COLUMNS].sum(axis=1).tolist() == pytest.approx([20., 40.])
        assert (result[PART_COLUMNS] >= 0).all().all()


def test_group_training_keeps_feature_names_separate_from_feature_frame(tmp_path, monkeypatch):
    raw = np.linspace(80, 120, 80)
    frame = pd.DataFrame({"order_id": [f"TEST-{index}" for index in range(80)],
        "data_origin": "synthetic_calibrated", "batch": np.repeat(np.arange(8), 10),
        "raw_material_kg": raw, "product_type": "sucuk", "yield_pct": .85, "finished_qty": raw * .85})
    features = ["raw_material_kg", "product_type"]
    original_candidates = models.candidate_models
    monkeypatch.setattr(models, "candidate_models", lambda quick: original_candidates(quick=True)[:2])
    monkeypatch.setattr(models, "dataset_version", lambda _: "unit-test-fixture")
    result = models.train_group(dataset_name="yield", frame=frame, features=features,
        targets=["yield_pct", "finished_qty"], group_column="batch", artifact_dir=tmp_path / "artifacts",
        report_dir=tmp_path / "reports", quick=True)
    assert result["metadata"]["feature_names"] == features
    assert result["X_test"].columns.tolist() == features
    assert set(result["pipelines"]) == {"yield_pct", "finished_qty"}
    assert result["metadata"]["training_sample_count"] + result["metadata"]["test_sample_count"] == len(frame)
    assert (tmp_path / "artifacts/model.joblib").exists()
