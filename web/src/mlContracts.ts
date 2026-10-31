// Data adapters only. The existing pages own all presentation.
export type ModelName = "yield" | "butchery_lamb" | "butchery_cattle";
export type ProductType = "sucuk" | "salam" | "sausage";
export interface MLReport {
  model_reports: Partial<Record<ModelName, { model_version: string; dataset_version: string }>>;
}
export interface PredictionMeta {
  model_version: string; dataset_version: string; model_status: string; data_origin: string; warnings: string[];
}
export interface YieldResponse extends PredictionMeta {
  prediction: { finished_qty: number; yield_pct: number; estimated_ekp: number };
}
export interface ButcheryResponse extends PredictionMeta { prediction: Record<string, number> }

export function ensurePredictionVersion(report: MLReport, name: ModelName, response: { model_version: string; dataset_version: string }): void {
  const model = report.model_reports?.[name];
  if (!model || model.model_version !== response.model_version || model.dataset_version !== response.dataset_version) {
    throw new Error("Model veya eğitim verisi sürümü değişti. Raporu yenileyip tekrar tahmin alın.");
  }
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

// Match exactly the parts that were weighed, including bone or recorded waste.
// Never compare a whole-carcass sum with only saleable meat or invent missing cuts.
export function recordedPartTotals(items: { product_id: number; weight: number; is_waste: boolean }[],
  products: { id: number; code: string }[], prediction: Record<string, number>) {
  const targets: Record<string, string> = { "PRD-101": "leg_weight", "PRD-102": "loin_weight",
    "PRD-103": "shoulder_weight", "PRD-104": "rib_weight", "PRD-105": "mince_weight",
    "PRD-106": "diced_meat_weight", "PRD-107": "bone_weight" };
  const measured = new Map<string, number>();
  for (const item of items) {
    const code = products.find((product) => product.id === item.product_id)?.code;
    const target = item.is_waste ? "waste_weight" : code && targets[code];
    if (!target || !Number.isFinite(item.weight) || item.weight < 0) throw new Error("Recorded part or unit mapping is invalid");
    measured.set(target, (measured.get(target) ?? 0) + item.weight);
  }
  if (!measured.size) throw new Error("No measured parts available");
  let actual = 0, predicted = 0;
  for (const [target, weight] of measured) {
    if (!Number.isFinite(prediction[target]) || prediction[target] < 0) throw new Error("Part prediction is missing");
    actual += weight;
    predicted += prediction[target];
  }
  return { actual, predicted };
}
