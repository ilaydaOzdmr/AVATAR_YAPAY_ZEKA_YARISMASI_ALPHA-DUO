import { useEffect, useState, type FormEvent } from "react";
import { get, mlGet, mlPost } from "../../api";

interface YieldResult {
  predicted_qty: number;
  predicted_unit_cost: number;
}

interface ProductionOrder {
  id: number;
  code: string;
  expected_qty: number;
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

interface ButcheryPart {
  product_code: string;
  weight: number;
}

interface ModelMetrics {
  model_mae: number;
  model_rmse: number;
  baseline_mae: number;
  baseline_rmse: number;
}

interface MlReport {
  yield: ModelMetrics;
  butchery: ModelMetrics;
  note: string;
}

export default function Predictions() {
  const [rawKg, setRawKg] = useState("100");
  const [yieldResult, setYieldResult] = useState<YieldResult | null>(null);

  const [animalType, setAnimalType] = useState("kuzu");
  const [inputWeight, setInputWeight] = useState("25");
  const [parts, setParts] = useState<ButcheryPart[]>([]);

  const [report, setReport] = useState<MlReport | null>(null);
  const [forecast, setForecast] = useState<ForecastRow[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    mlGet<MlReport>("/report").then(setReport).catch(() => {});
    get<ProductionOrder[]>("/production-orders").then(async (orders) => {
      const rows = await Promise.all(
        orders.map(async (order) => {
          try {
            const [summary, inputs] = await Promise.all([
              get<CostSummary>(`/production-orders/${order.id}`),
              get<{ qty: number }[]>(`/production-orders/${order.id}/inputs`).catch(() => []),
            ]);
            const rawKg = inputs.reduce((sum, input) => sum + input.qty, 0);
            let predicted = 0;
            if (rawKg > 0) {
              try {
                const ml = await mlPost<YieldResult>("/predict/yield", { raw_material_kg: rawKg });
                predicted = ml.predicted_qty;
              } catch {
                predicted = 0;
              }
            }
            return {
              code: order.code,
              predicted,
              expected: summary.expected_qty,
              actual: summary.finished_qty,
              diff: predicted - summary.finished_qty,
            };
          } catch {
            return null;
          }
        }),
      );
      setForecast(rows.filter((row): row is ForecastRow => row !== null));
    });
  }, []);

  async function runYield() {
    setError(null);
    try {
      setYieldResult(
        await mlPost<YieldResult>("/predict/yield", {
          raw_material_kg: Number.parseFloat(rawKg),
        }),
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  async function runButchery() {
    setError(null);
    try {
      const result = await mlPost<{ parts: ButcheryPart[] }>("/predict/butchery", {
        animal_type: animalType.trim(),
        input_weight: Number.parseFloat(inputWeight),
      });
      setParts(result.parts);
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
          <tr>
            <th>Metrik</th>
            <th>Model</th>
            <th>Referans</th>
          </tr>
        </thead>
        <tbody>
          <tr><td>MAE</td><td>{metrics.model_mae}</td><td>{metrics.baseline_mae}</td></tr>
          <tr><td>RMSE</td><td>{metrics.model_rmse}</td><td>{metrics.baseline_rmse}</td></tr>
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
          <p>Tahminleri gerçek sonuçlarla karşılaştırın. Operasyon kararlarını kesin kayıtlarla verin.</p>
        </div>
        <button className="btn btn-primary" type="button" onClick={refreshAll}>↻ Tahminleri güncelle</button>
      </div>

      {error && <p className="muted" style={{ marginBottom: 12 }}>{error}</p>}

      <div className="two-col">
        <div className="panel">
          <div className="panel-head">
            <div>
              <h2>Üretim verimi &amp; EKP tahmini</h2>
              <p>Model A · Üretim emirleri</p>
            </div>
          </div>
          <div style={{ padding: 20 }}>
            <form className="inline-form" onSubmit={predictYield}>
              <input value={rawKg} onChange={(e) => setRawKg(e.target.value)} placeholder="Hammadde kg" style={{ width: 140 }} />
              <button className="btn btn-secondary" type="submit">Tahmin et</button>
            </form>
            {yieldResult && (
              <div className="metrics" style={{ gridTemplateColumns: "1fr 1fr", marginTop: 16, marginBottom: 0 }}>
                <div className="metric">
                  <div className="metric-top">Beklenen mamul miktarı</div>
                  <strong>{yieldResult.predicted_qty} paket</strong>
                </div>
                <div className="metric">
                  <div className="metric-top">Tahmini EKP</div>
                  <strong style={{ color: "var(--green-3)" }}>€ {yieldResult.predicted_unit_cost.toFixed(3)}</strong>
                </div>
              </div>
            )}
          </div>
        </div>

        <div className="panel">
          <div className="panel-head">
            <div>
              <h2>Kuzu parçalama verimi</h2>
              <p>Model B · Giriş özellikleri ve parça ağırlıkları</p>
            </div>
          </div>
          <div style={{ padding: 20 }}>
            <form className="inline-form" onSubmit={predictButchery}>
              <input value={animalType} onChange={(e) => setAnimalType(e.target.value)} placeholder="Hayvan türü" style={{ width: 110 }} />
              <input value={inputWeight} onChange={(e) => setInputWeight(e.target.value)} placeholder="Giriş ağırlığı" style={{ width: 120 }} />
              <button className="btn btn-secondary" type="submit">Tahmin et</button>
            </form>
            {parts.length > 0 && (
              <div className="table-wrap" style={{ marginTop: 12 }}>
                <table>
                  <thead>
                    <tr><th>Ürün kodu</th><th>Tahmini ağırlık</th></tr>
                  </thead>
                  <tbody>
                    {parts.map((part) => (
                      <tr key={part.product_code}>
                        <td className="id">{part.product_code}</td>
                        <td>{part.weight} kg</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      </div>

      {forecast.length > 0 && (
        <div className="panel" style={{ marginTop: 16 }}>
          <div className="panel-head">
            <div>
              <h2>Model tahmini vs Gerçek</h2>
              <p>Üretim emirleri: ML model tahmini, plan ve gerçekleşen miktar karşılaştırması</p>
            </div>
          </div>
          <div className="table-wrap">
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
                  <b>Model notu</b>
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
