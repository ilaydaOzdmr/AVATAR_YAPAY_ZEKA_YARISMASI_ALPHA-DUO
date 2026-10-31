"""Evaluation plots. All functions are optional at runtime and never fatal."""

from __future__ import annotations

from pathlib import Path

import matplotlib

matplotlib.use("Agg")  # noqa: E402
import matplotlib.pyplot as plt  # noqa: E402
import numpy as np  # noqa: E402
import pandas as pd  # noqa: E402
from sklearn.inspection import permutation_importance  # noqa: E402


def _ensure(path: Path) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)


def save_actual_vs_predicted(y_true, y_pred, title: str, path: Path) -> None:
    _ensure(path)
    fig, ax = plt.subplots(figsize=(6, 5))
    ax.scatter(y_true, y_pred, s=14, alpha=0.65, edgecolors="none")
    low = float(min(np.min(y_true), np.min(y_pred)))
    high = float(max(np.max(y_true), np.max(y_pred)))
    ax.plot([low, high], [low, high], color="black", linewidth=1, linestyle="--")
    ax.set_xlabel("Actual")
    ax.set_ylabel("Predicted")
    ax.set_title(title)
    fig.tight_layout()
    fig.savefig(path, dpi=130)
    plt.close(fig)


def save_residuals(y_true, y_pred, title: str, path: Path) -> None:
    _ensure(path)
    residuals = np.asarray(y_pred) - np.asarray(y_true)
    fig, ax = plt.subplots(figsize=(6, 4))
    ax.scatter(y_pred, residuals, s=14, alpha=0.65, edgecolors="none")
    ax.axhline(0, color="black", linewidth=1, linestyle="--")
    ax.set_xlabel("Predicted")
    ax.set_ylabel("Residual (predicted - actual)")
    ax.set_title(title)
    fig.tight_layout()
    fig.savefig(path, dpi=130)
    plt.close(fig)


def save_model_comparison(rows: list[dict], target: str, path: Path) -> None:
    _ensure(path)
    valid = [row for row in rows if row.get("ok")]
    names = [row["model"] for row in valid]
    values = [row["cv_mae_mean"] for row in valid]
    fig, ax = plt.subplots(figsize=(8, 4))
    ax.bar(names, values, color="#2C7058")
    ax.set_ylabel("CV MAE (lower is better)")
    ax.set_title(f"Candidate comparison — {target}")
    ax.tick_params(axis="x", rotation=25)
    fig.tight_layout()
    fig.savefig(path, dpi=130)
    plt.close(fig)


def save_feature_importance(
    model,
    x_test: pd.DataFrame,
    y_test: pd.Series,
    title: str,
    path: Path,
    top_n: int = 15,
) -> None:
    _ensure(path)
    result = permutation_importance(
        model,
        x_test,
        y_test,
        scoring="neg_mean_absolute_error",
        n_repeats=5,
        random_state=42,
        n_jobs=-1,
    )
    order = np.argsort(result.importances_mean)[::-1][:top_n]
    names = [x_test.columns[i] for i in order]
    values = result.importances_mean[order]
    fig, ax = plt.subplots(figsize=(8, max(3, 0.35 * len(names))))
    ax.barh(names[::-1], values[::-1], color="#B97722")
    ax.set_xlabel("Mean decrease in MAE (permutation importance)")
    ax.set_title(title)
    fig.tight_layout()
    fig.savefig(path, dpi=130)
    plt.close(fig)


def try_shap_summary(
    model,
    x_test: pd.DataFrame,
    title: str,
    path: Path,
    max_samples: int = 200,
) -> bool:
    """Best-effort SHAP summary; returns False when SHAP cannot explain the model."""
    try:
        import shap
    except ImportError:
        return False

    try:
        sample = x_test.head(max_samples)
        preprocessor = model.named_steps["prep"]
        estimator = model.named_steps["model"]
        transformed = preprocessor.transform(sample)
        names = preprocessor.get_feature_names_out()
        if hasattr(shap, "TreeExplainer") and estimator.__class__.__name__ in {
            "RandomForestRegressor",
            "ExtraTreesRegressor",
            "HistGradientBoostingRegressor",
            "XGBRegressor",
            "CatBoostRegressor",
        }:
            explainer = shap.TreeExplainer(estimator)
            values = explainer.shap_values(transformed)
        else:
            explainer = shap.LinearExplainer(estimator, transformed)
            values = explainer.shap_values(transformed)
        plt.figure(figsize=(8, 5))
        shap.summary_plot(values, transformed, feature_names=list(names), show=False)
        plt.title(title)
        _ensure(path)
        plt.tight_layout()
        plt.savefig(path, dpi=130, bbox_inches="tight")
        plt.close()
        return True
    except Exception:  # noqa: BLE001 - SHAP is diagnostic, never fatal
        plt.close("all")
        return False
