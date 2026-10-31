"""FastAPI service.

Models are loaded from versioned artifacts at process start. Missing artifacts
return HTTP 503 instead of silently training or returning random values.
"""

from __future__ import annotations

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

from ml.src.schemas import (
    ButcheryV2Request,
    HealthResponse,
    LegacyButcheryRequest,
    LegacyYieldRequest,
    ModelInfo,
    PredictionResponse,
    YieldV2Request,
)
from ml.src.service import MLService, ModelUnavailableError
from ml.src.config import MODEL_VERSION

app = FastAPI(title="AVATAR ML Service", version=MODEL_VERSION)
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

SERVICE = MLService()


@app.exception_handler(ModelUnavailableError)
def model_unavailable_handler(_request, exc: ModelUnavailableError) -> JSONResponse:
    return JSONResponse(
        status_code=503,
        content={"error": str(exc), "model_status": "unavailable"},
    )


@app.exception_handler(ValueError)
def value_error_handler(_request, exc: ValueError) -> JSONResponse:
    return JSONResponse(status_code=400, content={"error": str(exc)})


@app.get("/health", response_model=HealthResponse)
def health() -> dict:
    return SERVICE.health()


@app.get("/models", response_model=list[ModelInfo])
def models() -> list[dict]:
    return SERVICE.models()


@app.get("/report")
def report() -> dict:
    try:
        return SERVICE.report()
    except ModelUnavailableError as exc:
        raise HTTPException(status_code=503, detail=str(exc)) from exc


@app.post("/predict/v2/yield", response_model=PredictionResponse)
def predict_yield_v2(payload: YieldV2Request) -> dict:
    return SERVICE.predict_yield_v2(payload)


@app.post("/predict/v2/butchery", response_model=PredictionResponse)
def predict_butchery_v2(payload: ButcheryV2Request) -> dict:
    return SERVICE.predict_butchery_v2(payload)


# --- v1 compatibility -------------------------------------------------------
@app.post("/predict/yield")
def predict_yield(payload: LegacyYieldRequest) -> dict:
    return SERVICE.predict_yield_legacy(payload)


@app.post("/predict/butchery")
def predict_butchery(payload: LegacyButcheryRequest) -> dict:
    return SERVICE.predict_butchery_legacy(payload)
