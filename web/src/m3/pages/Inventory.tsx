import { useCallback, useEffect, useMemo, useState, type FormEvent } from "react";
import { get, post } from "../../api";

interface Material {
  id: number;
  code: string;
  name: string;
  unit: string;
  default_purchase_price: number;
  shelf_life_days: number;
  created_at: string;
}

interface Product {
  id: number;
  code: string;
  name: string;
  unit: string;
  list_price: number;
  category: string;
  is_finished: boolean;
  shelf_life_days: number;
  created_at: string;
}

interface Lot {
  id: number;
  code: string;
  item_type: string;
  item_id: number;
  qty: number;
  unit: string;
  unit_cost: number;
  received_at: string | null;
  expiry_date: string | null;
  status: string;
}

const today = () => new Date().toISOString().slice(0, 10);

export default function Inventory() {
  const [materials, setMaterials] = useState<Material[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [lots, setLots] = useState<Lot[]>([]);
  const [query, setQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [showMaterialForm, setShowMaterialForm] = useState(false);
  const [showProductForm, setShowProductForm] = useState(false);
  const [showStockForm, setShowStockForm] = useState(false);
  const [showNewLotForm, setShowNewLotForm] = useState(false);
  const [feedback, setFeedback] = useState<string | null>(null);

  const [mCode, setMCode] = useState("");
  const [mName, setMName] = useState("");
  const [mUnit, setMUnit] = useState("kg");
  const [mPrice, setMPrice] = useState("");
  const [mShelfLife, setMShelfLife] = useState("30");

  const [pCode, setPCode] = useState("");
  const [pName, setPName] = useState("");
  const [pUnit, setPUnit] = useState("kg");
  const [pPrice, setPPrice] = useState("");
  const [pCategory, setPCategory] = useState("finished");
  const [pShelfLife, setPShelfLife] = useState("30");

  const [stockLotId, setStockLotId] = useState("");
  const [stockQty, setStockQty] = useState("");
  const [stockMovementType, setStockMovementType] = useState("inbound");

  const [lotItemType, setLotItemType] = useState("material");
  const [lotItemId, setLotItemId] = useState("");
  const [lotCode, setLotCode] = useState("");
  const [lotQty, setLotQty] = useState("");
  const [lotUnitCost, setLotUnitCost] = useState("");
  const [lotProductionDate, setLotProductionDate] = useState(today());

  const load = useCallback(() => {
    get<Material[]>("/materials").then(setMaterials).catch(() => {});
    get<Product[]>("/products").then(setProducts).catch(() => {});
    get<Lot[]>("/lots").then(setLots).catch(() => {});
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const materialById = useMemo(() => new Map(materials.map((m) => [m.id, m])), [materials]);
  const productById = useMemo(() => new Map(products.map((p) => [p.id, p])), [products]);

  function itemName(lot: Lot) {
    if (lot.item_type === "material") return materialById.get(lot.item_id)?.name ?? "—";
    return productById.get(lot.item_id)?.name ?? "—";
  }

  const stockValue = lots.reduce((sum, lot) => sum + lot.qty * lot.unit_cost, 0);
  const totalKg = lots.reduce((sum, lot) => sum + lot.qty, 0);

  const rows = lots.filter((lot) => {
    const name = itemName(lot);
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
        shelf_life_days: Number.parseInt(mShelfLife, 10) || 0,
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

  async function createProduct(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFeedback(null);
    try {
      await post("/products", {
        code: pCode,
        name: pName,
        unit: pUnit,
        list_price: Number.parseFloat(pPrice) || 0,
        category: pCategory,
        is_finished: true,
        shelf_life_days: Number.parseInt(pShelfLife, 10) || 0,
      });
      setPCode("");
      setPName("");
      setPPrice("");
      setShowProductForm(false);
      setFeedback("Ürün eklendi.");
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
      setFeedback("Stok hareketi kaydedildi.");
      load();
    } catch (err) {
      setFeedback(err instanceof Error ? err.message : String(err));
    }
  }

  async function createLotWithStock(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFeedback(null);
    const itemId = Number.parseInt(lotItemId, 10);
    const item = lotItemType === "material" ? materialById.get(itemId) : productById.get(itemId);
    const qty = Number.parseFloat(lotQty);
    if (!item || Number.isNaN(qty) || qty <= 0 || !lotProductionDate) {
      setFeedback("Ürün/hammadde, üretim tarihi ve geçerli miktar seçin.");
      return;
    }
    const defaultCost = lotItemType === "material"
      ? (item as Material).default_purchase_price
      : (item as Product).list_price;
    const unitCost = Number.parseFloat(lotUnitCost);
    const code = lotCode.trim() || `LOT-${item.code}-${Date.now()}`;
    try {
      const created = await post<Lot>("/lots", {
        code,
        item_type: lotItemType,
        item_id: item.id,
        qty,
        unit: item.unit,
        purchase_price: Number.isFinite(unitCost) ? unitCost : defaultCost,
        unit_cost: Number.isFinite(unitCost) ? unitCost : defaultCost,
        received_at: `${lotProductionDate} 00:00:00`,
        status: "active",
      });
      setLotCode("");
      setLotQty("");
      setLotUnitCost("");
      setShowNewLotForm(false);
      setFeedback(
        `Lot oluşturuldu. Üretim: ${lotProductionDate}, SKT: ${created.expiry_date ?? "tanımsız (raf ömrü 0)"}.`,
      );
      load();
    } catch (err) {
      setFeedback(err instanceof Error ? err.message : String(err));
    }
  }

  const itemOptions = lotItemType === "material"
    ? materials.map((m) => ({ id: m.id, code: m.code, name: m.name, unit: m.unit, shelfLife: m.shelf_life_days }))
    : products.map((p) => ({ id: p.id, code: p.code, name: p.name, unit: p.unit, shelfLife: p.shelf_life_days }));
  const selectedItem = itemOptions.find((item) => item.id === Number.parseInt(lotItemId, 10));

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">DEPO YÖNETİMİ</div>
          <h1>Stok &amp; Ürünler</h1>
          <p>Hammadde/ürün kartları, ürün bazlı raf ömrü, lotlar ve stok hareketleri tek yerde.</p>
        </div>
        <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
          <button className="btn btn-secondary" type="button" onClick={() => setShowStockForm((v) => !v)}>↓ Mevcut lota stok</button>
          <button className="btn btn-primary" type="button" onClick={() => setShowNewLotForm((v) => !v)}>＋ Yeni lot / stok</button>
          <button className="btn btn-secondary" type="button" onClick={() => setShowMaterialForm((v) => !v)}>＋ Hammadde</button>
          <button className="btn btn-secondary" type="button" onClick={() => setShowProductForm((v) => !v)}>＋ Ürün</button>
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
              <input value={mShelfLife} onChange={(e) => setMShelfLife(e.target.value)} placeholder="Raf ömrü (gün)" style={{ width: 120 }} />
              <button className="btn btn-primary" type="submit">Kaydet</button>
            </form>
          </div>
        </div>
      )}

      {showProductForm && (
        <div className="panel" style={{ marginBottom: 16 }}>
          <div className="panel-head"><h2>Yeni ürün</h2></div>
          <div style={{ padding: 20 }}>
            <form className="inline-form" onSubmit={createProduct}>
              <input value={pCode} onChange={(e) => setPCode(e.target.value)} placeholder="Kod (PRD-...)" required />
              <input value={pName} onChange={(e) => setPName(e.target.value)} placeholder="Ad" required />
              <input value={pUnit} onChange={(e) => setPUnit(e.target.value)} placeholder="Birim" style={{ width: 80 }} />
              <input value={pPrice} onChange={(e) => setPPrice(e.target.value)} placeholder="Satış fiyatı" style={{ width: 110 }} />
              <input value={pCategory} onChange={(e) => setPCategory(e.target.value)} placeholder="Kategori" style={{ width: 110 }} />
              <input value={pShelfLife} onChange={(e) => setPShelfLife(e.target.value)} placeholder="Raf ömrü (gün)" style={{ width: 120 }} />
              <button className="btn btn-primary" type="submit">Kaydet</button>
            </form>
          </div>
        </div>
      )}

      {showNewLotForm && (
        <div className="panel" style={{ marginBottom: 16 }}>
          <div className="panel-head">
            <h2>Yeni lot + stok girişi</h2>
            <p>Üretim tarihi + ürün raf ömrü = SKT. Stok hareketi aynı işlemde yazılır.</p>
          </div>
          <div style={{ padding: 20 }}>
            <form className="inline-form" onSubmit={createLotWithStock}>
              <select value={lotItemType} onChange={(e) => { setLotItemType(e.target.value); setLotItemId(""); }}>
                <option value="material">Hammadde</option>
                <option value="product">Ürün</option>
              </select>
              <select value={lotItemId} onChange={(e) => setLotItemId(e.target.value)} required>
                <option value="">Seçin...</option>
                {itemOptions.map((item) => (
                  <option key={item.id} value={item.id}>{item.code} — {item.name}</option>
                ))}
              </select>
              <input value={lotCode} onChange={(e) => setLotCode(e.target.value)} placeholder="Lot kodu (ops.)" style={{ width: 160 }} />
              <input value={lotQty} onChange={(e) => setLotQty(e.target.value)} placeholder="Miktar" style={{ width: 100 }} required />
              <input value={lotUnitCost} onChange={(e) => setLotUnitCost(e.target.value)} placeholder="Birim maliyet" style={{ width: 120 }} />
              <input type="date" value={lotProductionDate} onChange={(e) => setLotProductionDate(e.target.value)} required />
              <button className="btn btn-primary" type="submit">Lot oluştur</button>
            </form>
            {selectedItem && (
              <p className="muted" style={{ marginTop: 10 }}>
                Seçili: {selectedItem.code} · Birim: {selectedItem.unit} · Raf ömrü: {selectedItem.shelfLife} gün
                {selectedItem.shelfLife > 0 ? ` → SKT: ${lotProductionDate} + ${selectedItem.shelfLife} gün` : " (SKT hesaplanmaz)"}
              </p>
            )}
          </div>
        </div>
      )}

      {showStockForm && (
        <div className="panel" style={{ marginBottom: 16 }}>
          <div className="panel-head"><h2>Mevcut lota stok hareketi</h2></div>
          <div style={{ padding: 20 }}>
            <form className="inline-form" onSubmit={createStock}>
              <select value={stockLotId} onChange={(e) => setStockLotId(e.target.value)}>
                <option value="">Lot seçin...</option>
                {lots.map((lot) => (
                  <option key={lot.id} value={lot.id}>{lot.code} — {itemName(lot)} ({lot.item_type})</option>
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
              <button className="btn btn-primary" type="submit">Kaydet</button>
            </form>
          </div>
        </div>
      )}

      {feedback && <p className="muted" style={{ marginBottom: 12 }}>{feedback}</p>}

      <div className="metrics">
        <div className="metric">
          <div className="metric-top">Aktif lot <span className="metric-icon">▤</span></div>
          <strong>{lots.length}</strong>
          <div className="metric-foot">{materials.length} hammadde · {products.length} ürün</div>
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
          <div className="metric-foot">Maliyete göre</div>
        </div>
      </div>

      {products.length > 0 && (
        <div className="panel" style={{ marginBottom: 16 }}>
          <div className="panel-head">
            <h2>Ürün kartları</h2>
            <p>Her ürünün eklenme tarihi ve raf ömrü ayrıdır; SKT lot oluşturulurken hesaplanır.</p>
          </div>
          <div className="table-wrap">
            <table>
              <thead>
                <tr><th>Kod</th><th>Ürün</th><th>Kategori</th><th>Raf ömrü (gün)</th><th>Eklenme</th><th>Satış fiyatı</th></tr>
              </thead>
              <tbody>
                {products.map((product) => (
                  <tr key={product.id}>
                    <td className="id">{product.code}</td>
                    <td>{product.name}</td>
                    <td>{product.category}</td>
                    <td>{product.shelf_life_days}</td>
                    <td>{product.created_at ?? "—"}</td>
                    <td>€ {product.list_price.toFixed(2)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      <div className="panel">
        <div className="panel-head">
          <h2>Lotlar</h2>
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
              <tr><th>Lot</th><th>Tip</th><th>Ürün/Hammadde</th><th>Stok</th><th>Birim maliyet</th><th>Üretim/Alış</th><th>SKT</th><th>Durum</th></tr>
            </thead>
            <tbody>
              {rows.map((lot) => (
                <tr key={lot.id}>
                  <td className="id">{lot.code}</td>
                  <td>{lot.item_type === "material" ? "Hammadde" : "Ürün"}</td>
                  <td>{itemName(lot)}</td>
                  <td>{lot.qty} {lot.unit}</td>
                  <td>€ {lot.unit_cost.toFixed(2)}</td>
                  <td>{lot.received_at ?? "—"}</td>
                  <td>{lot.expiry_date ?? "—"}</td>
                  <td>{lot.status === "blocked" ? <span className="status block">Bloke</span> : <span className="status ok">Aktif</span>}</td>
                </tr>
              ))}
              {rows.length === 0 && <tr><td colSpan={8} className="muted">Sonuç bulunamadı.</td></tr>}
            </tbody>
          </table>
        </div>
        <div className="table-foot"><span>{rows.length} lot</span><span>Lot hareketleri kayıt altındadır</span></div>
      </div>
    </section>
  );
}
