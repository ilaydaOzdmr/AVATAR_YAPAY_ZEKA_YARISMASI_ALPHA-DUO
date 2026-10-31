import { useCallback, useEffect, useState } from "react";
import { get, mlGet } from "../../api";
import { downloadCsv } from "../csv";

interface Lot {
  id: number;
  code: string;
  item_type: string;
}

interface Movement {
  movement_type: string;
  qty: number;
  occurred_at: string;
}

interface ProductionOrder {
  id: number;
  code: string;
}

interface CostSummary {
  ekp: number;
  total_cost: number;
  finished_qty: number;
}

interface Metrics {
  model_mae: number;
  model_rmse: number;
  baseline_mae: number;
  baseline_rmse: number;
}

interface MlReport {
  yield: Metrics;
  butchery: Metrics;
  note: string;
}

interface Material {
  id: number;
  name: string;
}

interface ButcheryBatch {
  id: number;
  code: string;
  animal_type: string;
  input_weight: number;
  total_cost: number;
}

interface Sale {
  id: number;
  code: string;
  market_id: number;
  sold_at: string;
  status: string;
}

interface Market {
  id: number;
  name: string;
}

interface ReportData {
  title: string;
  columns: string[];
  rows: (string | number)[][];
}

export default function Reports({ onNavigate }: { onNavigate: (view: string) => void }) {
  const [productionCount, setProductionCount] = useState(0);
  const [saleCount, setSaleCount] = useState(0);
  const [lotCount, setLotCount] = useState(0);
  const [materialCount, setMaterialCount] = useState(0);
  const [productCount, setProductCount] = useState(0);
  const [movementCount, setMovementCount] = useState(0);
  const [anomalyCount, setAnomalyCount] = useState(0);
  const [preview, setPreview] = useState<ReportData | null>(null);
  const [feedback, setFeedback] = useState<string | null>(null);

  const load = useCallback(() => {
    get<unknown[]>("/production-orders").then((r) => setProductionCount(r.length)).catch(() => {});
    get<unknown[]>("/sales").then((r) => setSaleCount(r.length)).catch(() => {});
    get<unknown[]>("/products").then((r) => setProductCount(r.length)).catch(() => {});
    get<unknown[]>("/anomalies").then((r) => setAnomalyCount(r.length)).catch(() => {});
    get<Lot[]>("/lots").then((lots) => {
      setLotCount(lots.length);
      setMaterialCount(lots.filter((l) => l.item_type === "material").length);
      Promise.all(
        lots.map((lot) =>
          get<unknown[]>(`/lots/${lot.id}/movements`)
            .then((m) => m.length)
            .catch(() => 0),
        ),
      ).then((counts) => setMovementCount(counts.reduce((a, b) => a + b, 0)));
    }).catch(() => {});
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  function showReport(data: ReportData) {
    setFeedback(null);
    setPreview(data);
  }

  async function generateCostReport(): Promise<ReportData> {
    const orders = await get<ProductionOrder[]>("/production-orders");
    const rows: (string | number)[][] = [];
    for (const order of orders) {
      try {
        const cost = await get<CostSummary>(`/production-orders/${order.id}`);
        rows.push([order.code, `€ ${cost.ekp.toFixed(2)}`, `€ ${cost.total_cost.toFixed(2)}`, cost.finished_qty]);
      } catch {
        rows.push([order.code, "—", "—", "—"]);
      }
    }
    return { title: "Gerçek Üretim Maliyeti Raporu", columns: ["Üretim emri", "EKP", "Toplam maliyet", "Mamul miktarı"], rows };
  }

  async function generatePurchaseReport(): Promise<ReportData> {
    const [lots, materials] = await Promise.all([
      get<{ id: number; code: string; item_id: number; qty: number; unit_cost: number }[]>("/lots?item_type=material"),
      get<Material[]>("/materials"),
    ]);
    const nameById = new Map(materials.map((m) => [m.id, m.name]));
    const rows: (string | number)[][] = lots.map((lot) => [
      lot.code,
      nameById.get(lot.item_id) ?? "—",
      `${lot.qty} kg`,
      `€ ${lot.unit_cost.toFixed(2)}`,
      `€ ${(lot.qty * lot.unit_cost).toFixed(2)}`,
    ]);
    return { title: "Alış Raporu", columns: ["Lot", "Hammadde", "Miktar", "Birim maliyet", "Toplam"], rows };
  }

  async function generateButcheryReport(): Promise<ReportData> {
    const batches = await get<ButcheryBatch[]>("/butchery-batches");
    const rows: (string | number)[][] = batches.map((batch) => [
      batch.code,
      batch.animal_type,
      `${batch.input_weight} kg`,
      `€ ${batch.total_cost.toFixed(2)}`,
    ]);
    return { title: "Parçalama Raporu", columns: ["Parti", "Hayvan", "Giriş ağırlığı", "Toplam maliyet"], rows };
  }

  async function generateSalesReport(): Promise<ReportData> {
    const [sales, markets] = await Promise.all([get<Sale[]>("/sales"), get<Market[]>("/markets")]);
    const nameById = new Map(markets.map((m) => [m.id, m.name]));
    const rows: (string | number)[][] = sales.map((sale) => [
      sale.code,
      nameById.get(sale.market_id) ?? `Market #${sale.market_id}`,
      sale.sold_at,
      sale.status,
    ]);
    return { title: "Satış Raporu", columns: ["Satış", "Market", "Tarih", "Durum"], rows };
  }

  async function generateLotHistory(): Promise<ReportData> {
    const lots = await get<Lot[]>("/lots");
    const rows: (string | number)[][] = [];
    for (const lot of lots) {
      try {
        const movements = await get<Movement[]>(`/lots/${lot.id}/movements`);
        for (const movement of movements) {
          rows.push([lot.code, movement.movement_type, movement.qty, movement.occurred_at]);
        }
      } catch {
        /* skip */
      }
    }
    return { title: "Lot Hareket Geçmişi", columns: ["Lot", "Hareket türü", "Miktar", "Tarih"], rows };
  }

  async function generateAiReport(): Promise<ReportData> {
    const report = await mlGet<MlReport>("/report");
    const rows: (string | number)[][] = [
      ["Üretim MAE", report.yield.model_mae, report.yield.baseline_mae],
      ["Üretim RMSE", report.yield.model_rmse, report.yield.baseline_rmse],
      ["Parçalama MAE", report.butchery.model_mae, report.butchery.baseline_mae],
      ["Parçalama RMSE", report.butchery.model_rmse, report.butchery.baseline_rmse],
    ];
    return { title: "AI Tahmin Hata Raporu", columns: ["Metrik", "Model", "Referans"], rows };
  }

  async function run(action: () => Promise<ReportData>) {
    setFeedback(null);
    try {
      showReport(await action());
    } catch (err) {
      setFeedback(err instanceof Error ? err.message : String(err));
    }
  }

  function downloadPdf() {
    window.print();
  }

  function downloadCsvCurrent() {
    if (!preview) return;
    downloadCsv(
      "rapor.csv",
      preview.columns,
      preview.rows.map((row) => row.map((cell) => String(cell).replace(/^€ /, ""))),
    );
  }

  const catalog: { name: string; scope: string; action: () => void }[] = [
    { name: "Gerçek üretim maliyeti", scope: "Hammadde + hizmet + EKP", action: () => run(generateCostReport) },
    { name: "Alış raporu", scope: "Hammadde lotları ve maliyet", action: () => run(generatePurchaseReport) },
    { name: "Parçalama raporu", scope: "Kuzu parçalama kayıtları", action: () => run(generateButcheryReport) },
    { name: "Satış raporu", scope: "Satış ve sevkiyat kayıtları", action: () => run(generateSalesReport) },
    { name: "Lot hareket geçmişi", scope: "Giriş, transfer, üretim, satış", action: () => run(generateLotHistory) },
    { name: "Geri çağırma alıcı listesi", scope: "Market ve miktar bazında", action: () => onNavigate("recall") },
    { name: "AI tahmin hata raporu", scope: "Gerçek / tahmin / MAE / RMSE", action: () => run(generateAiReport) },
  ];

  const today = new Date().toLocaleDateString("tr-TR", { day: "numeric", month: "long", year: "numeric" });

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">İŞLETME ANALİTİĞİ</div>
          <h1>Raporlar</h1>
          <p>Alış, üretim, parçalama, lot hareketleri, satış ve geri çağırma kayıtları.</p>
        </div>
        <button className="btn btn-primary" type="button" onClick={load}>↻ Yenile</button>
      </div>

      <div className="metrics">
        <div className="metric">
          <div className="metric-top">Üretim emri <span className="metric-icon">⚙</span></div>
          <strong>{productionCount}</strong>
          <div className="metric-foot">Kayıtlı emir</div>
        </div>
        <div className="metric">
          <div className="metric-top">Stok hareketi <span className="metric-icon">↔</span></div>
          <strong>{movementCount}</strong>
          <div className="metric-foot">Giriş, çıkış ve transfer</div>
        </div>
        <div className="metric">
          <div className="metric-top">Satış <span className="metric-icon">↗</span></div>
          <strong>{saleCount}</strong>
          <div className="metric-foot">Lot bazında kayıtlı</div>
        </div>
        <div className="metric">
          <div className="metric-top">Kalite olayı <span className="metric-icon" style={{ background: "var(--amber-bg)", color: "var(--amber)" }}>⚠</span></div>
          <strong>{anomalyCount}</strong>
          <div className="metric-foot">Takipte</div>
        </div>
      </div>

      <div className="metrics">
        <div className="metric">
          <div className="metric-top">Toplam lot <span className="metric-icon">▤</span></div>
          <strong>{lotCount}</strong>
          <div className="metric-foot">{materialCount} hammadde</div>
        </div>
        <div className="metric">
          <div className="metric-top">Ürün kartı <span className="metric-icon">▣</span></div>
          <strong>{productCount}</strong>
          <div className="metric-foot">Kayıtlı ürün</div>
        </div>
      </div>

      {feedback && <p className="muted" style={{ marginBottom: 12 }}>{feedback}</p>}

      <div className="panel">
        <div className="panel-head">
          <h2>Rapor kataloğu</h2>
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr><th>Rapor</th><th>Kapsam</th><th>Format</th><th></th></tr>
            </thead>
            <tbody>
              {catalog.map((report) => (
                <tr key={report.name}>
                  <td><b>{report.name}</b></td>
                  <td>{report.scope}</td>
                  <td>PDF / CSV</td>
                  <td><button className="btn btn-secondary" type="button" onClick={report.action}>Oluştur</button></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {preview && (
        <>
          <div className="panel" style={{ marginTop: 16 }}>
            <div className="panel-head">
              <h2>Rapor önizleme</h2>
              <div style={{ display: "flex", gap: 8 }}>
                <button className="btn btn-primary" type="button" onClick={downloadPdf}>PDF indir</button>
                <button className="btn btn-secondary" type="button" onClick={downloadCsvCurrent}>CSV indir</button>
              </div>
            </div>
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>{preview.columns.map((column) => <th key={column}>{column}</th>)}</tr>
                </thead>
                <tbody>
                  {preview.rows.map((row, index) => (
                    <tr key={index}>
                      {row.map((cell, cellIndex) => <td key={cellIndex}>{cell}</td>)}
                    </tr>
                  ))}
                  {preview.rows.length === 0 && (
                    <tr><td colSpan={preview.columns.length} className="muted">Kayıt bulunamadı.</td></tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>

          {/* Yazdırma/PDF için temiz rapor şablonu */}
          <div className="report-sheet">
            <div className="report-sheet__brand">AVATAR — Food Intelligence</div>
            <h2 className="report-sheet__title">{preview.title}</h2>
            <div className="report-sheet__meta">Oluşturma tarihi: {today}</div>
            <table className="report-sheet__table">
              <thead>
                <tr>{preview.columns.map((column) => <th key={column}>{column}</th>)}</tr>
              </thead>
              <tbody>
                {preview.rows.map((row, index) => (
                  <tr key={index}>{row.map((cell, cellIndex) => <td key={cellIndex}>{cell}</td>)}</tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}
    </section>
  );
}
