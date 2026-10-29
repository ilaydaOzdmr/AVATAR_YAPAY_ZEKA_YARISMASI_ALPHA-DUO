import { useCallback, useEffect, useState, type FormEvent } from "react";
import { get, post } from "../../api";

interface Market {
  id: number;
  code: string;
  name: string;
  contact: string;
}

interface Lot {
  id: number;
  code: string;
  item_type: string;
  item_id: number;
  qty: number;
  unit: string;
}

interface Product {
  id: number;
  name: string;
  list_price: number;
}

interface Sale {
  id: number;
  code: string;
  market_id: number;
  sold_at: string;
  status: string;
}

export default function Sales() {
  const [markets, setMarkets] = useState<Market[]>([]);
  const [lots, setLots] = useState<Lot[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [sales, setSales] = useState<Sale[]>([]);

  const [lotId, setLotId] = useState("");
  const [marketId, setMarketId] = useState("");
  const [qty, setQty] = useState("1");
  const [feedback, setFeedback] = useState<string | null>(null);

  const load = useCallback(() => {
    get<Market[]>("/markets").then(setMarkets).catch(() => {});
    get<Lot[]>("/lots?item_type=product").then(setLots).catch(() => {});
    get<Product[]>("/products").then(setProducts).catch(() => {});
    get<Sale[]>("/sales").then(setSales).catch(() => {});
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const productById = (id: number) => products.find((p) => p.id === id);
  const marketName = (id: number) => markets.find((m) => m.id === id)?.name ?? `Market #${id}`;

  async function quickSale(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFeedback(null);
    const lot = lots.find((l) => l.id === Number.parseInt(lotId, 10));
    const quantity = Number.parseFloat(qty);
    if (!lot || Number.isNaN(quantity) || quantity <= 0 || !marketId) {
      setFeedback("Lot, market ve geçerli bir miktar seçin.");
      return;
    }
    const price = productById(lot.item_id)?.list_price ?? 0;
    try {
      await post("/sales", {
        code: `SAT-${Date.now()}`,
        market_id: Number.parseInt(marketId, 10),
        sold_at: new Date().toISOString().slice(0, 19).replace("T", " "),
        status: "completed",
        client_id: "web",
        lines: [{ lot_id: lot.id, qty: quantity, unit_price: price, total_price: price * quantity }],
      });
      setQty("1");
      setFeedback("Satış kaydedildi.");
      load();
    } catch (err) {
      setFeedback(err instanceof Error ? err.message : String(err));
    }
  }

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">SATIŞ &amp; DAĞITIM</div>
          <h1>Satış &amp; Sevkiyat</h1>
          <p>Satış satırlarını lotla eşleştirin; sevk edilen her paketin izini koruyun.</p>
        </div>
      </div>

      <div className="metrics">
        <div className="metric">
          <div className="metric-top">Mamul lotu <span className="metric-icon">▣</span></div>
          <strong>{lots.length}</strong>
          <div className="metric-foot">Satışa hazır</div>
        </div>
        <div className="metric">
          <div className="metric-top">Toplam satış <span className="metric-icon">↗</span></div>
          <strong>{sales.length}</strong>
          <div className="metric-foot">Kayıtlı işlem</div>
        </div>
        <div className="metric">
          <div className="metric-top">Market <span className="metric-icon">▤</span></div>
          <strong>{markets.length}</strong>
          <div className="metric-foot">Alıcı</div>
        </div>
        <div className="metric">
          <div className="metric-top">Bloke kontrolü <span className="metric-icon" style={{ background: "var(--red-bg)", color: "var(--red)" }}>⛨</span></div>
          <strong>Aktif</strong>
          <div className="metric-foot positive">Satış öncesi doğrulama</div>
        </div>
      </div>

      <div className="two-col">
        <div className="panel">
          <div className="panel-head"><h2>Hızlı satış</h2></div>
          <div style={{ padding: 20 }}>
            <form className="stack-form" onSubmit={quickSale}>
              <label>Mamul lotu
                <select value={lotId} onChange={(e) => setLotId(e.target.value)}>
                  <option value="">Lot seçin...</option>
                  {lots.map((lot) => <option key={lot.id} value={lot.id}>{lot.code}</option>)}
                </select>
              </label>
              <label>Market
                <select value={marketId} onChange={(e) => setMarketId(e.target.value)}>
                  <option value="">Market seçin...</option>
                  {markets.map((m) => <option key={m.id} value={m.id}>{m.name} ({m.code})</option>)}
                </select>
              </label>
              <label>Miktar
                <input value={qty} onChange={(e) => setQty(e.target.value)} placeholder="1" />
              </label>
              <button className="btn btn-primary" type="submit">Satışı kaydet</button>
            </form>
            {feedback && <p className="muted" style={{ marginTop: 8 }}>{feedback}</p>}
          </div>
        </div>

        <div className="panel">
          <div className="panel-head"><h2>Son satışlar</h2></div>
          <div className="table-wrap">
            <table>
              <thead>
                <tr><th>Satış</th><th>Market</th><th>Tarih</th><th>Durum</th></tr>
              </thead>
              <tbody>
                {sales.map((sale) => (
                  <tr key={sale.id}>
                    <td className="id">{sale.code}</td>
                    <td>{marketName(sale.market_id)}</td>
                    <td>{sale.sold_at}</td>
                    <td><span className="status ok">{sale.status}</span></td>
                  </tr>
                ))}
                {sales.length === 0 && <tr><td colSpan={4} className="muted">Satış bulunamadı.</td></tr>}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </section>
  );
}
