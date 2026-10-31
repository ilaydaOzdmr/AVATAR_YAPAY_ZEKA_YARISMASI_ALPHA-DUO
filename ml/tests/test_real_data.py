from pathlib import Path
import zipfile

import pandas as pd
import pytest

from ml.scripts.import_real_data import RAW, SOURCES, clean_nwfp, download, run, suitability
from ml.src.config import DATASET_FILES


def test_reference_with_incomplete_targets_cannot_enter_training():
    result = suitability(pd.DataFrame({"live_weight_kg": [100], "hot_carcass_weight_kg": [50]}))
    assert all(not entry["eligible_for_training"] for entry in result.values())
    assert "leg_weight" in result["butchery"]["missing_targets"]
    assert all("synthetic_v2" in str(path) for path in DATASET_FILES.values())


def test_trainer_rejects_real_reference_even_when_called_directly(tmp_path):
    from ml.src.models import train_group
    with pytest.raises(ValueError, match="only accepts explicit synthetic"):
        train_group(dataset_name="yield", frame=pd.DataFrame({"data_origin": ["real_reference"]}),
                    features=[], targets=[], group_column="batch", artifact_dir=tmp_path / "artifacts",
                    report_dir=tmp_path / "reports", quick=True)
    assert not (tmp_path / "artifacts").exists()


def test_import_refuses_unverified_file(tmp_path, monkeypatch):
    monkeypatch.setattr("ml.scripts.import_real_data.RAW", tmp_path)
    (tmp_path / "bad.csv").write_text("not source data")
    with pytest.raises(ValueError, match="checksum"):
        download({"filename": "bad.csv", "sha256": "0" * 64}, offline=True)


def test_future_live_weights_not_used(tmp_path):
    # Small artificial unit-test fixture, never exported as real data.
    fixture = pd.DataFrame({"Tag": ["test-animal"], "Sex": ["F"], "System": ["test"],
        "Year": [2019], "Date.of.birth": ["01/03/2019"], "Date.sold": ["20/09/2019"],
        "Sold.to": ["Abattoir"], "Carcass weight": [20], "Birth.dam.tag": ["test-dam"],
        "Rearing.dam.tag": ["test-dam"], "W.19.09.19": [44], "W.19.09.26": [999]})
    path = tmp_path / "fixture.zip"
    with zipfile.ZipFile(path, "w") as archive:
        archive.writestr("NWFP_sheep_data_v1.csv", fixture.to_csv(index=False))
    result, _ = clean_nwfp(path)
    assert result.iloc[0]["last_measured_live_weight_kg"] == 44
    assert result.iloc[0]["live_measurement_lag_days"] == 1


@pytest.mark.skipif(not all((RAW / source["filename"]).exists() for source in SOURCES.values()), reason="download licensed real sources first")
def test_verified_real_import_counts_and_isolation():
    report = run(offline=True)
    assert {name: entry["rows"] for name, entry in report["datasets"].items()} == {
        "usda_cooking": 173, "nwfp_lamb": 2978, "hararghe_cattle": 61}
    assert report["datasets"]["nwfp_lamb"]["quality"]["dictionary_rows_removed"] == 1
    assert report["competition_real_training"] == "blocked_missing_targets"
    assert not report["synthetic_data_mixed"]
