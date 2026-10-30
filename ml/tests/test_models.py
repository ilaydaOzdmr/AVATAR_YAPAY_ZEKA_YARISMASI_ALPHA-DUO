import numpy as np
import pandas as pd
import pytest

from ml.src.config import ARTIFACTS_DIR
from ml.src.metrics import regression_metrics
from ml.src.models import train_target
from ml.src.registry import load_bundle


def test_train_target_beats_dummy_on_linear_signal():
    rng = np.random.default_rng(3)
    n = 120
    x = rng.uniform(10, 100, n)
    category = rng.choice(["a", "b"], n)
    y = 2.5 * x + np.where(category == "a", 5.0, -5.0) + rng.normal(0, 1.0, n)
    frame = pd.DataFrame({"x": x, "category": category, "y": y})
    groups = pd.Series([f"g{i % 10}" for i in range(n)])
    split = 90

    result = train_target(
        frame[["x", "category"]].iloc[:split],
        frame["y"].iloc[:split],
        frame[["x", "category"]].iloc[split:],
        frame["y"].iloc[split:],
        groups.iloc[:split],
        quick=True,
    )
    assert result["beats_baseline"] is True
    assert result["selected_model"] != "dummy_mean"
    assert result["selected"]["test_r2"] > 0.9


def test_regression_metrics_handles_zero_targets():
    metrics = regression_metrics([0, 1, 2], [0, 1, 2])
    assert metrics["mae"] == 0.0
    assert metrics["rmse"] == 0.0
    assert metrics["r2"] == 1.0
    # MAPE ignores zero-valued actuals instead of dividing by zero.
    assert metrics["mape"] == pytest.approx(0.0)


def test_artifacts_load_and_predict():
    bundle = load_bundle(ARTIFACTS_DIR / "yield_model", "yield")
    if bundle is None:
        pytest.skip("train artifacts not built yet")
    assert bundle.status == "synthetic_only"
    assert "yield_pct" in bundle.targets

    from ml.src.data import production_frame

    row = production_frame().iloc[[0]]
    prediction = bundle.predict("yield_pct", row[bundle.metadata["feature_names"]])
    assert len(prediction) == 1
    assert 0.5 < float(prediction[0]) < 1.0
