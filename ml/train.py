"""Train and compare models for production yield and butchery cut weights.

Usage (from the repository root):

    python ml/train.py
    python ml/train.py --quick

Artifacts are written to ``ml/artifacts``; reports to ``ml/reports``.
No real business data is used: processed datasets are ``synthetic_calibrated``.
"""

from __future__ import annotations

import argparse
import json
import sys
import warnings
from pathlib import Path

warnings.filterwarnings("ignore", message=".*sklearn.utils.parallel.delayed.*")

REPO_ROOT = Path(__file__).resolve().parents[1]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from ml.src.config import (  # noqa: E402
    ARTIFACTS_DIR,
    BUTCHERY_FEATURES,
    BUTCHERY_GROUP,
    BUTCHERY_TARGETS,
    MODEL_STATUS,
    MODEL_VERSION,
    PLOTS_DIR,
    PRODUCTION_FEATURES,
    PRODUCTION_GROUP,
    PRODUCTION_TARGETS,
    REPORTS_DIR,
)
from ml.src.data import butchery_frame, production_frame  # noqa: E402
from ml.src.models import train_group  # noqa: E402
from ml.src.plots import (  # noqa: E402
    save_actual_vs_predicted,
    save_feature_importance,
    save_model_comparison,
    save_residuals,
    try_shap_summary,
)

PRIMARY_TARGET = {
    "yield": "yield_pct",
    "butchery_lamb": "total_saleable_weight",
    "butchery_cattle": "total_saleable_weight",
}


def _comparison_markdown(results: dict[str, dict]) -> str:
    lines = [
        "# Model comparison",
        "",
        "All models were selected by grouped cross-validation MAE. The test set was "
        "used only for reporting, never for selection or hyperparameter tuning.",
        "",
        "> Data origin: `synthetic_calibrated`. These are not real field results.",
        "",
    ]
    for group, payload in results.items():
        lines.append(f"## {group}")
        lines.append("")
        for target, report in payload["metadata"]["targets"].items():
            lines.append(f"### {target}")
            lines.append("")
            lines.append("| Model | CV MAE | CV std | Test MAE | Test RMSE | Test R² | MedAE | MAPE % |")
            lines.append("|---|---:|---:|---:|---:|---:|---:|---:|")
            rows = sorted(
                [row for row in report["candidates"] if row.get("ok")],
                key=lambda row: row["cv_mae_mean"],
            )
            for row in rows:
                marker = " **selected**" if row["model"] == report["selected_model"] else ""
                lines.append(
                    f"| {row['model']}{marker} | {row['cv_mae_mean']:.4f} | "
                    f"{row['cv_mae_std']:.4f} | {row['test_mae']:.4f} | "
                    f"{row['test_rmse']:.4f} | {row['test_r2']:.4f} | "
                    f"{row['test_medae']:.4f} | {row['test_mape']:.2f} |"
                )
            lines.append("")
            lines.append(
                f"Selected: **{report['selected_model']}**; beats mean baseline: "
                f"**{report['beats_baseline']}**."
            )
            lines.append("")
    return "\n".join(lines)


def _summary_markdown(results: dict[str, dict]) -> str:
    lines = [
        "# Model report",
        "",
        f"Model version: `{MODEL_VERSION}` · Status: `{MODEL_STATUS}`",
        "",
        "## What was trained",
        "",
        "| Group | Targets | Train rows | Test rows | Validation |",
        "|---|---|---:|---:|---|",
    ]
    for group, payload in results.items():
        metadata = payload["metadata"]
        lines.append(
            f"| {group} | {len(metadata['target_names'])} | "
            f"{metadata['training_sample_count']} | {metadata['test_sample_count']} | "
            f"{metadata['validation_method']} |"
        )
    lines.extend(
        [
            "",
            "## Selected models",
            "",
            "| Group | Target | Model | CV MAE | Test MAE | Test R² | Beats baseline |",
            "|---|---|---|---:|---:|---:|---|",
        ]
    )
    for group, payload in results.items():
        for target, report in payload["metadata"]["targets"].items():
            selected = report["selected"]
            lines.append(
                f"| {group} | {target} | {report['selected_model']} | "
                f"{selected['cv_mae_mean']:.4f} | {selected['test_mae']:.4f} | "
                f"{selected['test_r2']:.4f} | {report['beats_baseline']} |"
            )
    lines.extend(["", "## Deployed predictions (physical constraints applied)", "",
        "CV selects raw regressions. API quantities use reconciled parts and meat × predicted yield.", "",
        "| Group | Target | Train MAE | Test MAE | Test RMSE | R² | MAPE % | Proportional baseline MAE |", "|---|---|---:|---:|---:|---:|---:|---:|"])
    for group, payload in results.items():
        for target, report in payload["metadata"]["deployed_metrics"].items():
            test = report["test"]
            lines.append(f"| {group} | {target} | {report['train']['mae']} | {test['mae']} | {test['rmse']} | {test['r2']} | {test['mape']} | {report['proportional_baseline_test']['mae']} |")
    lines.extend(
        [
            "",
            "## Limitations",
            "",
            "- Training data is `synthetic_calibrated`, not real production or carcass records.",
            "- These metrics show that the pipeline learns the generator, not that it "
            "generalises to a real plant.",
            "- Production EKP is a separate accounting identity and is not trained as a target.",
            "- Real validation is blocked until real order/carcass data is provided.",
            "- Real reference measurements exist (docs/DATA_SOURCES.md) but lack competition targets.",
            "- No standalone validation partition: GroupKFold validates within train; fold counts in metadata.",
            "",
        ]
    )
    return "\n".join(lines)


def run(quick: bool = False) -> dict[str, dict]:
    REPORTS_DIR.mkdir(parents=True, exist_ok=True)
    PLOTS_DIR.mkdir(parents=True, exist_ok=True)
    results: dict[str, dict] = {}

    production = production_frame()
    results["yield"] = train_group(
        dataset_name="yield",
        frame=production,
        features=PRODUCTION_FEATURES,
        targets=PRODUCTION_TARGETS,
        group_column=PRODUCTION_GROUP,
        artifact_dir=ARTIFACTS_DIR / "yield_model",
        report_dir=REPORTS_DIR,
        quick=quick,
    )

    for animal_type in ("lamb", "cattle"):
        frame = butchery_frame(animal_type)
        name = f"butchery_{animal_type}"
        results[name] = train_group(
            dataset_name=name,
            frame=frame,
            features=BUTCHERY_FEATURES,
            targets=BUTCHERY_TARGETS,
            group_column=BUTCHERY_GROUP,
            artifact_dir=ARTIFACTS_DIR / f"{name}_model",
            report_dir=REPORTS_DIR,
            quick=quick,
        )

    comparison = {}
    for group, payload in results.items():
        comparison[group] = {"metadata": payload["metadata"]}

    (REPORTS_DIR / "model_comparison.json").write_text(
        json.dumps(comparison, indent=2, ensure_ascii=False) + "\n", encoding="utf-8"
    )
    (REPORTS_DIR / "model_comparison.md").write_text(
        _comparison_markdown(results) + "\n", encoding="utf-8"
    )
    (REPORTS_DIR / "model_report.md").write_text(
        _summary_markdown(results) + "\n", encoding="utf-8"
    )

    # Diagnostic plots for the primary target of each group.
    for group, payload in results.items():
        target = PRIMARY_TARGET[group]
        report = payload["metadata"]["targets"][target]
        pipeline = payload["pipelines"][target]
        frame = (
            production
            if group == "yield"
            else butchery_frame(group.removeprefix("butchery_"))
        )
        test_idx = payload["test_indices"]
        X_test = payload["X_test"]
        y_test = frame[target].astype(float).iloc[test_idx]
        prediction = pipeline.predict(X_test)
        title = f"{group} — {target} ({report['selected_model']})"
        save_actual_vs_predicted(
            y_test, prediction, title, PLOTS_DIR / f"{group}_{target}_actual_vs_predicted.png"
        )
        save_residuals(
            y_test, prediction, title, PLOTS_DIR / f"{group}_{target}_residuals.png"
        )
        save_model_comparison(
            report["candidates"], target, PLOTS_DIR / f"{group}_{target}_model_comparison.png"
        )
        save_feature_importance(
            pipeline,
            X_test,
            y_test,
            title,
            PLOTS_DIR / f"{group}_{target}_feature_importance.png",
        )
        try_shap_summary(
            pipeline,
            X_test,
            title,
            PLOTS_DIR / f"{group}_{target}_shap.png",
        )

    return results


def main() -> None:
    parser = argparse.ArgumentParser(description="Train AVATAR ML models")
    parser.add_argument("--quick", action="store_true", help="four candidates; retains grouped cross-validation")
    args = parser.parse_args()
    results = run(quick=args.quick)
    for group, payload in results.items():
        metadata = payload["metadata"]
        print(
            f"{group}: {metadata['training_sample_count']} train / "
            f"{metadata['test_sample_count']} test; artifacts={payload['artifacts']}"
        )
    print(f"reports: {REPORTS_DIR}")


if __name__ == "__main__":
    main()
