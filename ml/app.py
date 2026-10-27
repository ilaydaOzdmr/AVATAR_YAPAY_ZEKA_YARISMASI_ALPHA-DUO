import json

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel

from ml.model import predict
from ml.train import (
    MATERIAL_UNIT_COST,
    PRODUCT_CODES,
    REPORT_PATH,
    SERVICE_PER_PACKAGE,
    encode_animal_type,
    train_models,
)


app = FastAPI(title="AVATAR ML Service")
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)
MODELS = train_models()
REPORT = json.loads(REPORT_PATH.read_text(encoding="utf-8"))


class YieldRequest(BaseModel):
    raw_material_kg: float


class ButcheryRequest(BaseModel):
    animal_type: str
    input_weight: float


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/predict/yield")
def predict_yield(payload: YieldRequest):
    features = [[payload.raw_material_kg]]
    predicted_qty = float(predict(MODELS["yield"], features)[0])
    service_cost = predicted_qty * SERVICE_PER_PACKAGE
    total_cost = payload.raw_material_kg * MATERIAL_UNIT_COST + service_cost
    predicted_unit_cost = total_cost / predicted_qty if predicted_qty > 0 else 0.0
    return {
        "predicted_qty": round(predicted_qty, 3),
        "predicted_unit_cost": round(predicted_unit_cost, 3),
    }


@app.post("/predict/butchery")
def predict_butchery(payload: ButcheryRequest):
    try:
        animal_type_code = encode_animal_type(payload.animal_type)
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc

    features = [[animal_type_code, payload.input_weight]]
    weights = predict(MODELS["butchery"], features)[0]
    parts = [
        {"product_code": product_code, "weight": round(float(weight), 3)}
        for product_code, weight in zip(PRODUCT_CODES, weights)
    ]
    return {"parts": parts}


@app.get("/report")
def get_report():
    return REPORT
