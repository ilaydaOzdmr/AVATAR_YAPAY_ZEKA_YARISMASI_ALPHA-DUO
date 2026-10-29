import { useEffect, useMemo, useState } from "react";
import { get } from "../../api";

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

  useEffect(() => {
    get<Material[]>("/materials").then(setMaterials).catch(() => {});
    get<Lot[]>("/lots").then(setLots).catch(() => {});
  }, []);

  const nameById = useMemo(() => new Map(materials.map((m) => [m.id, m.name])), [materials]);
  const materialLots = lots.filter((lot) => lot.item_type === "material");
  const stockValue = materialLots.reduce((sum, lot) => sum + lot.qty * lot.unit_cost, 0);

  const rows = materialLots.filter((lot) => {
    const name = nameById.get(lot.item_id) ?? "";
    const matchesQuery = `${lot.code} ${name}`.toLowerCase().includes(query.toLowerCase());
    const blocked = lot.status === "blocked";
    const matchesStatus = statusFilter === "all" || (statusFilter === "blocked" ? blocked : !blocked);
    return matchesQuery && matchesStatus;
  });

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">DEPO YÖNETİMİ</div>
          <h1>Stok &amp; Hammaddeler</h1>
          <p>Hammadde kartları, alış fiyatları, lotlar ve stok hareketleri tek yerde.</p>
        </div>
        <button className="btn btn-primary" type="button">＋ Hammadde ekle</button>
      </div>

      <div className="metrics">
        <div className="metric">
          <div className="metric-top">Aktif hammadde lotu <span className="metric-icon">▤</span></div>
          <strong>{materialLots.length}</strong>
          <div className="metric-foot">{nameById.size} ürün grubu</div>
        </div>
        <div className="metric">
          <div className="metric-top">Bloke lot <span className="metric-icon" style={{ background: "var(--red-bg)", color: "var(--red)" }}>⚠</span></div>
          <strong>{materialLots.filter((l) => l.status === "blocked").length}</strong>
          <div className="metric-foot warning">Satış engelli</div>
        </div>
        <div className="metric">
          <div className="metric-top">Toplam stok miktarı <span className="metric-icon">↓</span></div>
          <strong>{materialLots.reduce((s, l) => s + l.qty, 0).toLocaleString("tr-TR")} kg</strong>
          <div className="metric-foot">Mevcut ağırlık</div>
        </div>
        <div className="metric">
          <div className="metric-top">Stok değerlemesi <span className="metric-icon">€</span></div>
          <strong>€ {stockValue.toLocaleString("tr-TR", { minimumFractionDigits: 2 })}</strong>
          <div className="metric-foot">Alış maliyetine göre</div>
        </div>
      </div>

      <div className="panel">
        <div className="panel-head">
          <h2>Hammadde lotları</h2>
          <div className="toolbar">
            <input
              className="search-wide"
              placeholder="Ürün veya lot ara..."
              value={query}
              onChange={(event) => setQuery(event.target.value)}
            />
            <select className="filter" value={statusFilter} onChange={(event) => setStatusFilter(event.target.value)}>
              <option value="all">Tüm durumlar</option>
              <option value="active">Aktif</option>
              <option value="blocked">Bloke</option>
            </select>
          </div>
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Lot numarası</th>
                <th>Hammadde</th>
                <th>Mevcut stok</th>
                <th>Birim maliyet</th>
                <th>SKT</th>
                <th>Durum</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((lot) => (
                <tr key={lot.id}>
                  <td className="id">{lot.code}</td>
                  <td>{nameById.get(lot.item_id) ?? "—"}</td>
                  <td>{lot.qty} {lot.unit}</td>
                  <td>€ {lot.unit_cost.toFixed(2)}</td>
                  <td>{lot.expiry_date ?? "—"}</td>
                  <td>
                    {lot.status === "blocked" ? (
                      <span className="status block">Bloke</span>
                    ) : (
                      <span className="status ok">Aktif</span>
                    )}
                  </td>
                </tr>
              ))}
              {rows.length === 0 && (
                <tr>
                  <td colSpan={6} className="muted">Sonuç bulunamadı.</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        <div className="table-foot">
          <span>{rows.length} lot gösteriliyor</span>
          <span>Lot hareketleri kayıt altındadır</span>
        </div>
      </div>
    </section>
  );
}
