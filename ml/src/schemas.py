"""Pydantic request/response schemas for the ML API."""

from __future__ import annotations

from typing import Any, Literal

from pydantic import BaseModel, ConfigDict, Field, model_validator


class RequestBase(BaseModel):
    model_config = ConfigDict(allow_inf_nan=False, extra="forbid")


class LegacyYieldRequest(RequestBase):
    raw_material_kg: float = Field(gt=0, le=10000)
    service_cost: float | None = Field(default=None, ge=0)


class LegacyButcheryRequest(RequestBase):
    animal_type: str
    input_weight: float = Field(gt=0, le=2000)


class YieldV2Request(RequestBase):
    raw_material_kg: float = Field(gt=0, le=10000)
    spice_kg: float = Field(default=0, ge=0, le=500)
    casing_kg: float = Field(default=0, ge=0, le=500)
    other_material_kg: float = Field(default=0, ge=0, le=500)
    material_unit_price_meat: float = Field(default=9.70, ge=0, le=1000)
    material_unit_price_spice: float = Field(default=24.00, ge=0, le=1000)
    material_unit_price_casing: float = Field(default=11.60, ge=0, le=1000)
    material_unit_price_other: float = Field(default=3.50, ge=0, le=1000)
    service_cost: float = Field(default=0, ge=0, le=1000000)
    labor_cost: float = Field(default=0, ge=0, le=1000000)
    service_unit_cost: float | None = Field(default=None, ge=0, le=1000)
    product_type: Literal["sucuk", "salam", "sausage"] = "sucuk"
    actual_finished_qty: float | None = Field(default=None, gt=0, le=100000)

    @model_validator(mode="after")
    def one_service_basis(self) -> "YieldV2Request":
        if self.service_unit_cost is not None and self.service_cost != 0:
            raise ValueError("provide either service_cost or service_unit_cost")
        return self


class ButcheryV2Request(RequestBase):
    animal_type: Literal["lamb", "cattle"]
    live_weight_kg: float | None = Field(default=None, gt=0, le=2000)
    carcass_weight_kg: float = Field(ge=0.001, le=1500)
    breed: str | None = None
    sex: Literal["male", "female"] = "male"
    age_months: float | None = Field(default=None, ge=1, le=120)
    conformation_score: float = Field(default=3, ge=1, le=5)
    fat_score: float = Field(default=3, ge=1, le=5)

    @model_validator(mode="after")
    def carcass_not_above_live(self) -> "ButcheryV2Request":
        if self.live_weight_kg is not None and self.carcass_weight_kg > self.live_weight_kg:
            raise ValueError("carcass_weight_kg cannot exceed live_weight_kg")
        return self


class ModelInfo(BaseModel):
    name: str
    model_version: str
    model_status: str
    dataset_version: str
    targets: list[str]
    training_sample_count: int
    test_sample_count: int
    warnings: list[str]
    data_origin: str
    training_date: str
    feature_names: list[str]
    validation_method: str
    input_ranges: dict[str, Any]


class HealthResponse(BaseModel):
    status: str
    models_loaded: int
    models: list[ModelInfo]


class PredictionResponse(BaseModel):
    model_config = ConfigDict(extra="allow")

    prediction: dict[str, float]
    model_version: str
    dataset_version: str
    model_status: str
    data_origin: str
    applicable_input_range: dict[str, Any]
    warnings: list[str]
    uncertainty: dict[str, Any]
    cost_breakdown: dict[str, float] | None = None
    physical_check: dict[str, Any] | None = None
