import test from "node:test";
import assert from "node:assert/strict";
import { butcheryComparison, productionFeatures, predictionError, nonnegativeNumber } from "../src/predictionInputs.ts";

test("PDF recipe preserves 100 kg meat, 2 kg spice, 5 kg casing", () => {
  const cards = [1, 2, 3].map((id) => ({ id, code: `MAT-00${id}`, unit: "kg" }));
  const lots = [1, 2, 3].map((id) => ({ id, item_id: id, item_type: "material", unit: "kg", unit_cost: [9.7, 24, 11.6][id - 1] }));
  const inputs = [100, 2, 5].map((qty, i) => ({ material_lot_id: i + 1, qty }));
  const result = productionFeatures(inputs, lots, cards);
  assert.equal(result.raw_material_kg, 100);
  assert.equal(result.spice_kg, 2);
  assert.equal(result.casing_kg, 5);
  assert.equal(result.material_unit_price_meat, 9.7);
  assert.throws(() => productionFeatures(inputs, lots.map((l) => ({ ...l, unit: "g" })), cards));
  assert.throws(() => productionFeatures(inputs, [], cards));
});

test("blank, invalid, negative and nonfinite form fields are rejected", () => {
  for (const value of ["", " ", "NaN", "100kg", "-1", "Infinity"]) assert.throws(() => nonnegativeNumber(value, "test"));
  assert.equal(nonnegativeNumber("0", "test"), 0);
});

test("order errors use actual denominator and preserve unavailable predictions", () => {
  assert.deepEqual(predictionError(80, 100), { signed: -20, absolute: 20, percentage: 20 });
  assert.deepEqual(predictionError(100, 80), { signed: 20, absolute: 20, percentage: 25 });
  assert.equal(predictionError(10, 0).percentage, null);
  assert.equal(predictionError(null, 80).absolute, null);
});

test("butchery compares the recorded cuts only, excludes bone and waste", () => {
  const items = [{ product_id: 2, weight: 4, is_waste: false }, { product_id: 8, weight: 2, is_waste: false },
    { product_id: 999, weight: 1, is_waste: true }];
  const products = [{ id: 2, code: "PRD-101" }, { id: 8, code: "PRD-107" }];
  const result = butcheryComparison(items, products, { leg_weight: 5, total_saleable_weight: 20 });
  assert.equal(result.predicted, 5);
  assert.equal(result.actual, 4);
  assert.equal(result.parts.length, 1);
  assert.throws(() => butcheryComparison(items, [], { leg_weight: 5 }));
});
