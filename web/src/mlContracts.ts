import { nonnegativeNumber } from "./predictionInputs.ts";

export type ModelName = "yield" | "butchery_lamb" | "butchery_cattle";
export type ProductType = "sucuk" | "salam" | "sausage";
export type AnimalType = "lamb" | "cattle";
export interface InputRange { min?: number; max?: number; allowed?: string[] }
export interface ModelInfo {
  name: ModelName; model_version: string; model_status: string; data_origin: string;
  dataset_version: string; training_date: string; targets: string[]; feature_names: string[];
  training_sample_count: number; test_sample_count: number; warnings: string[];
  validation_method: string; input_ranges: Record<string, InputRange>;
}
export interface MLHealth { status: string; models_loaded: number; models: ModelInfo[] }
export interface Metrics { mae: number | null; rmse: number | null; r2: number | null; mape: number | null }
export interface TargetReport { train: Metrics; test: Metrics; proportional_baseline_test: Metrics; train_test_mae_gap: number }
export interface MLReport {
  model_reports: Partial<Record<ModelName, { model_version: string; dataset_version: string;
    validation_sample_counts: number[]; metrics: Record<string, TargetReport> }>>;
}
export interface Uncertainty { test_rmse: number; test_mae: number; coverage_validated: boolean }
export interface PredictionMeta {
  model_version: string; dataset_version: string; model_status: string; data_origin: string; warnings: string[];
  uncertainty: Record<string, Uncertainty | null>;
}
export interface YieldResponse extends PredictionMeta {
  prediction: { finished_qty: number; yield_pct: number; estimated_ekp: number };
  cost_breakdown: Record<string, number>; accounting_cost_breakdown?: Record<string, number>;
  accounting_ekp?: number;
  comparison?: { actual: number; predicted: number; absolute_error: number; percentage_error: number };
}
export interface ButcheryResponse extends PredictionMeta {
  prediction: Record<string, number>;
  physical_check: { carcass_weight_kg: number; parts_sum_kg: number; saleable_sum_kg: number; parts_over_carcass: boolean };
}
export interface YieldForm {
  product_type: ProductType; raw_material_kg: string; spice_kg: string; casing_kg: string;
  other_material_kg: string; material_unit_price_meat: string; material_unit_price_spice: string;
  material_unit_price_casing: string; material_unit_price_other: string;
  service_basis: "unit" | "total"; service_value: string; labor_cost: string; actual_finished_qty: string;
}
export interface ButcheryForm {
  animal_type: AnimalType; carcass_weight_kg: string; live_weight_kg: string;
  breed: string; sex: "male" | "female"; age_months: string; conformation_score: string; fat_score: string;
}
export const YIELD_INITIAL: YieldForm = {
  product_type: "sucuk", raw_material_kg: "100", spice_kg: "2", casing_kg: "5", other_material_kg: "0",
  material_unit_price_meat: "9.70", material_unit_price_spice: "24", material_unit_price_casing: "11.60",
  material_unit_price_other: "3.50", service_basis: "unit", service_value: "5.50", labor_cost: "0", actual_finished_qty: "",
};
export function initialButchery(animal_type: AnimalType): ButcheryForm {
  return { animal_type, carcass_weight_kg: animal_type === "lamb" ? "21" : "300", live_weight_kg: "",
    breed: "", sex: "male", age_months: "", conformation_score: "3", fat_score: "3" };
}
export const FEATURE_LABELS: Record<string, string> = {
  raw_material_kg: "Et", spice_kg: "Baharat", casing_kg: "Bağırsak", other_material_kg: "Diğer malzeme",
  material_unit_price_meat: "Et birim fiyatı", material_unit_price_spice: "Baharat birim fiyatı",
  material_unit_price_casing: "Bağırsak birim fiyatı", material_unit_price_other: "Diğer malzeme birim fiyatı",
  service_value: "Hizmet bedeli", service_cost: "Hizmet bedeli", service_unit_cost: "Birim hizmet tarifesi", labor_cost: "Ek işçilik", actual_finished_qty: "Fiili mamul miktarı",
  product_type: "Ürün tipi", breed: "Irk", sex: "Cinsiyet", age_months: "Yaş",
  carcass_weight_kg: "Karkas", live_weight_kg: "Canlı ağırlık", conformation_score: "Konformasyon skoru", fat_score: "Yağ skoru",
};
export const PART_LABELS: Record<string, string> = {
  leg_weight: "But", loin_weight: "Sırt", shoulder_weight: "Omuz", rib_weight: "Kaburga",
  mince_weight: "Kıyma", diced_meat_weight: "Kuşbaşı", bone_weight: "Kemik", waste_weight: "Fire",
  total_saleable_weight: "Toplam satılabilir", cutting_yield_pct: "Parçalama verimi",
};
export function numberText(value: number | null | undefined, digits = 2): string {
  return typeof value === "number" && Number.isFinite(value)
    ? value.toLocaleString("tr-TR", { minimumFractionDigits: digits, maximumFractionDigits: digits }) : "—";
}
export function statusText(status?: string): string {
  return status === "synthetic_only" ? "Sentetik veriyle eğitim"
    : status === "validated_real" ? "Gerçek veriyle doğrulanmış" : "Doğrulama durumu bilinmiyor";
}
export function originText(origin?: string): string {
  return origin?.startsWith("synthetic") ? "Sentetik" : origin === "real_reference" ? "Gerçek referans" : origin ?? "Bilinmiyor";
}
function bounded(value: string, name: string, min: number, max: number): number {
  const numeric = nonnegativeNumber(value, FEATURE_LABELS[name] ?? name);
  if (numeric < min || numeric > max) throw new Error(`${FEATURE_LABELS[name] ?? name}: ${min}–${max} aralığında olmalı.`);
  return numeric;
}
export function yieldPayload(form: YieldForm) {
  if (!["sucuk", "salam", "sausage"].includes(form.product_type)) throw new Error("Ürün tipi desteklenmiyor.");
  return {
    product_type: form.product_type,
    raw_material_kg: bounded(form.raw_material_kg, "raw_material_kg", .001, 10000),
    spice_kg: bounded(form.spice_kg, "spice_kg", 0, 500),
    casing_kg: bounded(form.casing_kg, "casing_kg", 0, 500),
    other_material_kg: bounded(form.other_material_kg, "other_material_kg", 0, 500),
    material_unit_price_meat: bounded(form.material_unit_price_meat, "material_unit_price_meat", 0, 1000),
    material_unit_price_spice: bounded(form.material_unit_price_spice, "material_unit_price_spice", 0, 1000),
    material_unit_price_casing: bounded(form.material_unit_price_casing, "material_unit_price_casing", 0, 1000),
    material_unit_price_other: bounded(form.material_unit_price_other, "material_unit_price_other", 0, 1000),
    service_cost: form.service_basis === "total" ? bounded(form.service_value, "service_value", 0, 1000000) : 0,
    ...(form.service_basis === "unit" ? { service_unit_cost: bounded(form.service_value, "service_value", 0, 1000) } : {}),
    labor_cost: bounded(form.labor_cost, "labor_cost", 0, 1000000),
    ...(form.actual_finished_qty.trim() ? { actual_finished_qty: bounded(form.actual_finished_qty, "actual_finished_qty", .001, 100000) } : {}),
  };
}
export function butcheryPayload(form: ButcheryForm) {
  if (!["lamb", "cattle"].includes(form.animal_type)) throw new Error("Hayvan türü desteklenmiyor.");
  const carcass = bounded(form.carcass_weight_kg, "carcass_weight_kg", .001, 1500);
  const live = form.live_weight_kg.trim() ? bounded(form.live_weight_kg, "live_weight_kg", .001, 2000) : undefined;
  if (live !== undefined && carcass > live) throw new Error("Karkas ağırlığı canlı ağırlığı aşamaz.");
  return { animal_type: form.animal_type, carcass_weight_kg: carcass,
    ...(live === undefined ? {} : { live_weight_kg: live }),
    ...(form.breed.trim() ? { breed: form.breed.trim() } : {}), sex: form.sex,
    ...(form.age_months.trim() ? { age_months: bounded(form.age_months, "age_months", 1, 120) } : {}),
    conformation_score: bounded(form.conformation_score, "conformation_score", 1, 5),
    fat_score: bounded(form.fat_score, "fat_score", 1, 5),
  };
}
export function modelMetrics(report: MLReport | null, model: ModelInfo | undefined, target: string): TargetReport | undefined {
  if (!model) return undefined;
  const group = report?.model_reports?.[model.name];
  // Do not attach metrics from another dataset/model to a displayed prediction.
  return group?.model_version === model.model_version && group.dataset_version === model.dataset_version
    ? group.metrics[target] : undefined;
}
export function ensurePredictionVersion(report: MLReport, name: ModelName, response: { model_version: string; dataset_version: string }): void {
  const model = report.model_reports?.[name];
  if (!model || model.model_version !== response.model_version || model.dataset_version !== response.dataset_version) {
    throw new Error("Model veya eğitim verisi sürümü değişti. Raporu yenileyip tekrar tahmin alın.");
  }
}
export function warningText(warning: string): string {
  if (warning.startsWith("Model is trained on synthetic")) return "Model sentetik veriyle eğitildi; gerçek işletme kayıtlarıyla saha doğrulaması yapılmadı.";
  if (warning.startsWith("Part weights reconciled")) return "Parçalar ölçülmüş karkas ağırlığına dengelendi; satılabilir toplam ve verim bu parçalardan hesaplandı.";
  if (warning.startsWith("Missing breed/age/sex")) return "Eksik ırk, yaş veya cinsiyet için demo varsayımları kullanıldı.";
  if (warning.startsWith("Unspecified fields")) return "Demo varsayımı kullanılan alanlar: " + warning.split(": ").slice(1).join(": ").replace(/\.$/, "").split(", ").map((field) => FEATURE_LABELS[field] ?? field).join(", ") + ".";
  const range = warning.match(/^(.+?)=(.+?) is outside the training range \[(.+?), (.+?)\]/);
  if (range) return `${FEATURE_LABELS[range[1]] ?? range[1]}: ${range[2]} eğitim aralığının (${range[3]}–${range[4]}) dışında. Tahminin güvenilirliği sınırlı.`;
  if (warning.includes("outside the training categories")) return "Seçilen kategori eğitim verisinde bulunmuyor. " + warning;
  return warning;
}

export interface OutputProduct { id: number; name: string; unit: string }
export function productionProduct(outputs: { product_lot_id: number }[], lots: { id: number; item_id: number; item_type: string; unit: string }[], products: OutputProduct[]): { product_type: ProductType; kg_per_unit: number } {
  const types = new Set<ProductType>();
  const factors = new Set<number>();
  if (!outputs.length) throw new Error("Mamul kabul kaydı yok.");
  for (const output of outputs) {
    const lot = lots.find((entry) => entry.id === output.product_lot_id);
    const product = products.find((entry) => entry.id === lot?.item_id);
    if (!lot || lot.item_type !== "product" || !product) throw new Error("Mamul kartı bulunamadı.");
    let factor = 1;
    if (lot.unit !== "kg" || product.unit !== "kg") {
      const pack = product.name.match(/\((\d+(?:[.,]\d+)?)\s*(kg|g)\)/i);
      if (lot.unit !== "paket" || product.unit !== "paket" || !pack) throw new Error("Mamulün kg dönüşümü için açık paket ağırlığı bulunamadı.");
      factor = Number(pack[1].replace(",", ".")) / (pack[2].toLowerCase() === "g" ? 1000 : 1);
      if (!(factor > 0)) throw new Error("Paket ağırlığı geçersiz.");
    }
    const name = product.name.toLocaleLowerCase("tr-TR");
    const matches = [name.includes("sucuk") && "sucuk", name.includes("salam") && "salam", (name.includes("sosis") || name.includes("sausage")) && "sausage"].filter(Boolean) as ProductType[];
    if (matches.length !== 1) throw new Error("Mamul tipi desteklenmiyor veya belirsiz.");
    types.add(matches[0]);
    factors.add(factor);
  }
  if (types.size !== 1) throw new Error("Aynı emirde birden fazla ürün tipi var; otomatik karşılaştırma yapılamıyor.");
  if (factors.size !== 1) throw new Error("Aynı emirde farklı paket ağırlıkları var; planın birimi belirsiz.");
  return { product_type: [...types][0], kg_per_unit: [...factors][0] };
}
