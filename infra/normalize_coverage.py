"""Map coverage paths to the repository root without changing measured counts."""
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def normalize() -> None:
    go_report = ROOT / "coverage/go/coverage.out"
    if go_report.exists():
        content = go_report.read_text(encoding="utf-8")
        go_report.write_text(content.replace("avatar/backend/", "backend/"), encoding="utf-8")
    web_report = ROOT / "coverage/web/lcov.info"
    if web_report.exists():
        lines = web_report.read_text(encoding="utf-8").splitlines()
        mapped = []
        for line in lines:
            if line.startswith("SF:"):
                path = line[3:].replace("\\", "/")
                if path.startswith("src/"):
                    path = "web/" + path
                line = "SF:" + path
            mapped.append(line)
        web_report.write_text("\n".join(mapped) + "\n", encoding="utf-8")


if __name__ == "__main__":
    normalize()
