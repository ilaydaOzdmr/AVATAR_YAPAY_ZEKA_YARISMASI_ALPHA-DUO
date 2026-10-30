from fastapi.testclient import TestClient

from ml.app import app

client = TestClient(app)


def test_health_reports_loaded_models():
    response = client.get("/health")
    assert response.status_code == 200
    body = response.json()
    assert body["status"] == "ok"
    assert body["models_loaded"] == 3
    assert {model["name"] for model in body["models"]} == {
        "yield",
        "butchery_lamb",
        "butchery_cattle",
    }


def test_models_and_report_are_available():
    models = client.get("/models")
    assert models.status_code == 200
    assert len(models.json()) == 3
    assert all(item["model_status"] == "synthetic_only" for item in models.json())

    report = client.get("/report")
    assert report.status_code == 200
    body = report.json()
    assert "yield" in body and "butchery" in body
    assert body["model_status"] == "synthetic_only"
    assert "synthetic" in body["note"].lower()


def test_yield_v2_returns_metadata_and_uncertainty():
    response = client.post(
        "/predict/v2/yield",
        json={"raw_material_kg": 100, "spice_kg": 2, "casing_kg": 3},
    )
    assert response.status_code == 200
    body = response.json()
    assert body["prediction"]["finished_qty"] > 0
    assert 0.5 < body["prediction"]["yield_pct"] < 1.0
    assert body["model_version"] == "1.0.0"
    assert body["model_status"] == "synthetic_only"
    assert body["warnings"]
    assert body["uncertainty"]["finished_qty"]["test_rmse"] >= 0
    assert body["cost_breakdown"]["total_cost"] > 0
    assert body["ekp_kind"] == "estimated_pre_production"


def test_yield_v2_separates_accounting_ekp():
    response = client.post(
        "/predict/v2/yield",
        json={"raw_material_kg": 100, "service_cost": 400, "actual_finished_qty": 80},
    )
    assert response.status_code == 200
    body = response.json()
    assert body["ekp_kind"] == "actual_accounting"
    assert body["accounting_ekp"] == round(body["cost_breakdown"]["total_cost"] / 80, 4)


def test_yield_v2_warns_outside_training_range():
    response = client.post("/predict/v2/yield", json={"raw_material_kg": 5000})
    assert response.status_code == 200
    assert any("outside the training range" in warning for warning in response.json()["warnings"])


def test_yield_v2_rejects_invalid_input():
    assert client.post("/predict/v2/yield", json={"raw_material_kg": 0}).status_code == 422
    assert client.post("/predict/v2/yield", json={"raw_material_kg": -5}).status_code == 422


def test_butchery_v2_returns_physical_check():
    response = client.post(
        "/predict/v2/butchery",
        json={"animal_type": "lamb", "live_weight_kg": 45, "carcass_weight_kg": 21},
    )
    assert response.status_code == 200
    body = response.json()
    assert body["prediction"]["leg_weight"] > 0
    assert body["prediction"]["total_saleable_weight"] > 0
    assert body["physical_check"]["carcass_weight_kg"] == 21
    assert "parts_over_carcass" in body["physical_check"]
    assert body["warnings"]


def test_butchery_v2_rejects_impossible_carcass():
    response = client.post(
        "/predict/v2/butchery",
        json={"animal_type": "cattle", "live_weight_kg": 400, "carcass_weight_kg": 500},
    )
    assert response.status_code == 422


def test_legacy_endpoints_keep_their_contract():
    yield_response = client.post("/predict/yield", json={"raw_material_kg": 100})
    assert yield_response.status_code == 200
    yield_body = yield_response.json()
    assert "predicted_qty" in yield_body
    assert "predicted_unit_cost" in yield_body

    butchery_response = client.post(
        "/predict/butchery", json={"animal_type": "kuzu", "input_weight": 25}
    )
    assert butchery_response.status_code == 200
    parts = butchery_response.json()["parts"]
    assert len(parts) == 7
    assert {part["product_code"] for part in parts} == {
        "PRD-101",
        "PRD-102",
        "PRD-103",
        "PRD-104",
        "PRD-105",
        "PRD-106",
        "PRD-107",
    }


def test_legacy_butchery_rejects_unknown_animal():
    response = client.post(
        "/predict/butchery", json={"animal_type": "devekuşu", "input_weight": 25}
    )
    assert response.status_code == 400
