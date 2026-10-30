"""Re-evaluate saved artifacts on their recorded test split.

Usage:

    python ml/evaluate.py

This script does not train or tune anything. It loads the exact test indices
saved by ``train.py`` so the reported metrics are reproducible.
"""

from __future__ import annotations

import json
import sys
from pathlib import Path

import joblib

REPO_ROOT = Path(__file__).resolve().parents[1]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from ml.src.config import (  # noqa: E402
    ARTIFACTS_DIR,
    BUTCHERY_FEATURES,
    BUTCHERY_TARGETS,
    PRODUCTION_FEATURES,
    PRODUCTION_TARGETS,
    REPORTS_DIR,
)
from ml.src.data import butchery_frame, production_frame  # noqa: E402
from ml.src.metrics import regression_metrics, rounded  # noqa: E402


def evaluate_group(
    name: str,
    artifact_dir: Path,
    frame,
    features: list[str],
    targets: list[str],
) -> dict:
    metadata = json.loads((artifact_dir / "metadata.json").read_text(encoding="utf-8"))
    split = json.loads((artifact_dir / "split.json").read_text(encoding="utf-8"))
    pipelines = joblib.load(artifact_dir / "model.joblib")
    test_idx = split["test_indices"]
    X_test = frame[features].iloc[test_idx]
    report: dict = {
        "model_version": metadata["model_version"],
        "model_status": metadata["model_status"],
        "dataset_version": metadata["dataset_version"],
        "test_sample_count": len(test_idx),
        "targets": {},
    }
    for target in targets:
        y_test = frame[target].astype(float).iloc[test_idx]
        prediction = pipelines[target].predict(X_test)
        report["targets"][target] = rounded(regression_metrics(y_test, prediction))
    return report


def main() -> None:
    production = production_frame()
    groups = {
        "yield": (
            ARTIFACTS_DIR / "yield_model",
            production,
            PRODUCTION_FEATURES,
            PRODUCTION_TARGETS,
        )
    }
    for animal_type in ("lamb", "cattle"):
        name = f"butchery_{animal_type}"
        groups[name] = (
            ARTIFACTS_DIR / f"{name}_model",
            butchery_frame(animal_type),
            BUTCHERY_FEATURES,
            BUTCHERY_TARGETS,
        )

    report = {}
    for name, (artifact_dir, frame, features, targets) in groups.items():
        if not (artifact_dir / "model.joblib").exists():
            report[name] = {"error": "artifact missing"}
            continue
        report[name] = evaluate_group(name, artifact_dir, frame, features, targets)

    REPORTS_DIR.mkdir(parents=True, exist_ok=True)
    target = REPORTS_DIR / "evaluation_report.json"
    target.write_text(json.dumps(report, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    print(json.dumps(report, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    main()
