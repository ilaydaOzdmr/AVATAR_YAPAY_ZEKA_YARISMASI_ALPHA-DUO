"""Download the public USDA ERS files used as real price inputs.

Run from the repository root or from ``ml``:

    python ml/scripts/download_usda.py

The script is idempotent: it re-downloads and rewrites the manifest so the
dataset version is reproducible.
"""

from __future__ import annotations

import hashlib
import json
import urllib.request
import zipfile
from pathlib import Path
from typing import Any

ML_ROOT = Path(__file__).resolve().parents[1]
RAW_DIR = ML_ROOT / "data" / "raw" / "usda"

MEAT_PRICE_SPREADS = "Meat Price Spreads"
LIVESTOCK_DOMESTIC_DATA = "Livestock and Meat Domestic Data"

SOURCES: list[dict[str, str]] = [
    {
        "name": "historical_monthly_price_spreads.csv",
        "url": "https://www.ers.usda.gov/media/5028/historical-monthly-price-spread-data-for-beef-pork-broilers.csv?v=98757",
        "product": MEAT_PRICE_SPREADS,
    },
    {
        "name": "retail_prices.csv",
        "url": "https://www.ers.usda.gov/media/5024/retail-prices-for-beef-pork-poultry-cuts-eggs-and-dairy-products.csv?v=25627",
        "product": MEAT_PRICE_SPREADS,
    },
    {
        "name": "choice_beef_values.csv",
        "url": "https://www.ers.usda.gov/media/5020/choice-beef-values-and-spreads-and-the-all-fresh-retail-value.csv?v=29592",
        "product": MEAT_PRICE_SPREADS,
    },
    {
        "name": "summary_retail_spreads.csv",
        "url": "https://www.ers.usda.gov/media/5025/summary-of-retail-prices-and-price-spreads.csv?v=31589",
        "product": MEAT_PRICE_SPREADS,
    },
    {
        "name": "livestock_prices.xlsx",
        "url": "https://www.ers.usda.gov/media/5536/livestock-prices.xlsx?v=37005",
        "product": LIVESTOCK_DOMESTIC_DATA,
    },
    {
        "name": "wholesale_prices.xlsx",
        "url": "https://www.ers.usda.gov/media/5538/wholesale-prices.xlsx?v=79003",
        "product": LIVESTOCK_DOMESTIC_DATA,
    },
    {
        "name": "livestock_domestic_data.zip",
        "url": "https://www.ers.usda.gov/media/5539/livestock-domestic-data-machine-readable-files.zip?v=88464",
        "product": LIVESTOCK_DOMESTIC_DATA,
    },
]


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def download(name: str, url: str) -> Path:
    RAW_DIR.mkdir(parents=True, exist_ok=True)
    target = RAW_DIR / name
    request = urllib.request.Request(url, headers={"User-Agent": "avatar-ml/1.0"})
    with urllib.request.urlopen(request, timeout=120) as response:
        target.write_bytes(response.read())
    return target


def main() -> None:
    manifest: dict[str, Any] = {
        "provider": "USDA Economic Research Service",
        "license": "U.S. Government work; public domain",
        "products": [
            "https://www.ers.usda.gov/data-products/meat-price-spreads",
            "https://www.ers.usda.gov/data-products/livestock-and-meat-domestic-data",
        ],
        "files": [],
    }
    for source in SOURCES:
        path = download(source["name"], source["url"])
        manifest["files"].append(
            {
                "name": source["name"],
                "url": source["url"],
                "product": source["product"],
                "bytes": path.stat().st_size,
                "sha256": sha256(path),
            }
        )
        print(f"downloaded {source['name']} ({path.stat().st_size} bytes)")

    archive = RAW_DIR / "livestock_domestic_data.zip"
    extract_dir = RAW_DIR / "livestock_domestic_data"
    if archive.exists():
        extract_dir.mkdir(parents=True, exist_ok=True)
        with zipfile.ZipFile(archive) as zipped:
            zipped.extractall(extract_dir)
        print(f"extracted {archive.name}")

    (ML_ROOT / "data" / "usda_manifest.json").write_text(
        json.dumps(manifest, indent=2, ensure_ascii=False) + "\n", encoding="utf-8"
    )
    print(f"wrote {ML_ROOT / 'data' / 'usda_manifest.json'}")


if __name__ == "__main__":
    main()
