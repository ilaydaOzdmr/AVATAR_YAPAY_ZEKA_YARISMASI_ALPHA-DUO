"""Make both ``ml.*`` and ``src.*`` importable in local and container runs."""

import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
ML_ROOT = Path(__file__).resolve().parents[1]

for path in (str(REPO_ROOT), str(ML_ROOT)):
    if path not in sys.path:
        sys.path.insert(0, path)
