// Explicit catalog mapping: unsupported cards/units require configuration.
export interface MaterialCard { id: number; code: string; unit: string }
export interface InputLot { id: number; item_id: number; item_type: string; unit: string; unit_cost: number }
export interface OrderInput { material_lot_id: number; qty: number }

export function nonnegativeNumber(value: string, label: string): number {
  if (!value.trim()) throw new Error(`${label} gerekli.`);
  const parsed = Number(value);
  if (!Number.isFinite(parsed) || parsed < 0) throw new Error(`${label} geçerli, negatif olmayan bir sayı olmalı.`);
  return parsed;
}

export function productionFeatures(inputs: OrderInput[], lots: InputLot[], cards: MaterialCard[]) {
  const quantities = { raw_material_kg: 0, spice_kg: 0, casing_kg: 0, other_material_kg: 0 };
  const costs = { raw_material_kg: 0, spice_kg: 0, casing_kg: 0, other_material_kg: 0 };
  const mapping: Record<string, keyof typeof quantities> = {
    "MAT-001": "raw_material_kg", "MAT-002": "spice_kg", "MAT-003": "casing_kg", "MAT-004": "other_material_kg",
  };
  for (const input of inputs) {
    const lot = lots.find((candidate) => candidate.id === input.material_lot_id);
    const card = cards.find((candidate) => candidate.id === lot?.item_id);
    const field = card && mapping[card.code];
    if (!lot || lot.item_type !== "material" || !card || !field) throw new Error("Hammadde eşlemesi bulunamadı; tahmin yapılamıyor.");
    if (card.unit !== "kg" || lot.unit !== "kg") throw new Error("Tahmin için kg biriminde hammadde gerekli.");
    if (!Number.isFinite(input.qty) || input.qty < 0 || !Number.isFinite(lot.unit_cost) || lot.unit_cost < 0) throw new Error("Girdi miktarı veya fiyatı geçersiz.");
    quantities[field] += input.qty;
    costs[field] += input.qty * lot.unit_cost;
  }
  if (quantities.raw_material_kg <= 0) throw new Error("Et miktarı eksik veya sıfır.");
  return { ...quantities,
    material_unit_price_meat: costs.raw_material_kg / quantities.raw_material_kg,
    material_unit_price_spice: quantities.spice_kg ? costs.spice_kg / quantities.spice_kg : 0,
    material_unit_price_casing: quantities.casing_kg ? costs.casing_kg / quantities.casing_kg : 0,
    material_unit_price_other: quantities.other_material_kg ? costs.other_material_kg / quantities.other_material_kg : 0,
    product_type: "sucuk",
  };
}

export function predictionError(predicted: number | null, actual: number) {
  if (predicted === null) return { signed: null, absolute: null, percentage: null };
  const signed = predicted - actual;
  return { signed, absolute: Math.abs(signed), percentage: actual > 0 ? Math.abs(signed) / actual * 100 : null };
}

export function butcheryComparison(items: { product_id: number; weight: number; is_waste: boolean }[],
  products: { id: number; code: string }[], prediction: Record<string, number>) {
  const targets: Record<string, string> = { "PRD-101": "leg_weight", "PRD-102": "loin_weight",
    "PRD-103": "shoulder_weight", "PRD-104": "rib_weight", "PRD-105": "mince_weight", "PRD-106": "diced_meat_weight" };
  const actuals = new Map<string, number>();
  for (const item of items) {
    if (item.is_waste) continue;
    const code = products.find((p) => p.id === item.product_id)?.code;
    if (!code || (!targets[code] && code !== "PRD-107")) throw new Error("Parça eşlemesi eksik; karşılaştırma yapılamıyor.");
    if (!Number.isFinite(item.weight) || item.weight < 0) throw new Error("Kayıtlı parça miktarı geçersiz.");
    if (code === "PRD-107") continue;
    actuals.set(code, (actuals.get(code) ?? 0) + item.weight);
  }
  if (actuals.size === 0) throw new Error("Karşılaştırılabilir satılabilir parça kaydı yok.");
  const parts = [...actuals].map(([code, actual]) => {
    const predicted = prediction[targets[code]];
    if (!Number.isFinite(predicted)) throw new Error("Parça tahmini eksik.");
    return { code, actual, predicted };
  });
  return { parts, actual: parts.reduce((sum, part) => sum + part.actual, 0),
    predicted: parts.reduce((sum, part) => sum + part.predicted, 0) };
}
