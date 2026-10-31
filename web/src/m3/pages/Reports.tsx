import { useEffect, useState } from "react";
import { get } from "../../api";
import { downloadCsv } from "../csv";

interface PurchaseRow {
  lot_code: string;
  material_code: string;
  material_name: string;
  supplier_code: string;
  qty: number;
  unit_cost: number;
  total_cost: number;
  received_at: string | null;
}

interface ProductionRow {
  code: string;
  status: string;
  input_kg: number;
  finished_qty: number;
  total_cost: number;
  ekp: number;
}

interface ButcheryRow {
  code: string;
  animal_type: string;
  input_weight: number;
  parts_weight: number;
  fire_weight: number;
  total_cost: number;
  part_count: number;
}

interface MovementRow {
  lot_code: string;
  movement_type: string;
  qty: number;
  unit_cost: number;
  total_cost: number;
  occurred_at: string;
}

interface SalesRow {
  code: string;
  market_code: string;
  sold_at: string;
  line_count: number;
  qty: number;
  total_price: number;
}

interface RecallRow {
  code: string;
  reason: string;
  created_at: string;
  blocked_lots: string[];
  blocked_lot_count: number;
}

interface ReportSummary {
  generated_at: string;
  purchases: PurchaseRow[];
  production: ProductionRow[];
  butchery: ButcheryRow[];
  lot_movements: MovementRow[];
  sales: SalesRow[];
  recalls: RecallRow[];
}

interface ReportData {
  key: string;
  title: string;
  columns: string[];
  rows: (string | number)[][];
}

export default function Reports() {
  const [summary, setSummary] = useState<ReportSummary | null>(null);
  const [active, setActive] = useState<string>("purchases");
  const [feedback, setFeedback] = useState<string | null>(null);

  useEffect(() => {
    get<ReportSummary>("/reports/summary")
      .then(setSummary)
      .catch((err) => setFeedback(err instanceof Error ? err.message : String(err)));
  }, []);

  const today = new Date().toLocaleDateString("tr-TR", { day: "numeric", month: "long", year: "numeric" });

  const reports: ReportData[] = summary
    ? [
        {
          key: "purchases",
          title: "Alış Raporu",
          columns: ["Lot", "Malzeme", "Tedarikçi", "Miktar", "Birim maliyet", "Toplam"],
          rows: summary.purchases.map((row) => [
            row.lot_code,
            `${row.material_name} (${row.material_code})`,
            row.supplier_code,
            `${row.qty} kg`,
            `€ ${row.unit_cost.toFixed(2)}`,
            `€ ${row.total_cost.toFixed(2)}`,
          ]),
        },
        {
          key: "production",
          title: "Üretim Maliyeti Raporu",
          columns: ["Emir", "Durum", "Girdi (kg)", "Mamul", "Toplam maliyet", "EKP"],
          rows: summary.production.map((row) => [
            row.code,
            row.status,
            row.input_kg,
            row.finished_qty,
            `€ ${row.total_cost.toFixed(2)}`,
            `€ ${row.ekp.toFixed(2)}`,
          ]),
        },
        {
          key: "butchery",
          title: "Parçalama Raporu",
          columns: ["Parti", "Hayvan", "Giriş (kg)", "Parça (kg)", "Fire (kg)", "Maliyet", "Parça"],
          rows: summary.butchery.map((row) => [
            row.code,
            row.animal_type,
            row.input_weight,
            row.parts_weight,
            row.fire_weight,
            `€ ${row.total_cost.toFixed(2)}`,
            row.part_count,
          ]),
        },
        {
          key: "movements",
          title: "Lot Hareket Geçmişi",
          columns: ["Lot", "Hareket türü", "Miktar", "Birim maliyet", "Toplam", "Tarih"],
          rows: summary.lot_movements.map((row) => [
            row.lot_code,
            row.movement_type,
            row.qty,
            `€ ${row.unit_cost.toFixed(2)}`,
            `€ ${row.total_cost.toFixed(2)}`,
            row.occurred_at,
          ]),
        },
        {
          key: "sales",
          title: "Satış Raporu",
          columns: ["Satış", "Market", "Tarih", "Satır", "Miktar", "Toplam"],
          rows: summary.sales.map((row) => [
            row.code,
            row.market_code,
            row.sold_at,
            row.line_count,
            row.qty,
            `€ ${row.total_price.toFixed(2)}`,
          ]),
        },
        {
          key: "recalls",
          title: "Geri Çağırma Raporu",
          columns: ["Kod", "Sebep", "Tarih", "Bloke lot sayısı", "Bloke lotlar"],
          rows: summary.recalls.map((row) => [
            row.code,
            row.reason,
            row.created_at,
            row.blocked_lot_count,
            row.blocked_lots.join(", "),
          ]),
        },
      ]
    : [];

  const activeReport = reports.find((report) => report.key === active) ?? reports[0];

  function downloadPdf() {
    window.print();
  }

  function downloadCsvCurrent() {
    if (!activeReport) return;
    downloadCsv(
      `${activeReport.key}-rapor.csv`,
      activeReport.columns,
      activeReport.rows.map((row) => row.map((cell) => String(cell).replace(/^€ /, ""))),
    );
  }

  const counts = summary
    ? {
        purchases: summary.purchases.length,
        production: summary.production.length,
        butchery: summary.butchery.length,
        movements: summary.lot_movements.length,
        sales: summary.sales.length,
        recalls: summary.recalls.length,
      }
    : null;

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">İŞLETME ANALİTİĞİ</div>
          <h1>Raporlar</h1>
          <p>Alış, üretim, parçalama, lot hareketleri, satış ve geri çağırma raporları.</p>
        </div>
      </div>

      {feedback && <p className="muted" style={{ marginBottom: 12 }}>{feedback}</p>}

      {counts && (
        <div className="metrics">
          <div className="metric">
            <div className="metric-top">Alış <span className="metric-icon">↓</span></div>
            <strong>{counts.purchases}</strong>
            <div className="metric-foot">Hammadde lotu</div>
          </div>
          <div className="metric">
            <div className="metric-top">Üretim <span className="metric-icon">⚙</span></div>
            <strong>{counts.production}</strong>
            <div className="metric-foot">Üretim emri</div>
          </div>
          <div className="metric">
            <div className="metric-top">Parçalama <span className="metric-icon">♧</span></div>
            <strong>{counts.butchery}</strong>
            <div className="metric-foot">Parçalama kaydı</div>
          </div>
          <div className="metric">
            <div className="metric-top">Satış <span className="metric-icon">↗</span></div>
            <strong>{counts.sales}</strong>
            <div className="metric-foot">Satış kaydı</div>
          </div>
        </div>
      )}

      {reports.length > 0 && activeReport && (
        <div className="panel">
          <div className="panel-head">
            <div className="tabs">
              {reports.map((report) => (
                <button
                  key={report.key}
                  type="button"
                  className={report.key === active ? "tab active" : "tab"}
                  onClick={() => setActive(report.key)}
                >
                  {report.title}
                </button>
              ))}
            </div>
            <div style={{ display: "flex", gap: 8 }}>
              <button className="btn btn-primary" type="button" onClick={downloadPdf}>PDF indir</button>
              <button className="btn btn-secondary" type="button" onClick={downloadCsvCurrent}>CSV indir</button>
            </div>
          </div>
          <div className="table-wrap">
            <table>
              <thead>
                <tr>{activeReport.columns.map((column) => <th key={column}>{column}</th>)}</tr>
              </thead>
              <tbody>
                {activeReport.rows.map((row, index) => (
                  <tr key={index}>{row.map((cell, cellIndex) => <td key={cellIndex}>{cell}</td>)}</tr>
                ))}
                {activeReport.rows.length === 0 && (
                  <tr><td colSpan={activeReport.columns.length} className="muted">Kayıt bulunamadı.</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Yazdırma/PDF için temiz rapor şablonu */}
      {activeReport && (
        <div className="report-sheet">
          <div className="report-sheet__brand">AVATAR — Food Intelligence</div>
          <h2 className="report-sheet__title">{activeReport.title}</h2>
          <div className="report-sheet__meta">Oluşturma tarihi: {today}</div>
          <table className="report-sheet__table">
            <thead>
              <tr>{activeReport.columns.map((column) => <th key={column}>{column}</th>)}</tr>
            </thead>
            <tbody>
              {activeReport.rows.map((row, index) => (
                <tr key={index}>{row.map((cell, cellIndex) => <td key={cellIndex}>{cell}</td>)}</tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
