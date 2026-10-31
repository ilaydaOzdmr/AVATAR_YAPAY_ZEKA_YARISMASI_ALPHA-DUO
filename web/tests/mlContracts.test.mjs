import test from "node:test";
import assert from "node:assert/strict";
import { yieldPayload, butcheryPayload, initialButchery, modelMetrics, numberText, productionProduct,
  statusText, warningText, YIELD_INITIAL, ensurePredictionVersion } from "../src/mlContracts.ts";

test("PDF form sends all materials separately and service as a per-unit rate", () => {
  const payload = yieldPayload({ ...YIELD_INITIAL, actual_finished_qty: "85" });
  assert.equal(payload.raw_material_kg, 100);
  assert.equal(payload.spice_kg, 2);
  assert.equal(payload.casing_kg, 5);
  assert.equal(payload.service_unit_cost, 5.5);
  assert.equal(payload.service_cost, 0);
  assert.equal(payload.actual_finished_qty, 85);
  assert.equal(payload.labor_cost, 0);
  const fixed = yieldPayload({ ...YIELD_INITIAL, service_basis: "total", service_value: "467.5" });
  assert.equal(fixed.service_cost, 467.5);
  assert.ok(!("service_unit_cost" in fixed));
  assert.ok(!("actual_finished_qty" in fixed));
});

test("forms reject empty/invalid inputs and conflicting live/carcass values", () => {
  for (const value of ["", "NaN", "100kg", "Infinity", "-1", "0"]) assert.throws(() => yieldPayload({ ...YIELD_INITIAL, raw_material_kg: value }));
  assert.throws(() => yieldPayload({ ...YIELD_INITIAL, material_unit_price_meat: "" }));
  assert.throws(() => butcheryPayload({ ...initialButchery("lamb"), live_weight_kg: "20", carcass_weight_kg: "25" }));
  assert.throws(() => butcheryPayload({ ...initialButchery("cattle"), fat_score: "6" }));
  assert.throws(() => butcheryPayload({ ...initialButchery("lamb"), carcass_weight_kg: "" }));
});

test("carcass stays measured and optional missing attributes are omitted", () => {
  for (const animal of ["lamb", "cattle"]) {
    const form = initialButchery(animal), payload = butcheryPayload(form);
    assert.equal(payload.animal_type, animal);
    assert.equal(payload.carcass_weight_kg, Number(form.carcass_weight_kg));
    assert.ok(!("live_weight_kg" in payload));
    assert.ok(!("age_months" in payload));
    assert.ok(!("breed" in payload));
  }
  assert.equal(butcheryPayload({ ...initialButchery("cattle"), breed: "Angus", age_months: "24", live_weight_kg: "550" }).breed, "Angus");
});

test("metrics are shown only for matching model AND dataset versions", () => {
  const metric = { test: { mae: 2.097 } };
  const model = { name: "butchery_cattle", model_version: "1.1.0", dataset_version: "data-b" };
  const report = { model_reports: { butchery_cattle: { ...model, metrics: { total_saleable_weight: metric } } } };
  assert.equal(modelMetrics(report, model, "total_saleable_weight"), metric);
  assert.equal(modelMetrics(report, { ...model, model_version: "1.2.0" }, "total_saleable_weight"), undefined);
  assert.equal(modelMetrics(report, { ...model, dataset_version: "data-c" }, "total_saleable_weight"), undefined);
  assert.equal(modelMetrics(null, model, "total_saleable_weight"), undefined);
});

test("missing metrics are not formatted as zero or a success percentage", () => {
  assert.equal(numberText(null), "—"); assert.equal(numberText(NaN), "—");
  assert.equal(numberText(0), "0,00");
  assert.match(statusText("synthetic_only"), /Sentetik/);
  assert.match(statusText(undefined), /bilinmiyor/);
  assert.match(warningText("raw_material_kg=5000 is outside the training range [60, 160]; prediction is extrapolated."), /Et: 5000 eğitim aralığının/);
});

test("prediction responses must match the displayed report versions", () => {
  const report = { model_reports: { yield: { model_version: "1.1.0", dataset_version: "data-a" } } };
  ensurePredictionVersion(report, "yield", { model_version: "1.1.0", dataset_version: "data-a" });
  assert.throws(() => ensurePredictionVersion(report, "yield", { model_version: "1.1.0", dataset_version: "data-b" }));
  assert.throws(() => ensurePredictionVersion(report, "yield", { model_version: "1.2.0", dataset_version: "data-a" }));
});

test("history identifies output product and converts only explicit pack weights", () => {
  const output = [{ product_lot_id: 1 }];
  const lots = [{ id: 1, item_id: 1, item_type: "product", unit: "paket" }];
  const product = { id: 1, name: "Narin Sucuk (1 kg)", unit: "paket" };
  assert.deepEqual(productionProduct(output, lots, [product]), { product_type: "sucuk", kg_per_unit: 1 });
  assert.deepEqual(productionProduct(output, lots, [{ ...product, name: "Sucuk (500 g)" }]), { product_type: "sucuk", kg_per_unit: .5 });
  assert.throws(() => productionProduct(output, lots, [{ ...product, name: "Sucuk" }]));
  assert.throws(() => productionProduct(output, lots, [{ ...product, name: "Sucuk Salam (1 kg)" }]));
  assert.throws(() => productionProduct([], lots, [product]));
  assert.throws(() => productionProduct(output, lots, []));
});
