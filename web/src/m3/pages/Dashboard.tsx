import { useCallback, useEffect, useState } from "react";
import { get } from "../../api";

interface Lot {
  id: number;
  item_type: string;
  qty: number;
  unit_cost: number;
}

interface Anomaly {
  type: string;
  severity: string;
  message: string;
  ref: string;
}

interface ProductionOrder {
  id: number;
  code: string;
  status: string;
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

export default function Dashboard() {
  const [stockValue, setStockValue] = useState(0);
  const [orderCount, setOrderCount] = useState(0);
  const [avgEkp, setAvgEkp] = useState(0);
  const [anomalies, setAnomalies] = useState<Anomaly[]>([]);
  const [orders, setOrders] = useState<ProductionOrder[]>([]);
  const [sales, setSales] = useState<Sale[]>([]);
  const [markets, setMarkets] = useState<Market[]>([]);

  const load = useCallback(() => {
    get<Lot[]>("/lots").then((lots) => {
      const value = lots
        .filter((l) => l.item_type === "material")
        .reduce((sum, l) => sum + l.qty * l.unit_cost, 0);
      setStockValue(value);
    }).catch(() => {});

    get<Anomaly[]>("/anomalies").then(setAnomalies).catch(() => {});
    get<Market[]>("/markets").then(setMarkets).catch(() => {});

    get<ProductionOrder[]>("/production-orders").then((list) => {
      setOrders(list);
      setOrderCount(list.length);
      Promise.all(
        list.map((o) =>
          get<{ ekp: number }>(`/production-orders/${o.id}`)
            .then((s) => s.ekp)
            .catch(() => 0),
        ),
      ).then((ekps) => {
        const positive = ekps.filter((e) => e > 0);
        setAvgEkp(positive.length ? positive.reduce((a, b) => a + b, 0) / positive.length : 0);
      });
    }).catch(() => {});

    get<Sale[]>("/sales").then(setSales).catch(() => {});
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const critical = anomalies.filter((a) => a.severity === "critical").length;
  const marketName = (id: number) => markets.find((m) => m.id === id)?.name ?? `Market #${id}`;

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">İŞLETME ÖZETİ</div>
          <h1>Günaydın, yönetici 👋</h1>
          <p>İşletmenizin bugünkü operasyon özeti ve dikkat gerektiren noktalar.</p>
        </div>
        <button className="btn btn-primary" type="button" onClick={load}>↻ Yenile</button>
      </div>

      <div className="metrics">
        <div className="metric">
          <div className="metric-top">Toplam stok değeri <span className="metric-icon">▤</span></div>
          <strong>€ {stockValue.toLocaleString("tr-TR", { minimumFractionDigits: 2 })}</strong>
          <div className="metric-foot">Hammadde lotlarına göre</div>
        </div>
        <div className="metric">
          <div className="metric-top">Üretim emri <span className="metric-icon">⚙</span></div>
          <strong>{orderCount}</strong>
          <div className="metric-foot">{orders.filter((o) => o.status === "completed").length} tamamlandı</div>
        </div>
        <div className="metric">
          <div className="metric-top">Ortalama EKP <span className="metric-icon">€</span></div>
          <strong style={{ color: "var(--green-3)" }}>€ {avgEkp.toFixed(2)}</strong>
          <div className="metric-foot">Üretim emirlerine göre</div>
        </div>
        <div className="metric">
          <div className="metric-top">Kritik uyarı <span className="metric-icon" style={{ background: "var(--amber-bg)", color: "var(--amber)" }}>⚠</span></div>
          <strong>{critical}</strong>
          <div className="metric-foot">{anomalies.length} toplam anomali</div>
        </div>
      </div>

      <div className="dashboard-grid">
        <div className="panel">
          <div className="panel-head">
            <div>
              <h2>Dikkat gerektirenler</h2>
              <p>Anomali kurallarından üretildi</p>
            </div>
            <span className="status warn">{anomalies.length} bildirim</span>
          </div>
          <div className="alert-list">
            {anomalies.map((anomaly, index) => (
              <div className="alert" key={`${anomaly.ref}-${index}`}>
                <div className={`alert-symbol ${anomaly.severity === "critical" ? "red" : "amber"}`}>
                  {anomaly.severity === "critical" ? "⚠" : "↗"}
                </div>
                <div>
                  <b>{anomaly.message} · {anomaly.ref}</b>
                  <p>Anomali türü: {anomaly.type} · Öncelik: {anomaly.severity}</p>
                </div>
              </div>
            ))}
            {anomalies.length === 0 && <p className="muted">Açık anomali bulunmuyor.</p>}
          </div>
        </div>

        <div className="panel">
          <div className="panel-head">
            <div>
              <h2>Üretim durumu</h2>
              <p>Emirlerin dağılımı</p>
            </div>
          </div>
          <div className="table-wrap">
            <table>
              <thead>
                <tr><th>Durum</th><th>Adet</th></tr>
              </thead>
              <tbody>
                <tr><td><span className="status ok">Tamamlandı</span></td><td>{orders.filter((o) => o.status === "completed").length}</td></tr>
                <tr><td><span className="status progress">İşlemde</span></td><td>{orders.filter((o) => o.status !== "completed").length}</td></tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <div className="panel" style={{ marginTop: 16 }}>
        <div className="panel-head">
          <div>
            <h2>Son işlemler</h2>
            <p>Satış ve üretim hareketleri</p>
          </div>
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr><th>Referans</th><th>Tür</th><th>Market / Durum</th></tr>
            </thead>
            <tbody>
              {sales.slice(0, 3).map((sale) => (
                <tr key={`sale-${sale.id}`}>
                  <td className="id">{sale.code}</td>
                  <td>Satış</td>
                  <td>{marketName(sale.market_id)} · <span className="status ok">{sale.status}</span></td>
                </tr>
              ))}
              {orders.slice(0, 3).map((order) => (
                <tr key={`order-${order.id}`}>
                  <td className="id">{order.code}</td>
                  <td>Üretim</td>
                  <td><span className={order.status === "completed" ? "status ok" : "status progress"}>{order.status}</span></td>
                </tr>
              ))}
              {sales.length === 0 && orders.length === 0 && (
                <tr><td colSpan={3} className="muted">İşlem bulunamadı.</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </section>
  );
}
