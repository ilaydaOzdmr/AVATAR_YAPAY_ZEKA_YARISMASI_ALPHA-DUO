import math

import pandas as pd
import pytest

from src.synthetic import (
    TARGET_PARTS,
    build_all,
    generate_butchery,
    generate_production,
    load_usda_beef_price_index,
)


def test_production_generator_is_consistent():
    frame = generate_production(n_orders=120, n_batches=15, seed=7)
    assert len(frame) == 120
    assert frame["data_origin"].eq("synthetic_calibrated").all()
    assert frame["production_batch"].nunique() == 15
    assert frame[["raw_material_kg", "finished_qty", "yield_pct", "actual_ekp"]].notna().all().all()
    assert frame["yield_pct"].between(0.72, 0.94).all()
    assert frame["raw_material_kg"].between(60, 160).all()
    assert (frame["finished_qty"] > 0).all()
    assert (frame["actual_ekp"] > 0).all()

    # The accounting identity must hold, not be a model prediction.
    expected = frame["actual_total_cost"] / frame["finished_qty"]
    assert (abs(expected - frame["actual_ekp"]) < 0.011).all()


@pytest.mark.parametrize("animal_type", ["lamb", "cattle"])
def test_butchery_physical_constraints(animal_type):
    frame = generate_butchery(animal_type, n_animals=200, n_farms=8, seed=11)
    assert len(frame) == 200
    assert frame["data_origin"].eq("synthetic_calibrated").all()
    assert frame["farm_id"].nunique() == 8
    assert frame["animal_id"].is_unique

    parts = frame[TARGET_PARTS].sum(axis=1)
    assert (abs(parts - frame["carcass_weight_kg"]) < 1e-6).all()
    saleable = frame[
        [
            "leg_weight",
            "loin_weight",
            "shoulder_weight",
            "rib_weight",
            "mince_weight",
            "diced_meat_weight",
        ]
    ].sum(axis=1)
    assert (abs(saleable - frame["total_saleable_weight"]) < 1e-6).all()
    assert (saleable <= frame["carcass_weight_kg"] + 1e-9).all()
    assert frame[TARGET_PARTS].ge(0).all().all()
    assert frame["cutting_yield_pct"].between(40, 90).all()
    assert frame["live_weight_kg"].gt(0).all()
    assert frame["carcass_weight_kg"].gt(0).all()


def test_lamb_and_cattle_are_separate():
    lamb = generate_butchery("lamb", n_animals=50, seed=1)
    cattle = generate_butchery("cattle", n_animals=50, seed=2)
    assert set(lamb["animal_type"]) == {"lamb"}
    assert set(cattle["animal_type"]) == {"cattle"}
    assert lamb["carcass_weight_kg"].mean() < cattle["carcass_weight_kg"].mean()


def test_price_index_fallback(tmp_path):
    missing = tmp_path / "missing.csv"
    index = load_usda_beef_price_index(missing)
    assert len(index) == 1
    assert math.isclose(float(index.iloc[0]), 1.0)


def test_build_all_writes_expected_files(tmp_path):
    counts = build_all(tmp_path, tmp_path / "missing_usda.csv", seed=3)
    assert counts == {
        "production_orders_synthetic": 480,
        "butchery_lamb_synthetic": 800,
        "butchery_cattle_synthetic": 800,
    }
    for name in counts:
        path = tmp_path / f"{name}.csv"
        assert path.exists()
        frame = pd.read_csv(path)
        assert len(frame) == counts[name]
        assert "data_origin" in frame.columns
