import json
from pathlib import Path

import numpy as np
import pandas as pd

if __package__:
    from .model import baseline_mean, evaluate, predict, train_yield
else:
    from model import baseline_mean, evaluate, predict, train_yield


# --- Realistic domain constants -------------------------------------------------
# Sucuk (fermente sosis) üretimi: et maliyeti + işçilik/hizmet; ~%85 verim.
MATERIAL_UNIT_COST = 9.70
SERVICE_PER_PACKAGE = 5.50
YIELD_RATIO_MEAN = 0.85
YIELD_RATIO_STD = 0.015

PRODUCT_CODES = [
    "PRD-101",  # but
    "PRD-102",  # sırt
    "PRD-103",  # omuz
    "PRD-104",  # kaburga
    "PRD-105",  # kıyma
    "PRD-106",  # kuşbaşı
    "PRD-107",  # kemik / yan ürün
]
ANIMAL_TYPE_CODES = {"kuzu": 0, "dana": 1}

# Gerçekçi parça oranları (karkas ağırlığının payı; toplam ~0.80, kalan fire).
BUTCHERY_RATIOS = {
    "kuzu": [0.18, 0.10, 0.15, 0.08, 0.15, 0.06, 0.08],
    "dana": [0.20, 0.12, 0.14, 0.09, 0.13, 0.07, 0.05],
}
BUTCHERY_INPUT_RANGES = {"kuzu": (18.0, 30.0), "dana": (35.0, 55.0)}

N_SAMPLES = 200
TEST_FRACTION = 0.2
SEED = 42

REPORT_NOTE = (
    "Veri seti gerçekçi üretim/parçalama oranlarından üretilmiş 200'er örnekten oluşur "
    "(anonymize); metrikler ayrı test kümesi (%20) üzerinden raporlanır."
)
REPORT_PATH = Path(__file__).with_name("report.json")

RNG = np.random.default_rng(SEED)


def build_production_dataframe():
    raw_material_kg = RNG.uniform(70.0, 150.0, N_SAMPLES)
    yield_ratio = RNG.normal(YIELD_RATIO_MEAN, YIELD_RATIO_STD, N_SAMPLES)
    finished_qty = np.round(raw_material_kg * yield_ratio, 1)
    service_cost = np.round(finished_qty * SERVICE_PER_PACKAGE, 2)
    material_cost = raw_material_kg * MATERIAL_UNIT_COST
    unit_cost = np.round((material_cost + service_cost) / finished_qty, 2)

    frame = pd.DataFrame(
        {
            "raw_material_kg": np.round(raw_material_kg, 1),
            "service_cost": service_cost,
            "finished_qty": finished_qty,
            "unit_cost": unit_cost,
        }
    )
    return frame[["raw_material_kg", "service_cost", "finished_qty", "unit_cost"]]


def build_butchery_dataframe():
    records = []
    for animal_type, (lo, hi) in BUTCHERY_INPUT_RANGES.items():
        ratios = BUTCHERY_RATIOS[animal_type]
        input_weights = RNG.uniform(lo, hi, N_SAMPLES // 2)
        for input_weight in input_weights:
            record = {
                "animal_type": animal_type,
                "animal_type_code": ANIMAL_TYPE_CODES[animal_type],
                "input_weight": round(float(input_weight), 1),
            }
            for product_code, ratio in zip(PRODUCT_CODES, ratios):
                weight = input_weight * ratio + RNG.normal(0.0, 0.08)
                record[product_code] = round(float(max(weight, 0.0)), 3)
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

    yield_features = production[["raw_material_kg"]].to_numpy()
    yield_target = production["finished_qty"].to_numpy()

    butchery_features = butchery[["animal_type_code", "input_weight"]].to_numpy()
    butchery_target = butchery[PRODUCT_CODES].to_numpy()

    return {
        "yield": train_yield(yield_features, yield_target),
        "butchery": train_yield(butchery_features, butchery_target),
    }


def _train_test_split(features, target):
    n = len(features)
    indices = RNG.permutation(n)
    split = int(n * (1 - TEST_FRACTION))
    return features[indices[:split]], features[indices[split:]], target[indices[:split]], target[indices[split:]]


def _rounded(metrics):
    return {name: round(value, 4) for name, value in metrics.items()}


def _evaluate_model(features, target):
    X_train, X_test, y_train, y_test = _train_test_split(features, target)
    model = train_yield(X_train, y_train)
    prediction = predict(model, X_test)
    baseline = np.tile(np.mean(y_train, axis=0), (len(y_test), 1))
    return _rounded(evaluate(y_test, prediction)), _rounded(evaluate(y_test, baseline))


def build_report():
    production = build_production_dataframe()
    butchery = build_butchery_dataframe()

    yield_model_metrics, yield_baseline_metrics = _evaluate_model(
        production[["raw_material_kg"]].to_numpy(),
        production["finished_qty"].to_numpy(),
    )
    butchery_model_metrics, butchery_baseline_metrics = _evaluate_model(
        butchery[["animal_type_code", "input_weight"]].to_numpy(),
        butchery[PRODUCT_CODES].to_numpy(),
    )

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
