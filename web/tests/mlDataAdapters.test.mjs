import test from "node:test";
import assert from "node:assert/strict";
import { ensurePredictionVersion, productionProduct, recordedPartTotals } from "../src/mlContracts.ts";

test("a new model is used only with its matching report and dataset", () => {
  const report = { model_reports: { yield: { model_version: "1.1.0", dataset_version: "data-a" } } };
  ensurePredictionVersion(report, "yield", { model_version: "1.1.0", dataset_version: "data-a" });
  assert.throws(() => ensurePredictionVersion(report, "yield", { model_version: "1.0.0", dataset_version: "data-a" }));
  assert.throws(() => ensurePredictionVersion(report, "yield", { model_version: "1.1.0", dataset_version: "other-data" }));
});

test("recorded package amounts are converted using explicit product weights", () => {
  const output = [{ product_lot_id: 1 }], lots = [{ id: 1, item_id: 1, item_type: "product", unit: "paket" }];
  const product = { id: 1, name: "Narin Sucuk (1 kg)", unit: "paket" };
  assert.deepEqual(productionProduct(output, lots, [product]), { product_type: "sucuk", kg_per_unit: 1 });
  assert.deepEqual(productionProduct(output, lots, [{ ...product, name: "Sucuk (500 g)" }]), { product_type: "sucuk", kg_per_unit: .5 });
  assert.throws(() => productionProduct(output, lots, [{ ...product, name: "Sucuk" }]));
  assert.throws(() => productionProduct(output, lots, [{ ...product, name: "Sucuk Salam (1 kg)" }]));
  assert.throws(() => productionProduct([], lots, [product]));
});

test("the existing total-parts table compares the same weighed cuts and bone", () => {
  const products = [{ id: 2, code: "PRD-101" }, { id: 8, code: "PRD-107" }];
  const items = [{ product_id: 2, weight: 4, is_waste: false }, { product_id: 8, weight: 2, is_waste: false }];
  const prediction = { leg_weight: 4.2, bone_weight: 1.8, waste_weight: 1, total_saleable_weight: 20 };
  assert.deepEqual(recordedPartTotals(items, products, prediction), { actual: 6, predicted: 6 });
  assert.deepEqual(recordedPartTotals([...items, { product_id: 0, weight: .5, is_waste: true }], products, prediction), { actual: 6.5, predicted: 7 });
  assert.throws(() => recordedPartTotals(items, [], prediction));
  assert.throws(() => recordedPartTotals(items, products, {}));
  assert.throws(() => recordedPartTotals([], products, prediction));
});
