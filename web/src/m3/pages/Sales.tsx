import { useEffect, useState, type FormEvent } from "react";
import { get, post } from "../../api";

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
  code: string;
  name: string;
  list_price: number;
}

export default function Sales() {
  const [lots, setLots] = useState<Lot[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [lotId, setLotId] = useState("");
  const [qty, setQty] = useState("1");
  const [message, setMessage] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    get<Lot[]>("/lots?item_type=product").then(setLots).catch(() => {});
    get<Product[]>("/products").then(setProducts).catch(() => {});
  }, []);

  const productName = (id: number) => products.find((p) => p.id === id)?.name ?? "—";
  const productListPrice = (id: number) => products.find((p) => p.id === id)?.list_price ?? 0;

  async function quickSale(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setMessage("");
    const id = Number.parseInt(lotId, 10);
    const quantity = Number.parseFloat(qty);
    if (Number.isNaN(id) || Number.isNaN(quantity) || quantity <= 0) {
      setError("Lot ve geçerli bir miktar seçin.");
      return;
    }
    const lot = lots.find((l) => l.id === id);
    if (!lot) {
      setError("Lot bulunamadı.");
      return;
    }
    const price = productListPrice(lot.item_id);
    try {
      await post("/sales", {
        code: `SAT-${Date.now()}`,
        market_id: 1,
        sold_at: new Date().toISOString().slice(0, 19).replace("T", " "),
        status: "completed",
        client_id: "web",
        lines: [{ lot_id: id, qty: quantity, unit_price: price, total_price: price * quantity }],
      });
      setMessage("Satış kaydedildi.");
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
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
        <button className="btn btn-primary" type="button">＋ Yeni satış</button>
      </div>

      <div className="metrics">
        <div className="metric">
          <div className="metric-top">Mamul lotu <span className="metric-icon">▣</span></div>
          <strong>{lots.length}</strong>
          <div className="metric-foot">Satışa hazır</div>
        </div>
        <div className="metric">
          <div className="metric-top">Toplam mamul stok <span className="metric-icon">▤</span></div>
          <strong>{lots.reduce((s, l) => s + l.qty, 0).toLocaleString("tr-TR")} paket</strong>
          <div className="metric-foot">Mevcut miktar</div>
        </div>
        <div className="metric">
          <div className="metric-top">İzlenebilir satış satırı <span className="metric-icon">⌘</span></div>
          <strong>100%</strong>
          <div className="metric-foot">Lot kaydı zorunlu</div>
        </div>
        <div className="metric">
          <div className="metric-top">Bloke lot kontrolü <span className="metric-icon" style={{ background: "var(--red-bg)", color: "var(--red)" }}>⛨</span></div>
          <strong>Aktif</strong>
          <div className="metric-foot positive">Satış öncesi doğrulama</div>
        </div>
      </div>

      <div className="two-col">
        <div className="panel">
          <div className="panel-head">
            <h2>Mamul lotları</h2>
          </div>
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Lot</th>
                  <th>Ürün</th>
                  <th>Mevcut stok</th>
                </tr>
              </thead>
              <tbody>
                {lots.map((lot) => (
                  <tr key={lot.id}>
                    <td className="id">{lot.code}</td>
                    <td>{productName(lot.item_id)}</td>
                    <td>{lot.qty} {lot.unit}</td>
                  </tr>
                ))}
                {lots.length === 0 && (
                  <tr><td colSpan={3} className="muted">Mamul lotu bulunamadı.</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>

        <div className="panel">
          <div className="panel-head">
            <h2>Hızlı satış</h2>
          </div>
          <div style={{ padding: 20 }}>
            <form className="stack-form" onSubmit={quickSale}>
              <label>
                Mamul lotu
                <select value={lotId} onChange={(event) => setLotId(event.target.value)}>
                  <option value="">Lot seçin...</option>
                  {lots.map((lot) => (
                    <option key={lot.id} value={lot.id}>
                      {lot.code}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Miktar
                <input value={qty} onChange={(event) => setQty(event.target.value)} placeholder="1" />
              </label>
              <button className="btn btn-primary" type="submit">Satışı kaydet</button>
            </form>
            {error && <p className="muted" style={{ marginTop: 8 }}>{error}</p>}
            {message && <p className="muted" style={{ marginTop: 8, color: "var(--green-4)" }}>{message}</p>}
          </div>
        </div>
      </div>
    </section>
  );
}
