import { useState, type FormEvent } from "react";
import { get } from "../../api";

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

type Mode = "forward" | "backward";

export default function Traceability() {
  const [mode, setMode] = useState<Mode>("forward");
  const [lotCode, setLotCode] = useState("ET-2026-001");
  const [affected, setAffected] = useState<AffectedLot[]>([]);
  const [sources, setSources] = useState<RawMaterialSource[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [searched, setSearched] = useState(false);

  async function trace(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setAffected([]);
    setSources([]);
    setSearched(false);
    const code = lotCode.trim();
    if (!code) {
      setError("Lütfen bir lot numarası girin.");
      return;
    }
    try {
      if (mode === "forward") {
        setAffected(await get<AffectedLot[]>(`/trace/forward?lot=${encodeURIComponent(code)}`));
      } else {
        setSources(await get<RawMaterialSource[]>(`/trace/backward?lot=${encodeURIComponent(code)}`));
      }
      setSearched(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  const totalQty = affected.reduce((sum, lot) => sum + lot.qty, 0);

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">UÇTAN UCA İZLENEBİLİRLİK</div>
          <h1>Lot İzlenebilirliği</h1>
          <p>Bir hammaddeden etkilenen bütün mamullere veya mamulden tedarikçiye ulaşın.</p>
        </div>
        <button className="btn btn-secondary" type="button">⇩ Rapor indir</button>
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
