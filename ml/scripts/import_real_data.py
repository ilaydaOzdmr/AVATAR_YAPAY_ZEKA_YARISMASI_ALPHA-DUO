"""Download licensed measurements, verify hashes, and build reference tables.

Run: python -m ml.scripts.import_real_data [--offline]
No rows are generated, no missing targets are imputed, and no training or
artifact registration occurs here. See docs/DATA_SOURCES.md for attribution.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import urllib.request
import zipfile
from pathlib import Path

import numpy as np
import pandas as pd

from ml.src.config import DATA_DIR, REPORTS_DIR, PRODUCTION_TARGETS, BUTCHERY_TARGETS

RAW = DATA_DIR / "raw" / "verified_real"
REAL = DATA_DIR / "real"
SOURCES = {
    "usda_cooking": {
        "url": "https://ndownloader.figshare.com/files/43365402",
        "filename": "usda_cooking.csv",
        "license": "CC0-1.0",
        "source": "https://doi.org/10.15482/USDA.ADC/1409031",
        "sha256": "e6144d36488781c3eab2428bd4fd576d85092a93fc709aa05868d20fe6ba1f4f",
        "unit_of_observation": "published food/preparation summary, not an individual batch",
    },
    "nwfp_lamb": {
        "url": "https://data.mendeley.com/public-files/datasets/xy3ndcy8jd/files/20ced925-37ec-4991-80af-82b8ea2cc681/file_downloaded",
        "filename": "nwfp.zip",
        "license": "CC-BY-4.0",
        "source": "https://doi.org/10.17632/xy3ndcy8jd.1",
        "sha256": "0e63f6c39170ad72d6a2d7e6e658bed5f32d4d48b260be57c95f00abdfc4ebe1",
        "unit_of_observation": "one lamb (measured on farm and at abattoir)",
    },
    "hararghe_cattle": {
        "url": "https://data.mendeley.com/public-files/datasets/w9wbn9mpr8/files/ef186adb-558d-472d-b95f-5e851eaddd43/file_downloaded",
        "filename": "hararghe.xlsx",
        "license": "CC-BY-4.0",
        "source": "https://doi.org/10.17632/w9wbn9mpr8.1",
        "sha256": "bcdc200aa606146d17e01ddccf53a810cdb587989b38ea9ee159a6e18fde4397",
        "unit_of_observation": "one slaughtered animal (no animal/farm identifiers supplied)",
    },
}


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def download(source: dict, offline: bool) -> Path:
    target = RAW / source["filename"]
    if not target.exists():
        if offline:
            raise FileNotFoundError(target)
        RAW.mkdir(parents=True, exist_ok=True)
        req = urllib.request.Request(source["url"], headers={"User-Agent": "AVATAR-data-audit/1.0"})
        with urllib.request.urlopen(req, timeout=60) as response:
            payload = response.read(10_000_001)
        if len(payload) > 10_000_000:
            raise ValueError("source exceeded 10 MB download limit")
        if hashlib.sha256(payload).hexdigest() != source["sha256"]:
            raise ValueError(f"checksum mismatch for {source['filename']}; review upstream version")
        target.write_bytes(payload)
    if digest(target) != source["sha256"]:
        raise ValueError(f"checksum mismatch for cached {target}")
    return target


def clean_usda(path: Path) -> tuple[pd.DataFrame, dict]:
    raw = pd.read_csv(path)
    raw.columns = raw.columns.str.strip()
    frame = raw.rename(columns={
        "Food Group Code": "food_group_code", "NDB": "ndb_id",
        "Yield Description": "description", "Preparation Method1": "preparation_method",
        "Cooking Yield %": "cooking_yield_pct", "n": "study_sample_count",
        "SD": "yield_sd", "Yield Minimum %": "yield_min_pct",
        "Yield Maximum %": "yield_max_pct", "Moisture Gain/ Loss %": "moisture_change_pct",
        "Fat   Gain/ Loss %": "fat_change_pct", "Release Year2": "release_year",
    })
    frame.insert(0, "source_row", np.arange(2, len(frame) + 2))
    for column in ["description", "preparation_method"]:
        frame[column] = frame[column].str.strip()
    valid = frame["cooking_yield_pct"].between(0, 100) & frame["description"].notna()
    return frame.loc[valid].copy(), {"raw_rows": len(raw), "rejected_rows": int((~valid).sum()),
                                    "note": "n is a published sample count; never expanded into pseudo-records"}


def clean_nwfp(path: Path) -> tuple[pd.DataFrame, dict]:
    # Read a named member; never extract paths from an untrusted archive.
    with zipfile.ZipFile(path) as archive:
        info = archive.getinfo("NWFP_sheep_data_v1.csv")
        if info.file_size > 10_000_000:
            raise ValueError("unexpected archive size")
        raw = pd.read_csv(archive.open(info), low_memory=False, dtype=str)
    year = pd.to_numeric(raw["Year"], errors="coerce")
    # The first data row is a human-readable column dictionary, not an animal.
    animals = raw.loc[year.notna()].copy()
    sold_date = pd.to_datetime(animals["Date.sold"], dayfirst=True, errors="coerce")
    birth_date = pd.to_datetime(animals["Date.of.birth"], dayfirst=True, errors="coerce")
    weights = animals.filter(regex=r"^W\.\d{2}\.\d{2}\.\d{2}$").apply(pd.to_numeric, errors="coerce")
    dates = pd.to_datetime(weights.columns.str[2:], format="%y.%m.%d")
    last_weight = pd.Series(np.nan, index=animals.index)
    last_date = pd.Series(pd.NaT, index=animals.index, dtype="datetime64[ns]")
    for pos in np.argsort(dates):
        when = dates[pos]
        eligible = weights.iloc[:, pos].gt(0) & sold_date.ge(when) & birth_date.le(when)
        last_weight.loc[eligible] = weights.iloc[:, pos].loc[eligible]
        last_date.loc[eligible] = when
    # Values at 4/8 weeks are interpolated in the publication; don't relabel
    # those values as direct measurements or as slaughter live weight.
    frame = pd.DataFrame({
        "animal_id": animals["Tag"].str.strip(), "sex": animals["Sex"],
        "production_year": year.loc[animals.index].astype(int), "grazing_system": animals["System"],
        "birth_dam_id": animals["Birth.dam.tag"].str.strip(),
        "rearing_dam_id": animals["Rearing.dam.tag"].str.strip(),
        "last_measured_live_weight_kg": last_weight, "live_measurement_date": last_date,
        "sale_date": sold_date, "live_measurement_lag_days": (sold_date - last_date).dt.days,
        "cold_carcass_weight_kg": pd.to_numeric(animals["Carcass weight"], errors="coerce"),
    })
    frame.insert(0, "source_row", animals.index + 2)
    missing = frame[["last_measured_live_weight_kg", "cold_carcass_weight_kg", "sale_date"]].isna().any(axis=1)
    physical = frame["cold_carcass_weight_kg"].gt(0) & frame["cold_carcass_weight_kg"].le(last_weight)
    eligible = ~missing & physical & animals["Sold.to"].eq("Abattoir")
    # Preserve all real animal records and missing cells in the reference table;
    # this flag only identifies paired measurements usable for future research.
    frame["paired_measurements_valid"] = eligible
    return frame, {"raw_rows": len(raw), "animal_rows": len(frame),
        "dictionary_rows_removed": int(year.isna().sum()),
        "duplicate_animal_ids": int(frame["animal_id"].duplicated().sum()),
        "paired_measurements_valid": int(eligible.sum()),
        "missing_pair_or_date": int(missing.sum()),
        "invalid_observed_weight_pairs": int((~missing & ~physical).sum()),
        "note": "latest measured live weight before sale, NOT necessarily pre-slaughter weight; missing values retained"}


def clean_hararghe(path: Path) -> tuple[pd.DataFrame, dict]:
    raw = pd.read_excel(path, sheet_name="Sheet1")
    raw.columns = raw.columns.str.strip()
    frame = pd.DataFrame({
        "source_row": np.arange(2, len(raw) + 2),
        "live_weight_kg": pd.to_numeric(raw["Wt (kg)"], errors="coerce"),
        "hot_carcass_weight_kg": pd.to_numeric(raw["HCW"], errors="coerce"),
        "empty_body_weight_kg": pd.to_numeric(raw["EBW"], errors="coerce"),
        "reported_hdp_pct": pd.to_numeric(raw["HDP"], errors="coerce"),
    })
    frame["calculated_dressing_pct"] = 100 * frame["hot_carcass_weight_kg"] / frame["live_weight_kg"]
    valid = frame["live_weight_kg"].gt(0) & frame["hot_carcass_weight_kg"].gt(0) & frame["hot_carcass_weight_kg"].le(frame["live_weight_kg"])
    # Do not treat ambiguous 'Bone' as the carcass bone target, or HDP as cutting yield.
    return frame.loc[valid].copy(), {"raw_rows": len(raw), "rejected_rows": int((~valid).sum()),
        "publication_claimed_animals": 62, "file_observed_animals": len(raw),
        "hdp_disagrees_over_0_1pp": int((frame["reported_hdp_pct"] - frame["calculated_dressing_pct"]).abs().gt(0.1).sum()),
        "note": "source row is provenance, not a measured animal ID; no independent farm/site test possible"}


def suitability(frame: pd.DataFrame) -> dict:
    # These imports are references, never silently accepted as production data.
    return {task: {"eligible_for_training": False,
        "missing_targets": sorted(set(targets) - set(frame.columns)),
        "reason": "incomplete competition targets and/or incompatible observation unit"}
        for task, targets in {"production": PRODUCTION_TARGETS, "butchery": BUTCHERY_TARGETS}.items()}


def run(offline: bool = False) -> dict:
    REAL.mkdir(parents=True, exist_ok=True)
    report = {"verified_on": "2026-10-31", "datasets": {},
              "competition_real_training": "blocked_missing_targets",
              "synthetic_data_mixed": False}
    for name, cleaner in [("usda_cooking", clean_usda), ("nwfp_lamb", clean_nwfp), ("hararghe_cattle", clean_hararghe)]:
        source = SOURCES[name]
        path = download(source, offline)
        frame, quality = cleaner(path)
        frame["data_origin"] = "real_reference"
        frame["source_doi"] = source["source"]
        frame["license"] = source["license"]
        frame.to_csv(REAL / f"{name}.csv", index=False)
        report["datasets"][name] = {**source, "rows": len(frame), "columns": frame.columns.tolist(),
            "null_cells": int(frame.isna().sum().sum()), "quality": quality,
            "duplicate_rows_excluding_provenance": int(frame.drop(columns=["source_row"]).duplicated().sum()),
            "numeric_summary": {column: {"observed_n": int(values.notna().sum()),
                "min": float(values.min()), "max": float(values.max()), "mean": float(values.mean()),
                "iqr_outlier_count": int(((values < values.quantile(.25) - 1.5 * (values.quantile(.75) - values.quantile(.25))) |
                                          (values > values.quantile(.75) + 1.5 * (values.quantile(.75) - values.quantile(.25)))).sum())}
                for column in frame.select_dtypes(include="number") if column != "source_row"
                for values in [frame[column]]},
            "suitability": suitability(frame),
            "cleaned_sha256": digest(REAL / f"{name}.csv")}
    REPORTS_DIR.mkdir(parents=True, exist_ok=True)
    (REPORTS_DIR / "real_data_inventory.json").write_text(json.dumps(report, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    return report


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--offline", action="store_true")
    args = parser.parse_args()
    result = run(args.offline)
    print(json.dumps({key: value["quality"] for key, value in result["datasets"].items()}, indent=2))
