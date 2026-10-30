import { useCallback, useEffect, useMemo, useState, type FormEvent } from "react";
import { get, post } from "../../api";

interface Material {
  id: number;
  code: string;
  name: string;
  unit: string;
  default_purchase_price: number;
}

interface Lot {
  id: number;
  code: string;
  item_type: string;
  item_id: number;
  qty: number;
  unit: string;
  unit_cost: number;
  expiry_date: string | null;
  status: string;
}

export default function Inventory() {
  const [materials, setMaterials] = useState<Material[]>([]);
  const [lots, setLots] = useState<Lot[]>([]);
  const [query, setQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [showMaterialForm, setShowMaterialForm] = useState(false);
  const [showStockForm, setShowStockForm] = useState(false);
  const [feedback, setFeedback] = useState<string | null>(null);

  const [mCode, setMCode] = useState("");
  const [mName, setMName] = useState("");
  const [mUnit, setMUnit] = useState("kg");
  const [mPrice, setMPrice] = useState("");

  const [stockLotId, setStockLotId] = useState("");
  const [stockQty, setStockQty] = useState("");
  const [stockMovementType, setStockMovementType] = useState("inbound");

  const load = useCallback(() => {
    get<Material[]>("/materials").then(setMaterials).catch(() => {});
    get<Lot[]>("/lots?item_type=material").then(setLots).catch(() => {});
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const nameById = useMemo(() => new Map(materials.map((m) => [m.id, m.name])), [materials]);
  const stockValue = lots.reduce((sum, lot) => sum + lot.qty * lot.unit_cost, 0);
  const totalKg = lots.reduce((sum, lot) => sum + lot.qty, 0);

  const rows = lots.filter((lot) => {
    const name = nameById.get(lot.item_id) ?? "";
    const matchesQuery = `${lot.code} ${name}`.toLowerCase().includes(query.toLowerCase());
    const blocked = lot.status === "blocked";
    const matchesStatus = statusFilter === "all" || (statusFilter === "blocked" ? blocked : !blocked);
    return matchesQuery && matchesStatus;
  });

  async function createMaterial(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFeedback(null);
    try {
      await post("/materials", {
        code: mCode,
        name: mName,
        unit: mUnit,
        is_raw: true,
        default_purchase_price: Number.parseFloat(mPrice) || 0,
      });
      setMCode("");
      setMName("");
      setMPrice("");
      setShowMaterialForm(false);
      setFeedback("Hammadde eklendi.");
      load();
    } catch (err) {
      setFeedback(err instanceof Error ? err.message : String(err));
    }
  }

  async function createStock(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFeedback(null);
    const lot = lots.find((l) => l.id === Number.parseInt(stockLotId, 10));
    const qty = Number.parseFloat(stockQty);
    if (!lot || Number.isNaN(qty) || qty <= 0) {
      setFeedback("Lot ve geçerli bir miktar seçin.");
      return;
    }
    try {
      await post("/stock-movements", {
        lot_id: lot.id,
        movement_type: stockMovementType,
        qty,
        unit_cost: lot.unit_cost,
        total_cost: qty * lot.unit_cost,
      });
      setStockQty("");
      setShowStockForm(false);
      setFeedback("Stok girişi kaydedildi.");
      load();
    } catch (err) {
      setFeedback(err instanceof Error ? err.message : String(err));
    }
  }

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">DEPO YÖNETİMİ</div>
          <h1>Stok &amp; Hammaddeler</h1>
          <p>Hammadde kartları, alış fiyatları, lotlar ve stok hareketleri tek yerde.</p>
        </div>
        <div style={{ display: "flex", gap: 8 }}>
          <button className="btn btn-secondary" type="button" onClick={() => setShowStockForm((v) => !v)}>↓ Stok girişi</button>
          <button className="btn btn-primary" type="button" onClick={() => setShowMaterialForm((v) => !v)}>＋ Hammadde ekle</button>
        </div>
      </div>

      {showMaterialForm && (
        <div className="panel" style={{ marginBottom: 16 }}>
          <div className="panel-head"><h2>Yeni hammadde</h2></div>
          <div style={{ padding: 20 }}>
            <form className="inline-form" onSubmit={createMaterial}>
              <input value={mCode} onChange={(e) => setMCode(e.target.value)} placeholder="Kod (MAT-...)" required />
              <input value={mName} onChange={(e) => setMName(e.target.value)} placeholder="Ad" required />
              <input value={mUnit} onChange={(e) => setMUnit(e.target.value)} placeholder="Birim" style={{ width: 80 }} />
              <input value={mPrice} onChange={(e) => setMPrice(e.target.value)} placeholder="Alış fiyatı" style={{ width: 110 }} />
              <button className="btn btn-primary" type="submit">Kaydet</button>
            </form>
          </div>
        </div>
      )}

      {showStockForm && (
        <div className="panel" style={{ marginBottom: 16 }}>
          <div className="panel-head"><h2>Stok girişi</h2></div>
          <div style={{ padding: 20 }}>
            <form className="inline-form" onSubmit={createStock}>
              <select value={stockLotId} onChange={(e) => setStockLotId(e.target.value)}>
                <option value="">Lot seçin...</option>
                {lots.map((lot) => (
                  <option key={lot.id} value={lot.id}>{lot.code} — {nameById.get(lot.item_id) ?? ""}</option>
                ))}
              </select>
              <select value={stockMovementType} onChange={(e) => setStockMovementType(e.target.value)}>
                <option value="inbound">Giriş (inbound)</option>
                <option value="transfer">Depo transferi</option>
                <option value="return">İade</option>
                <option value="rework">Yeniden işleme</option>
                <option value="shipment">Sevkiyat</option>
              </select>
              <input value={stockQty} onChange={(e) => setStockQty(e.target.value)} placeholder="Miktar (kg)" style={{ width: 120 }} />
              <button className="btn btn-primary" type="submit">Giriş yap</button>
            </form>
          </div>
        </div>
      )}

      {feedback && <p className="muted" style={{ marginBottom: 12 }}>{feedback}</p>}

      <div className="metrics">
        <div className="metric">
          <div className="metric-top">Aktif hammadde lotu <span className="metric-icon">▤</span></div>
          <strong>{lots.length}</strong>
          <div className="metric-foot">{materials.length} ürün grubu</div>
        </div>
        <div className="metric">
          <div className="metric-top">Bloke lot <span className="metric-icon" style={{ background: "var(--red-bg)", color: "var(--red)" }}>⚠</span></div>
          <strong>{lots.filter((l) => l.status === "blocked").length}</strong>
          <div className="metric-foot warning">Satış engelli</div>
        </div>
        <div className="metric">
          <div className="metric-top">Toplam stok <span className="metric-icon">↓</span></div>
          <strong>{totalKg.toLocaleString("tr-TR")} kg</strong>
          <div className="metric-foot">Mevcut ağırlık</div>
        </div>
        <div className="metric">
          <div className="metric-top">Stok değeri <span className="metric-icon">€</span></div>
          <strong>€ {stockValue.toLocaleString("tr-TR", { minimumFractionDigits: 2 })}</strong>
          <div className="metric-foot">Alış maliyetine göre</div>
        </div>
      </div>

      <div className="panel">
        <div className="panel-head">
          <h2>Hammadde lotları</h2>
          <div className="toolbar">
            <input className="search-wide" placeholder="Ürün veya lot ara..." value={query} onChange={(e) => setQuery(e.target.value)} />
            <select className="filter" value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
              <option value="all">Tüm durumlar</option>
              <option value="active">Aktif</option>
              <option value="blocked">Bloke</option>
            </select>
          </div>
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr><th>Lot</th><th>Hammadde</th><th>Stok</th><th>Birim maliyet</th><th>SKT</th><th>Durum</th></tr>
            </thead>
            <tbody>
              {rows.map((lot) => (
                <tr key={lot.id}>
                  <td className="id">{lot.code}</td>
                  <td>{nameById.get(lot.item_id) ?? "—"}</td>
                  <td>{lot.qty} {lot.unit}</td>
                  <td>€ {lot.unit_cost.toFixed(2)}</td>
                  <td>{lot.expiry_date ?? "—"}</td>
                  <td>{lot.status === "blocked" ? <span className="status block">Bloke</span> : <span className="status ok">Aktif</span>}</td>
                </tr>
              ))}
              {rows.length === 0 && <tr><td colSpan={6} className="muted">Sonuç bulunamadı.</td></tr>}
            </tbody>
          </table>
        </div>
        <div className="table-foot"><span>{rows.length} lot</span><span>Lot hareketleri kayıt altındadır</span></div>
      </div>
    </section>
  );
}
