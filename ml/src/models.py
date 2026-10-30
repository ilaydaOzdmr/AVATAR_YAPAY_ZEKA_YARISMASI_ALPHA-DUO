"""Model candidates, grouped validation and artifact-oriented training."""

from __future__ import annotations

import json
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

import joblib
import numpy as np
import pandas as pd
from sklearn.base import clone
from sklearn.compose import ColumnTransformer
from sklearn.dummy import DummyRegressor
from sklearn.ensemble import (
    ExtraTreesRegressor,
    HistGradientBoostingRegressor,
    RandomForestRegressor,
)
from sklearn.impute import SimpleImputer
from sklearn.linear_model import Lasso, LinearRegression, Ridge
from sklearn.model_selection import GroupKFold, GroupShuffleSplit, RandomizedSearchCV, cross_val_score
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import OneHotEncoder, StandardScaler

from .config import CV_FOLDS, MODEL_STATUS, MODEL_VERSION, RANDOM_STATE, TEST_SIZE
from .data import dataset_version
from .metrics import regression_metrics, rounded


@dataclass
class Candidate:
    name: str
    estimator: Any
    params: dict[str, list[Any]] = field(default_factory=dict)


def _optional_candidates() -> list[Candidate]:
    candidates: list[Candidate] = []
    try:  # pragma: no cover - optional dependency
        from xgboost import XGBRegressor

        candidates.append(
            Candidate(
                "xgboost",
                XGBRegressor(
                    objective="reg:squarederror",
                    random_state=RANDOM_STATE,
                    n_jobs=-1,
                    verbosity=0,
                ),
                {
                    "model__n_estimators": [100, 200],
                    "model__max_depth": [3, 6],
                    "model__learning_rate": [0.05, 0.1],
                },
            )
        )
    except ImportError:
        pass
    try:  # pragma: no cover - optional dependency
        from catboost import CatBoostRegressor

        candidates.append(
            Candidate(
                "catboost",
                CatBoostRegressor(random_seed=RANDOM_STATE, verbose=False, allow_writing_files=False),
                {
                    "model__iterations": [200, 400],
                    "model__depth": [4, 6],
                    "model__learning_rate": [0.05, 0.1],
                },
            )
        )
    except ImportError:
        pass
    return candidates


def candidate_models(quick: bool = False) -> list[Candidate]:
    candidates = [
        Candidate("dummy_mean", DummyRegressor(strategy="mean")),
        Candidate("linear", LinearRegression()),
        Candidate(
            "ridge",
            Ridge(random_state=RANDOM_STATE),
            {"model__alpha": [0.1, 1.0, 10.0, 100.0]},
        ),
        Candidate(
            "lasso",
            Lasso(max_iter=20000, random_state=RANDOM_STATE),
            {"model__alpha": [0.001, 0.01, 0.1, 1.0]},
        ),
        Candidate(
            "random_forest",
            RandomForestRegressor(random_state=RANDOM_STATE, n_jobs=-1),
            {
                "model__n_estimators": [80, 160],
                "model__max_depth": [None, 8, 16],
                "model__min_samples_leaf": [1, 3, 5],
            },
        ),
        Candidate(
            "extra_trees",
            ExtraTreesRegressor(random_state=RANDOM_STATE, n_jobs=-1),
            {
                "model__n_estimators": [80, 160],
                "model__max_depth": [None, 8, 16],
                "model__min_samples_leaf": [1, 3, 5],
            },
        ),
        Candidate(
            "hist_gradient_boosting",
            HistGradientBoostingRegressor(random_state=RANDOM_STATE, early_stopping=False),
            {
                "model__learning_rate": [0.05, 0.1],
                "model__max_iter": [120, 240],
                "model__max_leaf_nodes": [15, 31],
            },
        ),
    ]
    candidates.extend(_optional_candidates())
    if quick:
        keep = {"dummy_mean", "linear", "ridge", "random_forest"}
        candidates = [candidate for candidate in candidates if candidate.name in keep]
    return candidates


def make_preprocessor(features: pd.DataFrame) -> ColumnTransformer:
    numeric = features.select_dtypes(include=[np.number]).columns.tolist()
    categorical = [column for column in features.columns if column not in numeric]
    transformers: list[tuple[str, Any, list[str]]] = [
        (
            "num",
            Pipeline(
                [
                    ("impute", SimpleImputer(strategy="median")),
                    ("scale", StandardScaler()),
                ]
            ),
            numeric,
        )
    ]
    if categorical:
        transformers.append(
            (
                "cat",
                Pipeline(
                    [
                        ("impute", SimpleImputer(strategy="most_frequent")),
                        (
                            "encode",
                            OneHotEncoder(handle_unknown="ignore", sparse_output=False),
                        ),
                    ]
                ),
                categorical,
            )
        )
    return ColumnTransformer(transformers, remainder="drop")


def make_pipeline(candidate: Candidate, features: pd.DataFrame) -> Pipeline:
    return Pipeline(
        [
            ("prep", make_preprocessor(features)),
            ("model", clone(candidate.estimator)),
        ]
    )


def _cv_splits(groups: pd.Series) -> int:
    unique = groups.nunique()
    return max(2, min(CV_FOLDS, int(unique)))


def train_target(
    X_train: pd.DataFrame,
    y_train: pd.Series,
    X_test: pd.DataFrame,
    y_test: pd.Series,
    groups: pd.Series,
    *,
    quick: bool = False,
) -> dict[str, Any]:
    """Compare candidates with grouped CV, then fit the CV winner on train."""
    cv = GroupKFold(n_splits=_cv_splits(groups))
    rows: list[dict[str, Any]] = []
    fitted: dict[str, Pipeline] = {}

    for candidate in candidate_models(quick=quick):
        pipeline = make_pipeline(candidate, X_train)
        try:
            if candidate.params:
                search = RandomizedSearchCV(
                    pipeline,
                    candidate.params,
                    n_iter=min(4, _space_size(candidate.params)),
                    cv=cv,
                    scoring="neg_mean_absolute_error",
                    random_state=RANDOM_STATE,
                    n_jobs=-1,
                    refit=True,
                )
                search.fit(X_train, y_train, groups=groups)
                best_index = int(search.best_index_)
                cv_mean = -float(search.cv_results_["mean_test_score"][best_index])
                cv_std = float(search.cv_results_["std_test_score"][best_index])
                model = search.best_estimator_
            else:
                scores = cross_val_score(
                    pipeline,
                    X_train,
                    y_train,
                    groups=groups,
                    cv=cv,
                    scoring="neg_mean_absolute_error",
                    n_jobs=-1,
                )
                cv_mean = float(-np.mean(scores))
                cv_std = float(np.std(scores))
                model = pipeline.fit(X_train, y_train)
        except Exception as exc:  # noqa: BLE001 - report failed candidates
            rows.append(
                {
                    "model": candidate.name,
                    "ok": False,
                    "error": f"{type(exc).__name__}: {exc}",
                }
            )
            continue

        prediction = model.predict(X_test)
        test_metrics = rounded(regression_metrics(y_test, prediction))
        fitted[candidate.name] = model
        rows.append(
            {
                "model": candidate.name,
                "ok": True,
                "cv_mae_mean": round(cv_mean, 4),
                "cv_mae_std": round(cv_std, 4),
                **{f"test_{key}": value for key, value in test_metrics.items()},
            }
        )

    valid = [row for row in rows if row.get("ok")]
    if not valid:
        raise RuntimeError("all candidate models failed")
    selected = min(valid, key=lambda row: row["cv_mae_mean"])
    baseline = next((row for row in valid if row["model"] == "dummy_mean"), None)
    return {
        "selected_model": selected["model"],
        "selected": selected,
        "candidates": rows,
        "beats_baseline": bool(
            baseline is not None
            and selected["model"] != "dummy_mean"
            and selected["cv_mae_mean"] < baseline["cv_mae_mean"]
        ),
        "pipeline": fitted[selected["model"]],
    }


def _space_size(params: dict[str, list[Any]]) -> int:
    size = 1
    for values in params.values():
        size *= max(1, len(values))
    return size


def train_group(
    *,
    dataset_name: str,
    frame: pd.DataFrame,
    features: list[str],
    targets: list[str],
    group_column: str,
    artifact_dir: Path,
    report_dir: Path,
    quick: bool = False,
) -> dict[str, Any]:
    X = frame[features]
    groups = frame[group_column].astype(str)
    splitter = GroupShuffleSplit(
        n_splits=1, test_size=TEST_SIZE, random_state=RANDOM_STATE
    )
    train_idx, test_idx = next(splitter.split(X, groups=groups))

    X_train, X_test = X.iloc[train_idx], X.iloc[test_idx]
    train_groups = groups.iloc[train_idx]

    pipelines: dict[str, Pipeline] = {}
    target_reports: dict[str, Any] = {}
    for target in targets:
        y = frame[target].astype(float)
        result = train_target(
            X_train,
            y.iloc[train_idx],
            X_test,
            y.iloc[test_idx],
            train_groups,
            quick=quick,
        )
        pipelines[target] = result.pop("pipeline")
        target_reports[target] = result

    artifact_dir.mkdir(parents=True, exist_ok=True)
    report_dir.mkdir(parents=True, exist_ok=True)
    joblib.dump(pipelines, artifact_dir / "model.joblib")

    metadata = {
        "model_version": MODEL_VERSION,
        "training_date": datetime.now(timezone.utc).isoformat(),
        "dataset_name": dataset_name,
        "dataset_version": dataset_version(dataset_name),
        "dataset_source": "synthetic_calibrated (see ml/data/README.md)",
        "data_origin": "synthetic_calibrated",
        "model_status": MODEL_STATUS,
        "feature_names": features,
        "target_names": targets,
        "training_sample_count": int(len(train_idx)),
        "test_sample_count": int(len(test_idx)),
        "validation_method": (
            f"GroupShuffleSplit(test_size={TEST_SIZE}, seed={RANDOM_STATE}) + "
            f"GroupKFold({_cv_splits(train_groups)})"
        ),
        "group_column": group_column,
        "targets": target_reports,
        "warnings": [
            "Trained on calibrated synthetic data, not real business records.",
            "Metrics are not evidence of field performance.",
            "Production EKP is an accounting identity, not a model target.",
        ],
    }
    (artifact_dir / "metadata.json").write_text(
        json.dumps(metadata, indent=2, ensure_ascii=False) + "\n", encoding="utf-8"
    )
    (artifact_dir / "metrics.json").write_text(
        json.dumps(target_reports, indent=2, ensure_ascii=False) + "\n", encoding="utf-8"
    )
    (artifact_dir / "split.json").write_text(
        json.dumps(
            {
                "train_indices": [int(i) for i in train_idx],
                "test_indices": [int(i) for i in test_idx],
                "group_column": group_column,
                "random_state": RANDOM_STATE,
            },
            indent=2,
        )
        + "\n",
        encoding="utf-8",
    )
    return {
        "artifacts": str(artifact_dir),
        "metadata": metadata,
        "pipelines": pipelines,
        "X_test": X_test,
        "test_indices": test_idx,
    }
