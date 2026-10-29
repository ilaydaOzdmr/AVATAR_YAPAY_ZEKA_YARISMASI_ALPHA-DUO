import { useState, type FormEvent } from "react";
import { get } from "../../api";

interface CostSummary {
  input_cost: number;
  service_cost: number;
  total_cost: number;
  finished_qty: number;
  ekp: number;
  vkp: number;
  gross_profit: number;
  margin: number;
}

export default function Production() {
  const [orderId, setOrderId] = useState("1");
  const [summary, setSummary] = useState<CostSummary | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function load(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setSummary(null);
    const id = Number.parseInt(orderId, 10);
    if (Number.isNaN(id)) {
      setError("Geçerli bir üretim emri numarası girin.");
      return;
    }
    try {
      setSummary(await get<CostSummary>(`/production-orders/${id}`));
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">ÜRETİM MERKEZİ</div>
          <h1>Üretim Emirleri</h1>
          <p>Dış üretim, sevk edilen hammaddeler, mamul kabulü ve gerçek maliyet.</p>
        </div>
        <button className="btn btn-primary" type="button">＋ Yeni üretim emri</button>
      </div>

      <div className="metrics">
        <div className="metric">
          <div className="metric-top">Gerçekleşen miktar <span className="metric-icon">⚙</span></div>
          <strong>{summary ? summary.finished_qty : "—"} <small style={{ fontSize: 13 }}>paket</small></strong>
          <div className="metric-foot">Kabul edilen mamul</div>
        </div>
        <div className="metric">
          <div className="metric-top">Toplam maliyet <span className="metric-icon">€</span></div>
          <strong>€ {summary ? summary.total_cost.toFixed(2) : "—"}</strong>
          <div className="metric-foot">Hammadde + hizmet</div>
        </div>
        <div className="metric">
          <div className="metric-top">Birim gerçek maliyet (EKP) <span className="metric-icon">€</span></div>
          <strong style={{ color: "var(--green-3)" }}>€ {summary ? summary.ekp.toFixed(2) : "—"}</strong>
          <div className="metric-foot">Gerçek mamul miktarına göre</div>
        </div>
        <div className="metric">
          <div className="metric-top">Brüt kâr / marj <span className="metric-icon">↗</span></div>
          <strong>{summary ? `${summary.gross_profit.toFixed(2)} · %${summary.margin.toFixed(2)}` : "—"}</strong>
          <div className="metric-foot">VKP {summary ? summary.vkp.toFixed(2) : "—"} karşısında</div>
        </div>
      </div>

      <div className="two-col">
        <div className="panel">
          <div className="panel-head">
            <h2>Gerçek maliyet hesabı</h2>
            <form className="inline-form" onSubmit={load}>
              <input
                value={orderId}
                onChange={(event) => setOrderId(event.target.value)}
                placeholder="Üretim emri ID"
                style={{ width: 130 }}
              />
              <button className="btn btn-secondary" type="submit">Hesapla</button>
            </form>
          </div>
          {error && <p className="muted" style={{ padding: "0 20px" }}>{error}</p>}
          <div className="table-wrap">
            <table>
              <tbody>
                <tr><td>Hammadde maliyeti</td><td style={{ textAlign: "right" }}>€ {summary ? summary.input_cost.toFixed(2) : "—"}</td></tr>
                <tr><td>Hizmet maliyeti</td><td style={{ textAlign: "right" }}>€ {summary ? summary.service_cost.toFixed(2) : "—"}</td></tr>
                <tr><td><b>Toplam maliyet</b></td><td style={{ textAlign: "right" }}><b>€ {summary ? summary.total_cost.toFixed(2) : "—"}</b></td></tr>
              </tbody>
            </table>
          </div>
          <div style={{ padding: "16px 20px", background: "#f7faf7", borderRadius: "0 0 15px 15px" }}>
            <span className="muted">Birim gerçek maliyet (EKP)</span>
            <strong style={{ float: "right", fontSize: 20, color: "var(--green-3)" }}>
              € {summary ? summary.ekp.toFixed(2) : "—"}
            </strong>
          </div>
        </div>

        <div className="panel">
          <div className="panel-head">
            <div>
              <h2>Üretim akışı</h2>
              <p>İzlenebilir adımlar</p>
            </div>
          </div>
          <div className="trace-flow">
            <div className="flow-node">
              <div><strong>1. Hammadde sevki</strong><small>ET-2026-001 · 100 kg et</small></div>
              <span className="status ok">Tamam</span>
            </div>
            <div className="flow-line" />
            <div className="flow-node">
              <div><strong>2. Dış üretim</strong><small>URT-2026-045 · Hizmet € 467,50</small></div>
              <span className="status ok">Tamam</span>
            </div>
            <div className="flow-line" />
            <div className="flow-node">
              <div><strong>3. Mamul kabulü</strong><small>SUC-2026-045 · 85 paket</small></div>
              <span className="status ok">Tamam</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
