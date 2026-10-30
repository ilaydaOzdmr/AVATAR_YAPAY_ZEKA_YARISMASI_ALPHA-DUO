import { useCallback, useEffect, useMemo, useState, type FormEvent } from "react";
import { get, post } from "../../api";

interface ProductionOrder {
  id: number;
  code: string;
  status: string;
  service_name: string | null;
  service_unit_cost: number;
  service_qty: number;
  expected_qty: number;
}

interface CostSummary {
  input_cost: number;
  service_cost: number;
  total_cost: number;
  expected_qty: number;
  finished_qty: number;
  quantity_diff: number;
  ekp: number;
  vkp: number;
  gross_profit: number;
  margin: number;
}

interface Lot {
  id: number;
  code: string;
}

interface OrderLine {
  lot_id: string;
  qty: string;
}

export default function Production() {
  const [orders, setOrders] = useState<ProductionOrder[]>([]);
  const [selectedId, setSelectedId] = useState("");
  const [summary, setSummary] = useState<CostSummary | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [feedback, setFeedback] = useState<string | null>(null);

  const [code, setCode] = useState("");
  const [serviceName, setServiceName] = useState("");
  const [serviceUnitCost, setServiceUnitCost] = useState("");
  const [serviceQty, setServiceQty] = useState("");
  const [materialLots, setMaterialLots] = useState<Lot[]>([]);
  const [productLots, setProductLots] = useState<Lot[]>([]);
  const [inputs, setInputs] = useState<OrderLine[]>([]);
  const [outputs, setOutputs] = useState<OrderLine[]>([]);

  const load = useCallback(() => {
    get<ProductionOrder[]>("/production-orders").then(setOrders).catch(() => {});
    get<Lot[]>("/lots?item_type=material").then(setMaterialLots).catch(() => {});
    get<Lot[]>("/lots?item_type=product").then(setProductLots).catch(() => {});
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (!selectedId) {
      setSummary(null);
      return;
    }
    get<CostSummary>(`/production-orders/${selectedId}`).then(setSummary).catch(() => setSummary(null));
  }, [selectedId]);

  const openOrders = orders.filter((o) => o.status !== "completed").length;
  const completedOrders = orders.filter((o) => o.status === "completed").length;
  const avgEkp = useMemo(() => {
    // Ekp is derived on demand; here we show the selected order's EKP.
    return summary?.ekp ?? 0;
  }, [summary]);

  async function createOrder(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFeedback(null);
    const parsedInputs = inputs
      .map((line) => ({ material_lot_id: Number.parseInt(line.lot_id, 10), qty: Number.parseFloat(line.qty) || 0 }))
      .filter((line) => !Number.isNaN(line.material_lot_id));
    const parsedOutputs = outputs
      .map((line) => ({ product_lot_id: Number.parseInt(line.lot_id, 10), qty: Number.parseFloat(line.qty) || 0 }))
      .filter((line) => !Number.isNaN(line.product_lot_id));
    try {
      await post("/production-orders", {
        code,
        status: "draft",
        service_name: serviceName,
        service_unit_cost: Number.parseFloat(serviceUnitCost) || 0,
        service_qty: Number.parseFloat(serviceQty) || 0,
        inputs: parsedInputs,
        outputs: parsedOutputs,
      });
      setCode("");
      setServiceName("");
      setServiceUnitCost("");
      setServiceQty("");
      setInputs([]);
      setOutputs([]);
      setShowForm(false);
      setFeedback("Üretim emri oluşturuldu.");
      load();
    } catch (err) {
      setFeedback(err instanceof Error ? err.message : String(err));
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
        <button className="btn btn-primary" type="button" onClick={() => setShowForm((v) => !v)}>＋ Yeni üretim emri</button>
      </div>

      {showForm && (
        <div className="panel" style={{ marginBottom: 16 }}>
          <div className="panel-head"><h2>Yeni üretim emri</h2></div>
          <div style={{ padding: 20 }}>
            <form className="stack-form" onSubmit={createOrder}>
              <div className="inline-form">
                <input value={code} onChange={(e) => setCode(e.target.value)} placeholder="Kod (URT-...)" required />
                <input value={serviceName} onChange={(e) => setServiceName(e.target.value)} placeholder="Hizmet adı" />
                <input value={serviceUnitCost} onChange={(e) => setServiceUnitCost(e.target.value)} placeholder="Hizmet birim maliyeti" style={{ width: 150 }} />
                <input value={serviceQty} onChange={(e) => setServiceQty(e.target.value)} placeholder="Hizmet miktarı" style={{ width: 120 }} />
              </div>
              <div>
                <b style={{ fontSize: 13 }}>Hammadde girişleri (inputs)</b>
                {inputs.map((line, index) => (
                  <div key={index} className="inline-form" style={{ margin: "8px 0" }}>
                    <select value={line.lot_id} onChange={(e) => setInputs((prev) => prev.map((l, i) => (i === index ? { ...l, lot_id: e.target.value } : l)))}>
                      <option value="">Hammadde lotu...</option>
                      {materialLots.map((lot) => <option key={lot.id} value={lot.id}>{lot.code}</option>)}
                    </select>
                    <input value={line.qty} onChange={(e) => setInputs((prev) => prev.map((l, i) => (i === index ? { ...l, qty: e.target.value } : l)))} placeholder="Miktar" style={{ width: 100 }} />
                    <button className="btn btn-secondary" type="button" onClick={() => setInputs((prev) => prev.filter((_, i) => i !== index))}>✕</button>
                  </div>
                ))}
                <button className="btn btn-secondary" type="button" onClick={() => setInputs((prev) => [...prev, { lot_id: "", qty: "" }])}>＋ Hammadde ekle</button>
              </div>
              <div>
                <b style={{ fontSize: 13 }}>Mamul çıktıları (outputs)</b>
                {outputs.map((line, index) => (
                  <div key={index} className="inline-form" style={{ margin: "8px 0" }}>
                    <select value={line.lot_id} onChange={(e) => setOutputs((prev) => prev.map((l, i) => (i === index ? { ...l, lot_id: e.target.value } : l)))}>
                      <option value="">Mamul lotu...</option>
                      {productLots.map((lot) => <option key={lot.id} value={lot.id}>{lot.code}</option>)}
                    </select>
                    <input value={line.qty} onChange={(e) => setOutputs((prev) => prev.map((l, i) => (i === index ? { ...l, qty: e.target.value } : l)))} placeholder="Miktar" style={{ width: 100 }} />
                    <button className="btn btn-secondary" type="button" onClick={() => setOutputs((prev) => prev.filter((_, i) => i !== index))}>✕</button>
                  </div>
                ))}
                <button className="btn btn-secondary" type="button" onClick={() => setOutputs((prev) => [...prev, { lot_id: "", qty: "" }])}>＋ Mamul ekle</button>
              </div>
              <button className="btn btn-primary" type="submit">Oluştur</button>
            </form>
          </div>
        </div>
      )}

      {feedback && <p className="muted" style={{ marginBottom: 12 }}>{feedback}</p>}

      <div className="metrics">
        <div className="metric">
          <div className="metric-top">Açık emir <span className="metric-icon">⚙</span></div>
          <strong>{openOrders}</strong>
          <div className="metric-foot">İşlemde</div>
        </div>
        <div className="metric">
          <div className="metric-top">Tamamlanan <span className="metric-icon">✓</span></div>
          <strong>{completedOrders}</strong>
          <div className="metric-foot">Toplam {orders.length} kayıt</div>
        </div>
        <div className="metric">
          <div className="metric-top">Seçili EKP <span className="metric-icon">€</span></div>
          <strong style={{ color: "var(--green-3)" }}>€ {avgEkp.toFixed(2)}</strong>
          <div className="metric-foot">Gerçek mamul miktarına göre</div>
        </div>
        <div className="metric">
          <div className="metric-top">Brüt kâr / marj <span className="metric-icon">↗</span></div>
          <strong>{summary ? `${summary.gross_profit.toFixed(2)} · %${summary.margin.toFixed(2)}` : "—"}</strong>
          <div className="metric-foot">VKP {summary ? summary.vkp.toFixed(2) : "—"}</div>
        </div>
      </div>

      <div className="two-col">
        <div className="panel">
          <div className="panel-head">
            <h2>Üretim kayıtları</h2>
          </div>
          <div className="table-wrap">
            <table>
              <thead>
                <tr><th>Üretim emri</th><th>Beklenen</th><th>Hizmet miktarı</th><th>Durum</th></tr>
              </thead>
              <tbody>
                {orders.map((order) => (
                  <tr key={order.id} style={{ cursor: "pointer" }} onClick={() => setSelectedId(String(order.id))}>
                    <td className="id">{order.code}</td>
                    <td>{order.expected_qty}</td>
                    <td>{order.service_qty}</td>
                    <td>
                      {order.status === "completed" ? <span className="status ok">Tamamlandı</span> : <span className="status progress">İşlemde</span>}
                    </td>
                  </tr>
                ))}
                {orders.length === 0 && <tr><td colSpan={4} className="muted">Üretim emri bulunamadı.</td></tr>}
              </tbody>
            </table>
          </div>
        </div>

        <div className="panel">
          <div className="panel-head">
            <h2>Gerçek maliyet hesabı</h2>
            <form className="inline-form" onSubmit={(e) => e.preventDefault()}>
              <select value={selectedId} onChange={(e) => setSelectedId(e.target.value)}>
                <option value="">Emir seçin...</option>
                {orders.map((o) => <option key={o.id} value={o.id}>{o.code}</option>)}
              </select>
            </form>
          </div>
          <div className="table-wrap">
            <table>
              <tbody>
                <tr><td>Hammadde maliyeti</td><td style={{ textAlign: "right" }}>€ {summary ? summary.input_cost.toFixed(2) : "—"}</td></tr>
                <tr><td>Hizmet maliyeti</td><td style={{ textAlign: "right" }}>€ {summary ? summary.service_cost.toFixed(2) : "—"}</td></tr>
                <tr><td><b>Toplam maliyet</b></td><td style={{ textAlign: "right" }}><b>€ {summary ? summary.total_cost.toFixed(2) : "—"}</b></td></tr>
                <tr><td>Beklenen miktar</td><td style={{ textAlign: "right" }}>{summary ? summary.expected_qty : "—"}</td></tr>
                <tr><td>Gerçekleşen (mamul kabulü)</td><td style={{ textAlign: "right" }}>{summary ? summary.finished_qty : "—"}</td></tr>
                <tr>
                  <td>Miktar farkı</td>
                  <td style={{ textAlign: "right", color: summary && summary.quantity_diff > 0 ? "var(--amber)" : "var(--green-3)" }}>
                    {summary ? summary.quantity_diff : "—"}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div style={{ padding: "16px 20px", background: "#f7faf7", borderRadius: "0 0 15px 15px" }}>
            <span className="muted">Birim gerçek maliyet (EKP)</span>
            <strong style={{ float: "right", fontSize: 20, color: "var(--green-3)" }}>€ {summary ? summary.ekp.toFixed(2) : "—"}</strong>
          </div>
        </div>
      </div>
    </section>
  );
}
