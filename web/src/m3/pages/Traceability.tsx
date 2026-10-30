import { useState, type FormEvent } from "react";
import { get } from "../../api";
import { downloadCsv } from "../csv";

interface MarketImpact {
  market_code: string;
  market_name: string;
  contact: string;
  qty: number;
}

interface AffectedLot {
  lot_code: string;
  item_type: string;
  qty: number;
  unit: string;
  markets: MarketImpact[];
}

interface RawMaterialSource {
  lot_code: string;
  supplier_code: string;
  supplier_name: string;
  qty: number;
}

interface Redistribution {
  id: number;
  market_id: number;
  lot_id: number;
  sub_buyer: string;
  contact: string | null;
  qty: number;
}

interface Market {
  id: number;
  name: string;
}

interface RedistributionView {
  lot_code: string;
  market_name: string;
  sub_buyer: string;
  contact: string | null;
  qty: number;
}

type Mode = "forward" | "backward";

export default function Traceability() {
  const [mode, setMode] = useState<Mode>("forward");
  const [lotCode, setLotCode] = useState("ET-2026-001");
  const [affected, setAffected] = useState<AffectedLot[]>([]);
  const [sources, setSources] = useState<RawMaterialSource[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [searched, setSearched] = useState(false);
  const [redistributions, setRedistributions] = useState<RedistributionView[]>([]);

  async function loadRedistributions(affectedLots: AffectedLot[]) {
    setRedistributions([]);
    try {
      const [lots, markets] = await Promise.all([
        get<{ id: number; code: string }[]>("/lots"),
        get<Market[]>("/markets"),
      ]);
      const lotIdByCode = new Map(lots.map((lot) => [lot.code, lot.id]));
      const marketName = new Map(markets.map((m) => [m.id, m.name]));
      const views: RedistributionView[] = [];
      for (const lot of affectedLots) {
        const lotId = lotIdByCode.get(lot.lot_code);
        if (!lotId) continue;
        const reds = await get<Redistribution[]>(`/redistributions?lot_id=${lotId}`);
        for (const red of reds) {
          views.push({
            lot_code: lot.lot_code,
            market_name: marketName.get(red.market_id) ?? `Market #${red.market_id}`,
            sub_buyer: red.sub_buyer,
            contact: red.contact,
            qty: red.qty,
          });
        }
      }
      setRedistributions(views);
    } catch {
      setRedistributions([]);
    }
  }

  async function trace(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setAffected([]);
    setSources([]);
    setSearched(false);
    setRedistributions([]);
    const code = lotCode.trim();
    if (!code) {
      setError("Lütfen bir lot numarası girin.");
      return;
    }
    try {
      if (mode === "forward") {
        const lots = await get<AffectedLot[]>(`/trace/forward?lot=${encodeURIComponent(code)}`);
        setAffected(lots);
        void loadRedistributions(lots);
      } else {
        setSources(await get<RawMaterialSource[]>(`/trace/backward?lot=${encodeURIComponent(code)}`));
      }
      setSearched(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  const totalQty = affected.reduce((sum, lot) => sum + lot.qty, 0);

  function exportReport() {
    if (mode === "forward" && affected.length > 0) {
      const rows: (string | number)[][] = affected.flatMap((lot) =>
        lot.markets.length > 0
          ? lot.markets.map((m) => [lot.lot_code, lot.qty, lot.unit, m.market_code, m.market_name, m.qty, m.contact])
          : [[lot.lot_code, lot.qty, lot.unit, "—", "—", "—", "—"]],
      );
      downloadCsv(
        "izlenebilirlik-ileri-raporu.csv",
        ["Mamul lotu", "Miktar", "Birim", "Market kodu", "Market", "Sevk miktarı", "İletişim"],
        rows,
      );
    } else if (mode === "backward" && sources.length > 0) {
      const rows: (string | number)[][] = sources.map((s) => [s.lot_code, s.supplier_code, s.supplier_name, s.qty]);
      downloadCsv("izlenebilirlik-geri-raporu.csv", ["Hammadde lotu", "Tedarikçi kodu", "Tedarikçi", "Miktar"], rows);
    }
  }

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">UÇTAN UCA İZLENEBİLİRLİK</div>
          <h1>Lot İzlenebilirliği</h1>
          <p>Bir hammaddeden etkilenen bütün mamullere veya mamulden tedarikçiye ulaşın.</p>
        </div>
        <button className="btn btn-secondary" type="button" onClick={exportReport}>⇩ Rapor indir</button>
      </div>

      <div className="panel lot-search">
        <label htmlFor="lotInput">Hammadde veya mamul lot numarası</label>
        <form className="lot-input-row" onSubmit={trace}>
          <input
            id="lotInput"
            value={lotCode}
            onChange={(event) => setLotCode(event.target.value)}
            placeholder="Örn. ET-2026-001"
          />
          <button className="btn btn-primary" type="submit">İz sür →</button>
        </form>
        <div style={{ display: "flex", gap: 8, marginTop: 12 }}>
          <button
            className={mode === "forward" ? "btn btn-primary" : "btn btn-secondary"}
            type="button"
            onClick={() => setMode("forward")}
          >
            İleriye
          </button>
          <button
            className={mode === "backward" ? "btn btn-primary" : "btn btn-secondary"}
            type="button"
            onClick={() => setMode("backward")}
          >
            Geriye
          </button>
        </div>
      </div>

      {error && <p className="muted" style={{ marginBottom: 12 }}>{error}</p>}

      {searched && mode === "forward" && (
        <div className="panel">
          <div className="panel-head">
            <h2>İleriye izleme</h2>
            <span className="status block">{affected.length} mamul etkilendi</span>
          </div>
          <div className="trace-flow">
            <div className="flow-node">
              <div><strong>{lotCode}</strong><small>Kaynak lot</small></div>
              <span className="status block">Sorgulanıyor</span>
            </div>
            {affected.map((lot) => (
              <div key={lot.lot_code}>
                <div className="flow-line" />
                <div className="flow-node">
                  <div>
                    <strong>{lot.lot_code}</strong>
                    <small>{lot.qty} {lot.unit} · {lot.item_type}</small>
                  </div>
                  <span className="status block">Etkilendi</span>
                </div>
                {lot.markets.length > 0 && (
                  <>
                    <div className="flow-line" />
                    <div className="flow-children">
                      {lot.markets.map((market) => (
                        <div className="flow-node" key={market.market_code}>
                          <div>
                            <strong>{market.market_code}</strong>
                            <small>{market.qty} paket · {market.contact}</small>
                          </div>
                        </div>
                      ))}
                    </div>
                  </>
                )}
              </div>
            ))}
            {affected.length === 0 && <p className="muted">Etkilenen mamul bulunamadı.</p>}
          </div>
        </div>
      )}

      {searched && mode === "forward" && redistributions.length > 0 && (
        <div className="panel" style={{ marginTop: 16 }}>
          <div className="panel-head">
            <h2>İkinci kademe dağıtım</h2>
            <span className="status ok">Kaydedilmiş</span>
          </div>
          <div className="table-wrap">
            <table>
              <thead>
                <tr><th>Mamul lotu</th><th>Market</th><th>Alt alıcı</th><th>Miktar</th><th>İletişim</th></tr>
              </thead>
              <tbody>
                {redistributions.map((red, index) => (
                  <tr key={`${red.lot_code}-${index}`}>
                    <td className="id">{red.lot_code}</td>
                    <td>{red.market_name}</td>
                    <td><b>{red.sub_buyer}</b></td>
                    <td>{red.qty} paket</td>
                    <td>{red.contact ?? "—"}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {searched && mode === "backward" && (
        <div className="panel">
          <div className="panel-head">
            <h2>Geriye izleme</h2>
            <span className="status warn">{sources.length} hammadde kaynağı</span>
          </div>
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Hammadde lotu</th>
                  <th>Tedarikçi</th>
                  <th>Miktar</th>
                </tr>
              </thead>
              <tbody>
                {sources.map((source) => (
                  <tr key={source.lot_code}>
                    <td className="id">{source.lot_code}</td>
                    <td>{source.supplier_name} <span className="muted">({source.supplier_code})</span></td>
                    <td>{source.qty}</td>
                  </tr>
                ))}
                {sources.length === 0 && (
                  <tr><td colSpan={3} className="muted">Hammadde kaynağı bulunamadı.</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {searched && mode === "forward" && affected.length > 0 && (
        <div className="metrics" style={{ marginTop: 16, marginBottom: 0 }}>
          <div className="metric">
            <div className="metric-top">Etkilenen mamul lotu</div>
            <strong>{affected.length}</strong>
          </div>
          <div className="metric">
            <div className="metric-top">Toplam miktar</div>
            <strong>{totalQty} paket</strong>
          </div>
          <div className="metric">
            <div className="metric-top">Dağıtım noktası</div>
            <strong>{affected.reduce((sum, lot) => sum + lot.markets.length, 0)}</strong>
          </div>
        </div>
      )}
    </section>
  );
}
