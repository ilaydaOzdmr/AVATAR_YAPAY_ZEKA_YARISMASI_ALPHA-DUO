import { useState, type FormEvent } from "react";
import { get } from "../api";

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
  const [lotCode, setLotCode] = useState("");
  const [affectedLots, setAffectedLots] = useState<AffectedLot[]>([]);
  const [sources, setSources] = useState<RawMaterialSource[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [searched, setSearched] = useState(false);

  function selectMode(next: Mode) {
    setMode(next);
    setAffectedLots([]);
    setSources([]);
    setError(null);
    setSearched(false);
  }

  async function trace(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setAffectedLots([]);
    setSources([]);
    setSearched(false);

    const code = lotCode.trim();
    if (!code) {
      setError("Enter a lot code.");
      return;
    }

    try {
      if (mode === "forward") {
        setAffectedLots(await get<AffectedLot[]>(`/trace/forward?lot=${encodeURIComponent(code)}`));
      } else {
        setSources(await get<RawMaterialSource[]>(`/trace/backward?lot=${encodeURIComponent(code)}`));
      }
      setSearched(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  return (
    <section>
      <h2>Traceability</h2>
      <div className="tabs">
        <button
          type="button"
          className={mode === "forward" ? "tab active" : "tab"}
          onClick={() => selectMode("forward")}
        >
          Forward
        </button>
        <button
          type="button"
          className={mode === "backward" ? "tab active" : "tab"}
          onClick={() => selectMode("backward")}
        >
          Backward
        </button>
      </div>
      <form className="inline-form" onSubmit={trace}>
        <input
          value={lotCode}
          onChange={(event) => setLotCode(event.target.value)}
          placeholder="Lot code (e.g. ET-2026-001)"
        />
        <button type="submit">Trace</button>
      </form>
      {error && <p className="error">{error}</p>}
      {searched && mode === "forward" && affectedLots.length === 0 && (
        <p>No affected lots.</p>
      )}
      {searched && mode === "forward" && affectedLots.length > 0 && (
        <table>
          <thead>
            <tr>
              <th>Lot</th>
              <th>Type</th>
              <th>Quantity</th>
              <th>Unit</th>
              <th>Markets</th>
            </tr>
          </thead>
          <tbody>
            {affectedLots.map((lot) => (
              <tr key={lot.lot_code}>
                <td>{lot.lot_code}</td>
                <td>{lot.item_type}</td>
                <td>{lot.qty}</td>
                <td>{lot.unit}</td>
                <td>
                  {lot.markets.length === 0
                    ? "-"
                    : lot.markets.map((market) => (
                        <div key={`${lot.lot_code}-${market.market_code}`}>
                          {market.market_code} ({market.market_name}): {market.qty} — {market.contact}
                        </div>
                      ))}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {searched && mode === "backward" && sources.length === 0 && (
        <p>No raw material sources.</p>
      )}
      {searched && mode === "backward" && sources.length > 0 && (
        <table>
          <thead>
            <tr>
              <th>Lot</th>
              <th>Supplier code</th>
              <th>Supplier name</th>
              <th>Quantity</th>
            </tr>
          </thead>
          <tbody>
            {sources.map((source) => (
              <tr key={source.lot_code}>
                <td>{source.lot_code}</td>
                <td>{source.supplier_code}</td>
                <td>{source.supplier_name}</td>
                <td>{source.qty}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}
