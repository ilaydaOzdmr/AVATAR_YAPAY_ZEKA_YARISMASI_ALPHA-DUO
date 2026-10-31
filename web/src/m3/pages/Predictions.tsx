import { useEffect, useState, type FormEvent } from "react";
import { get, mlGet, mlPost } from "../../api";
import { nonnegativeNumber, productionFeatures, type InputLot, type MaterialCard, type OrderInput } from "../../predictionInputs";
import { ensurePredictionVersion, productionProduct, recordedPartTotals, type MLReport as VersionedReport } from "../../mlContracts";

interface PredictionMeta {
  model_version: string;
  dataset_version: string;
  model_status: string;
  warnings: string[];
  uncertainty?: Record<string, unknown>;
}

interface YieldResponse extends PredictionMeta {
  prediction: { finished_qty: number; yield_pct: number; estimated_ekp: number };
  cost_breakdown?: Record<string, number>;
  ekp_kind?: string;
  accounting_ekp?: number;
}

interface ButcheryResponse extends PredictionMeta {
  prediction: Record<string, number>;
  physical_check?: {
    carcass_weight_kg: number;
    parts_sum_kg: number;
    saleable_sum_kg: number;
    parts_over_carcass: boolean;
  };
}

interface ProductionOrder {
  id: number;
  code: string;
  expected_qty: number;
  service_unit_cost: number;
}

interface CostSummary {
  expected_qty: number;
  finished_qty: number;
  quantity_diff: number;
}

interface ForecastRow {
  code: string;
  predicted: number;
  expected: number;
  actual: number;
  diff: number;
}

interface ButcheryCompare {
  code: string;
  animal_type: string;
  input_weight: number;
  predicted: number;
  actual: number;
}

interface ModelMetrics {
  model_mae: number;
  model_rmse: number;
  model_r2?: number;
  baseline_mae: number;
  baseline_rmse: number;
}

interface MlReport extends VersionedReport {
  yield: ModelMetrics;
  butchery: ModelMetrics;
  butchery_lamb: ModelMetrics;
  butchery_cattle: ModelMetrics;
  note: string;
  model_version?: string;
  model_status?: string;
}

const PART_LABELS: Record<string, string> = {
  leg_weight: "But",
  loin_weight: "Sırt",
  shoulder_weight: "Omuz",
  rib_weight: "Kaburga",
  mince_weight: "Kıyma",
  diced_meat_weight: "Kuşbaşı",
  bone_weight: "Kemik",
  waste_weight: "Fire",
  total_saleable_weight: "Toplam satılabilir",
  cutting_yield_pct: "Verim %",
};

function WarningList({ items }: { items: string[] }) {
  if (items.length === 0) return null;
  return (
    <ul className="mt-2 space-y-1 text-sm" style={{ paddingLeft: 18, color: "#8A5A1B" }}>
      {items.map((item) => (
        <li key={item}>{item}</li>
      ))}
    </ul>
  );
}

function ForecastChart({ rows }: { rows: ForecastRow[] }) {
  const width = Math.max(420, rows.length * 110);
  const height = 190;
  const max = Math.max(...rows.flatMap((row) => [row.expected, row.actual, row.predicted]), 1);
  const barWidth = 20;
  return (
    <svg viewBox={`0 0 ${width} ${height}`} role="img" aria-label="Model tahmini, plan ve gerçek karşılaştırması" style={{ width: "100%", height: "auto" }}>
      <line x1="30" y1={height - 30} x2={width - 10} y2={height - 30} stroke="#c8d1ca" />
      {rows.map((row, index) => {
        const groupX = 30 + index * 110;
        const predictedHeight = ((height - 60) * row.predicted) / max;
        const expectedHeight = ((height - 60) * row.expected) / max;
        const actualHeight = ((height - 60) * row.actual) / max;
        return (
          <g key={row.code}>
            <rect x={groupX} y={height - 30 - predictedHeight} width={barWidth} height={predictedHeight} fill="#3B6EA5" />
            <rect x={groupX + barWidth + 3} y={height - 30 - expectedHeight} width={barWidth} height={expectedHeight} fill="#B97722" />
            <rect x={groupX + (barWidth + 3) * 2} y={height - 30 - actualHeight} width={barWidth} height={actualHeight} fill="#2C7058" />
            <text x={groupX} y={height - 12} fontSize="9" fill="#59645c">{row.code.slice(-6)}</text>
          </g>
        );
      })}
      <text x={30} y={14} fontSize="11" fill="#59645c">■ Model (mavi) · ■ Plan (turuncu) · ■ Gerçek (yeşil)</text>
    </svg>
  );
}

export default function Predictions() {
  const [rawKg, setRawKg] = useState("100");
  const [spiceKg, setSpiceKg] = useState("2");
  const [casingKg, setCasingKg] = useState("3");
  const [serviceCost, setServiceCost] = useState("400");
  const [laborCost, setLaborCost] = useState("80");
  const [yieldResult, setYieldResult] = useState<YieldResponse | null>(null);

  const [animalType, setAnimalType] = useState("lamb");
  const [liveWeight, setLiveWeight] = useState("45");
  const [carcassWeight, setCarcassWeight] = useState("21");
  const [butcheryResult, setButcheryResult] = useState<ButcheryResponse | null>(null);

  const [sourceReport, setReport] = useState<MlReport | null>(null);
  const report = sourceReport && { ...sourceReport,
    butchery: animalType === "cattle" ? sourceReport.butchery_cattle : sourceReport.butchery_lamb };
  const [forecast, setForecast] = useState<ForecastRow[]>([]);
  const [butcheryCompare, setButcheryCompare] = useState<ButcheryCompare[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const reportRequest = mlGet<MlReport>("/report");
    reportRequest.then(setReport).catch((err: unknown) => setError(String(err)));
    const catalog = Promise.all([get<InputLot[]>("/lots"), get<MaterialCard[]>("/materials"),
      get<{ id: number; code: string; name: string; unit: string }[]>("/products")]);
    Promise.all([get<ProductionOrder[]>("/production-orders"), catalog]).then(async ([orders, [lots, cards, products]]) => {
      const rows = await Promise.all(
        (orders ?? []).map(async (order) => {
          try {
            const [summary, inputs, outputs] = await Promise.all([
              get<CostSummary>(`/production-orders/${order.id}`),
              get<OrderInput[]>(`/production-orders/${order.id}/inputs`),
              get<{ product_lot_id: number }[]>(`/production-orders/${order.id}/outputs`),
            ]);
            if (summary.finished_qty <= 0) return null;
            const context = productionProduct(outputs ?? [], lots ?? [], products ?? []);
            const features = productionFeatures(inputs ?? [], lots ?? [], cards ?? []);
            const ml = await mlPost<YieldResponse>("/predict/v2/yield", { ...features,
              product_type: context.product_type,
              service_unit_cost: order.service_unit_cost / context.kg_per_unit });
            ensurePredictionVersion(await reportRequest, "yield", ml);
            const predicted = ml.prediction.finished_qty;
            return {
              code: order.code,
              predicted,
              expected: order.expected_qty * context.kg_per_unit,
              actual: summary.finished_qty * context.kg_per_unit,
              diff: predicted - summary.finished_qty * context.kg_per_unit,
            };
          } catch (err) {
            setError(err instanceof Error ? err.message : String(err));
            return null;
          }
        }),
      );
      setForecast(rows.filter((row): row is ForecastRow => row !== null));
    }).catch((err: unknown) => setError(String(err)));

    Promise.all([get<{ id: number; code: string; animal_type: string; input_weight: number }[]>("/butchery-batches"), catalog]).then(async ([batches, [, , products]]) => {
      const rows = await Promise.all(
        (batches ?? []).map(async (batch) => {
          try {
            const aliases: Record<string, "lamb" | "cattle"> = { kuzu: "lamb", dana: "cattle", lamb: "lamb", cattle: "cattle" };
            const animal = aliases[batch.animal_type.toLowerCase()];
            if (!animal) throw new Error("Unknown animal type");
            const [result, prediction] = await Promise.all([
              get<{ items: { product_id: number; weight: number; is_waste: boolean }[] }>(`/butchery-batches/${batch.id}`),
              mlPost<ButcheryResponse>("/predict/v2/butchery", {
                animal_type: animal,
                carcass_weight_kg: batch.input_weight,
              }),
            ]);
            ensurePredictionVersion(await reportRequest, animal === "cattle" ? "butchery_cattle" : "butchery_lamb", prediction);
            const { actual, predicted } = recordedPartTotals(result.items ?? [], products ?? [], prediction.prediction);
            return {
              code: batch.code,
              animal_type: batch.animal_type,
              input_weight: batch.input_weight,
              predicted,
              actual,
            };
          } catch (err) {
            setError(err instanceof Error ? err.message : String(err));
            return null;
          }
        }),
      );
      setButcheryCompare(rows.filter((row): row is ButcheryCompare => row !== null));
    }).catch((err: unknown) => setError(String(err)));
  }, []);

  async function runYield() {
    setError(null);
    setYieldResult(null);
    try {
      const result = await mlPost<YieldResponse>("/predict/v2/yield", {
        raw_material_kg: nonnegativeNumber(rawKg, "Hammadde kg"),
        spice_kg: nonnegativeNumber(spiceKg, "Baharat kg"),
        casing_kg: nonnegativeNumber(casingKg, "Bağırsak kg"),
        service_cost: nonnegativeNumber(serviceCost, "Hizmet"),
        labor_cost: nonnegativeNumber(laborCost, "İşçilik"),
        product_type: "sucuk",
      });
      ensurePredictionVersion(sourceReport ?? await mlGet<MlReport>("/report"), "yield", result);
      setYieldResult(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  async function runButchery() {
    setError(null);
    setButcheryResult(null);
    try {
      const result = await mlPost<ButcheryResponse>("/predict/v2/butchery", {
        animal_type: animalType,
        live_weight_kg: liveWeight.trim() ? nonnegativeNumber(liveWeight, "Canlı kg") : null,
        carcass_weight_kg: nonnegativeNumber(carcassWeight, "Karkas kg"),
      });
      ensurePredictionVersion(sourceReport ?? await mlGet<MlReport>("/report"), animalType === "cattle" ? "butchery_cattle" : "butchery_lamb", result);
      setButcheryResult(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  function predictYield(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    void runYield();
  }

  function predictButchery(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    void runButchery();
  }

  async function refreshAll() {
    setError(null);
    await Promise.all([runYield(), runButchery()]);
    mlGet<MlReport>("/report").then(setReport).catch(() => {});
  }

  function metricsTable(metrics: ModelMetrics) {
    return (
      <table>
        <thead>
          <tr><th>Metrik</th><th>Model</th><th>Baseline</th></tr>
        </thead>
        <tbody>
          <tr><td>MAE</td><td>{metrics.model_mae}</td><td>{metrics.baseline_mae}</td></tr>
          <tr><td>RMSE</td><td>{metrics.model_rmse}</td><td>{metrics.baseline_rmse}</td></tr>
          <tr><td>R²</td><td>{metrics.model_r2 ?? "—"}</td><td>—</td></tr>
        </tbody>
      </table>
    );
  }

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">KARAR DESTEK SİSTEMİ</div>
          <h1>Yapay Zekâ Merkezi</h1>
          <p>Tahminler Go backend üzerinden alınır. Model durumu ve belirsizlik her cevapta gösterilir.</p>
        </div>
        <button className="btn btn-primary" type="button" onClick={refreshAll}>↻ Tahminleri güncelle</button>
      </div>

      {error && <p className="muted" style={{ marginBottom: 12 }}>{error}</p>}

      <div className="two-col">
        <div className="panel">
          <div className="panel-head">
            <div>
              <h2>Üretim verimi &amp; EKP tahmini</h2>
              <p>Model A · Üretim öncesi tahmin</p>
            </div>
          </div>
          <div style={{ padding: 20 }}>
            <form className="inline-form" onSubmit={predictYield} style={{ flexWrap: "wrap", gap: 8 }}>
              <input value={rawKg} onChange={(e) => setRawKg(e.target.value)} placeholder="Hammadde kg" style={{ width: 120 }} />
              <input value={spiceKg} onChange={(e) => setSpiceKg(e.target.value)} placeholder="Baharat kg" style={{ width: 110 }} />
              <input value={casingKg} onChange={(e) => setCasingKg(e.target.value)} placeholder="Bağırsak kg" style={{ width: 115 }} />
              <input value={serviceCost} onChange={(e) => setServiceCost(e.target.value)} placeholder="Hizmet €" style={{ width: 100 }} />
              <input value={laborCost} onChange={(e) => setLaborCost(e.target.value)} placeholder="İşçilik €" style={{ width: 100 }} />
              <button className="btn btn-secondary" type="submit">Tahmin et</button>
            </form>
            {yieldResult && (
              <>
                <div className="metrics" style={{ gridTemplateColumns: "1fr 1fr 1fr", marginTop: 16, marginBottom: 0 }}>
                  <div className="metric">
                    <div className="metric-top">Beklenen mamul</div>
                    <strong>{yieldResult.prediction.finished_qty} kg</strong>
                  </div>
                  <div className="metric">
                    <div className="metric-top">Verim</div>
                    <strong>{(yieldResult.prediction.yield_pct * 100).toFixed(1)}%</strong>
                  </div>
                  <div className="metric">
                    <div className="metric-top">Tahmini EKP</div>
                    <strong style={{ color: "var(--green-3)" }}>€ {yieldResult.prediction.estimated_ekp.toFixed(3)}</strong>
                  </div>
                </div>
                <p className="muted" style={{ marginTop: 10 }}>
                  Model sürümü: <b>{yieldResult.model_version}</b> · Durum: <b>{yieldResult.model_status}</b>
                  {yieldResult.cost_breakdown ? ` · Toplam maliyet: € ${yieldResult.cost_breakdown.total_cost}` : ""}
                  {yieldResult.ekp_kind ? ` · EKP türü: ${yieldResult.ekp_kind}` : ""}
                </p>
                <WarningList items={yieldResult.warnings} />
              </>
            )}
          </div>
        </div>

        <div className="panel">
          <div className="panel-head">
            <div>
              <h2>Parçalama tahmini</h2>
              <p>Model B · Kuzu ve sığır ayrı modeller</p>
            </div>
          </div>
          <div style={{ padding: 20 }}>
            <form className="inline-form" onSubmit={predictButchery} style={{ flexWrap: "wrap", gap: 8 }}>
              <select value={animalType} onChange={(e) => setAnimalType(e.target.value)} style={{ width: 110 }}>
                <option value="lamb">Kuzu</option>
                <option value="cattle">Sığır</option>
              </select>
              <input value={liveWeight} onChange={(e) => setLiveWeight(e.target.value)} placeholder="Canlı kg" style={{ width: 110 }} />
              <input value={carcassWeight} onChange={(e) => setCarcassWeight(e.target.value)} placeholder="Karkas kg (ops.)" style={{ width: 135 }} />
              <button className="btn btn-secondary" type="submit">Tahmin et</button>
            </form>
            {butcheryResult && (
              <>
                <div className="table-wrap" style={{ marginTop: 12 }}>
                  <table>
                    <thead>
                      <tr><th>Parça</th><th>Tahmini değer</th></tr>
                    </thead>
                    <tbody>
                      {Object.entries(butcheryResult.prediction).map(([key, value]) => (
                        <tr key={key}>
                          <td>{PART_LABELS[key] ?? key}</td>
                          <td>{key === "cutting_yield_pct" ? `${value.toFixed(2)}%` : `${value.toFixed(3)} kg`}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                {butcheryResult.physical_check && (
                  <p className="muted" style={{ marginTop: 10 }}>
                    Karkas: {butcheryResult.physical_check.carcass_weight_kg} kg · Parça toplamı:{" "}
                    {butcheryResult.physical_check.parts_sum_kg} kg ·{" "}
                    {butcheryResult.physical_check.parts_over_carcass ? "Uyarı: parça toplamı karkası aşıyor." : "Fiziksel kontrol tutarlı."}
                  </p>
                )}
                <p className="muted">
                  Model sürümü: <b>{butcheryResult.model_version}</b> · Durum: <b>{butcheryResult.model_status}</b>
                </p>
                <WarningList items={butcheryResult.warnings} />
              </>
            )}
          </div>
        </div>
      </div>

      {forecast.length > 0 ? (
        <div className="panel" style={{ marginTop: 16 }}>
          <div className="panel-head">
            <div>
              <h2>Model tahmini vs Gerçek</h2>
              <p>Üretim emirleri: ML model tahmini, plan ve gerçekleşen miktar karşılaştırması. Grafik yalnızca gerçek kayıt varsa gösterilir.</p>
            </div>
          </div>
          <div style={{ padding: 16 }}>
            <ForecastChart rows={forecast} />
            <div className="table-wrap" style={{ marginTop: 12 }}>
              <table>
                <thead>
                  <tr><th>Üretim emri</th><th>Model tahmini</th><th>Plan</th><th>Gerçekleşen</th><th>Sapma</th></tr>
                </thead>
                <tbody>
                  {forecast.map((row) => (
                    <tr key={row.code}>
                      <td className="id">{row.code}</td>
                      <td>{row.predicted ? row.predicted.toFixed(1) : "—"}</td>
                      <td>{row.expected}</td>
                      <td>{row.actual}</td>
                      <td style={{ color: row.predicted > 0 && Math.abs(row.diff) > 0.5 ? "var(--amber)" : "var(--green-3)" }}>
                        {row.predicted ? `${row.diff >= 0 ? "+" : ""}${row.diff.toFixed(1)}` : "—"}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      ) : (
        <div className="panel" style={{ marginTop: 16, padding: 20 }}>
          <p className="muted">Gerçek üretim kaydı bulunamadı; tahmin–gerçek grafiği gösterilmiyor.</p>
        </div>
      )}

      {butcheryCompare.length > 0 && (
        <div className="panel" style={{ marginTop: 16 }}>
          <div className="panel-head">
            <div>
              <h2>Parçalama tahmini vs Gerçek</h2>
              <p>Kuzu/dana parçalama: ML tahmini toplam parça ağırlığı ve gerçek parça ağırlığı karşılaştırması</p>
            </div>
          </div>
          <div className="table-wrap">
            <table>
              <thead>
                <tr><th>Parti</th><th>Hayvan</th><th>Giriş</th><th>Tahmini (kg)</th><th>Gerçek (kg)</th><th>Sapma</th></tr>
              </thead>
              <tbody>
                {butcheryCompare.map((row) => (
                  <tr key={row.code}>
                    <td className="id">{row.code}</td>
                    <td>{row.animal_type}</td>
                    <td>{row.input_weight}</td>
                    <td>{row.predicted.toFixed(2)}</td>
                    <td>{row.actual.toFixed(2)}</td>
                    <td style={{ color: Math.abs(row.predicted - row.actual) > 1 ? "var(--amber)" : "var(--green-3)" }}>
                      {(row.predicted - row.actual).toFixed(2)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {report && (
        <div className="two-col" style={{ marginTop: 16 }}>
          <div className="panel">
            <div className="panel-head">
              <h2>Üretim modeli doğruluk</h2>
            </div>
            <div className="table-wrap">{metricsTable(report.yield)}</div>
          </div>
          <div className="panel">
            <div className="panel-head">
              <h2>Parçalama modeli doğruluk</h2>
            </div>
            <div className="table-wrap">{metricsTable(report.butchery)}</div>
          </div>
          <div className="panel" style={{ gridColumn: "1 / -1" }}>
            <div className="alert-list" style={{ padding: "14px 20px" }}>
              <div className="alert" style={{ borderBottom: 0 }}>
                <div className="alert-symbol amber">↗</div>
                <div>
                  <b>Model durumu: {report.model_status ?? "synthetic_only"} · Sürüm {report.model_version ?? "—"}</b>
                  <p>{report.note}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </section>
  );
}
