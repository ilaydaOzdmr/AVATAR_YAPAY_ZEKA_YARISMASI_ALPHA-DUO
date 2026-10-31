from infra import normalize_coverage


def test_report_path_mapping_preserves_every_measured_execution_count(tmp_path, monkeypatch):
    monkeypatch.setattr(normalize_coverage, "ROOT", tmp_path)
    go = tmp_path / "coverage/go/coverage.out"
    web = tmp_path / "coverage/web/lcov.info"
    go.parent.mkdir(parents=True)
    web.parent.mkdir(parents=True)
    go.write_text("mode: atomic\navatar/backend/internal/repo/repo.go:1.1,4.2 3 17\n")
    web.write_text("TN:\nSF:src\\m3\\App.tsx\nDA:12,9\nBRDA:12,0,0,8\nend_of_record\n")
    normalize_coverage.normalize()
    assert go.read_text().endswith("backend/internal/repo/repo.go:1.1,4.2 3 17\n")
    assert web.read_text().splitlines()[1:] == ["SF:web/src/m3/App.tsx", "DA:12,9", "BRDA:12,0,0,8", "end_of_record"]
    first = (go.read_bytes(), web.read_bytes())
    normalize_coverage.normalize()
    assert first == (go.read_bytes(), web.read_bytes())


def test_missing_coverage_reports_are_not_fabricated(tmp_path, monkeypatch):
    monkeypatch.setattr(normalize_coverage, "ROOT", tmp_path)
    normalize_coverage.normalize()
    assert not (tmp_path / "coverage").exists()
