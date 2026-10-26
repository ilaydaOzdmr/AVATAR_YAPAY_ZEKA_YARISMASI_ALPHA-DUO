import json
from pathlib import Path

import numpy as np
import pandas as pd

if __package__:
    from .model import baseline_mean, evaluate, predict, train_yield
else:
    from model import baseline_mean, evaluate, predict, train_yield


MATERIAL_UNIT_COST = 9.70
PRODUCT_CODES = [
    "PRD-101",
    "PRD-102",
    "PRD-103",
    "PRD-104",
    "PRD-105",
    "PRD-106",
    "PRD-107",
]
ANIMAL_TYPE_CODES = {"kuzu": 0, "dana": 1}

# Small, anonymized sample data. unit_cost is recorded for EKP consistency;
# the yield model predicts finished_qty and EKP is derived from that prediction.
PRODUCTION_SAMPLE = [
    {"raw_material_kg": 80.0, "service_cost": 390.0, "finished_qty": 68.0},
    {"raw_material_kg": 90.0, "service_cost": 420.0, "finished_qty": 76.0},
    {"raw_material_kg": 100.0, "service_cost": 467.5, "finished_qty": 85.0},
    {"raw_material_kg": 110.0, "service_cost": 510.0, "finished_qty": 93.0},
    {"raw_material_kg": 120.0, "service_cost": 545.0, "finished_qty": 101.0},
    {"raw_material_kg": 130.0, "service_cost": 590.0, "finished_qty": 110.0},
    {"raw_material_kg": 140.0, "service_cost": 625.0, "finished_qty": 118.0},
]

BUTCHERY_INPUT_WEIGHTS = {
    "kuzu": [20.0, 22.0, 25.0, 27.0],
    "dana": [35.0, 40.0],
}
BUTCHERY_RATIOS = {
    "kuzu": [0.18, 0.12, 0.10, 0.06, 0.05, 0.02, 0.02],
    "dana": [0.20, 0.14, 0.11, 0.06, 0.04, 0.03, 0.03],
}

REPORT_NOTE = "Veri seti örnek/anonymize ve küçüktür; sonuçlar gösterim amaçlıdır."
REPORT_PATH = Path(__file__).with_name("report.json")


def build_production_dataframe():
    frame = pd.DataFrame(PRODUCTION_SAMPLE)
    material_cost = frame["raw_material_kg"] * MATERIAL_UNIT_COST
    frame["unit_cost"] = ((material_cost + frame["service_cost"]) / frame["finished_qty"]).round(2)
    return frame[["raw_material_kg", "service_cost", "finished_qty", "unit_cost"]]


def build_butchery_dataframe():
    records = []
    for animal_type, input_weights in BUTCHERY_INPUT_WEIGHTS.items():
        ratios = BUTCHERY_RATIOS[animal_type]
        for input_weight in input_weights:
            record = {
                "animal_type": animal_type,
                "animal_type_code": ANIMAL_TYPE_CODES[animal_type],
                "input_weight": input_weight,
            }
            for product_code, ratio in zip(PRODUCT_CODES, ratios):
                record[product_code] = round(input_weight * ratio, 3)
            records.append(record)
    return pd.DataFrame(records)


def encode_animal_type(animal_type):
    key = animal_type.strip().lower()
    if key not in ANIMAL_TYPE_CODES:
        raise ValueError(f"unknown animal type: {animal_type}")
    return ANIMAL_TYPE_CODES[key]


def train_models():
    production = build_production_dataframe()
    butchery = build_butchery_dataframe()

    yield_features = production[["raw_material_kg", "service_cost"]].to_numpy()
    yield_target = production["finished_qty"].to_numpy()

    butchery_features = butchery[["animal_type_code", "input_weight"]].to_numpy()
    butchery_target = butchery[PRODUCT_CODES].to_numpy()

    return {
        "yield": train_yield(yield_features, yield_target),
        "butchery": train_yield(butchery_features, butchery_target),
    }


def _rounded(metrics):
    return {name: round(value, 4) for name, value in metrics.items()}


def build_report():
    production = build_production_dataframe()
    butchery = build_butchery_dataframe()

    yield_features = production[["raw_material_kg", "service_cost"]].to_numpy()
    yield_target = production["finished_qty"].to_numpy()
    yield_model = train_yield(yield_features, yield_target)
    yield_prediction = predict(yield_model, yield_features)
    yield_baseline = np.full_like(yield_target, baseline_mean(yield_target), dtype=float)

    butchery_features = butchery[["animal_type_code", "input_weight"]].to_numpy()
    butchery_target = butchery[PRODUCT_CODES].to_numpy()
    butchery_model = train_yield(butchery_features, butchery_target)
    butchery_prediction = predict(butchery_model, butchery_features)
    part_means = np.array(
        [baseline_mean(butchery_target[:, column]) for column in range(butchery_target.shape[1])]
    )
    butchery_baseline = np.tile(part_means, (len(butchery_target), 1))

    yield_model_metrics = _rounded(evaluate(yield_target, yield_prediction))
    yield_baseline_metrics = _rounded(evaluate(yield_target, yield_baseline))
    butchery_model_metrics = _rounded(evaluate(butchery_target, butchery_prediction))
    butchery_baseline_metrics = _rounded(evaluate(butchery_target, butchery_baseline))

    return {
        "yield": {
            "model_mae": yield_model_metrics["mae"],
            "model_rmse": yield_model_metrics["rmse"],
            "baseline_mae": yield_baseline_metrics["mae"],
            "baseline_rmse": yield_baseline_metrics["rmse"],
        },
        "butchery": {
            "model_mae": butchery_model_metrics["mae"],
            "model_rmse": butchery_model_metrics["rmse"],
            "baseline_mae": butchery_baseline_metrics["mae"],
            "baseline_rmse": butchery_baseline_metrics["rmse"],
        },
        "note": REPORT_NOTE,
    }


def write_report(path=REPORT_PATH):
    report = build_report()
    path.write_text(json.dumps(report, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    return report


if __name__ == "__main__":
    write_report()
