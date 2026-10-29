import { useEffect, useState, type FormEvent } from "react";
import { get, post } from "../../api";

interface Anomaly {
  type: string;
  severity: string;
  message: string;
  ref: string;
}

interface RecallEvent {
  code: string;
}

interface RecallMarket {
  market_code: string;
  market_name: string;
  contact: string;
  qty: number;
}

interface RecallResult {
  event: RecallEvent;
  source_lot_code: string;
  affected_lots: string[];
  markets: RecallMarket[];
  message: string;
}

export default function Recall() {
  const [sourceLotCode, setSourceLotCode] = useState("");
  const [reason, setReason] = useState("Kalite sorunu");
  const [result, setResult] = useState<RecallResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [confirmed, setConfirmed] = useState<Set<string>>(new Set());
  const [anomalies, setAnomalies] = useState<Anomaly[]>([]);

  useEffect(() => {
    get<Anomaly[]>("/anomalies").then(setAnomalies).catch(() => {});
  }, []);

  async function createRecall(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setResult(null);
    setConfirmed(new Set());
    const code = sourceLotCode.trim();
    if (!code) {
      setError("Kaynak lot numarası girin.");
      return;
    }
    try {
      setResult(await post<RecallResult>("/recalls", { source_lot_code: code, reason }));
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  function downloadCsv() {
    if (!result) return;
    const header = "Alıcı,Kod,Miktar,İletişim\n";
    const rows = result.markets
      .map((m) => `${m.market_name},${m.market_code},${m.qty},${m.contact}`)
      .join("\n");
    const blob = new Blob([`\ufeff${header}${rows}`], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "geri-cagirma-listesi.csv";
    a.click();
    URL.revokeObjectURL(url);
  }

  const totalQty = result ? result.markets.reduce((sum, m) => sum + m.qty, 0) : 0;
  const criticalAnomalies = anomalies.filter((a) => a.severity === "critical");

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">KALİTE &amp; GÜVENLİK</div>
          <h1>Geri Çağırma Merkezi</h1>
          <p>Etkilenen lotları bloke edin, alıcıları belirleyin ve aksiyonları kayıt altına alın.</p>
        </div>
      </div>

      {anomalies.length > 0 && (
        <div className="recall-banner">
          <span style={{ fontSize: 20 }}>⚠</span>
          <div>
            <strong>{anomalies.length} kalite/stok uyarısı</strong>
            <p>
              {criticalAnomalies.length > 0
                ? criticalAnomalies.map((a) => `${a.ref}: ${a.message}`).join(" · ")
                : anomalies.map((a) => `${a.ref}: ${a.message}`).join(" · ")}
            </p>
          </div>
        </div>
      )}

      <div className="panel lot-search" style={{ marginBottom: 24 }}>
        <label htmlFor="recallLot">Geri çağırma başlat</label>
        <form className="lot-input-row" onSubmit={createRecall}>
          <input
            id="recallLot"
            value={sourceLotCode}
            onChange={(event) => setSourceLotCode(event.target.value)}
            placeholder="Kaynak lot kodu (örn. ET-2026-001)"
          />
          <input
            value={reason}
            onChange={(event) => setReason(event.target.value)}
            placeholder="Sebep"
            style={{ flex: 1 }}
          />
          <button className="btn btn-danger" type="submit">Geri çağırma başlat</button>
        </form>
        {error && <p className="muted" style={{ marginTop: 8 }}>{error}</p>}
      </div>

      {result && (
        <>
          <div className="metrics">
            <div className="metric">
              <div className="metric-top">Etkilenen mamul <span className="metric-icon">▤</span></div>
              <strong>{totalQty} paket</strong>
              <div className="metric-foot">{result.affected_lots.join(", ") || result.source_lot_code}</div>
            </div>
            <div className="metric">
              <div className="metric-top">Dağıtım noktası <span className="metric-icon">↗</span></div>
              <strong>{result.markets.length}</strong>
              <div className="metric-foot">markete sevk edildi</div>
            </div>
            <div className="metric">
              <div className="metric-top">Alıcı teyidi <span className="metric-icon">✓</span></div>
              <strong>{confirmed.size} / {result.markets.length}</strong>
              <div className="metric-foot">Teyit bekleniyor</div>
            </div>
            <div className="metric">
              <div className="metric-top">Satış blokajı <span className="metric-icon" style={{ background: "var(--red-bg)", color: "var(--red)" }}>⛨</span></div>
              <strong>Aktif</strong>
              <div className="metric-foot positive">Lot bazlı satış engeli</div>
            </div>
          </div>

          <div className="panel">
            <div className="panel-head">
              <h2>Alıcı iletişim ve miktar listesi</h2>
              <button className="btn btn-secondary" type="button" onClick={downloadCsv}>⇩ Listeyi dışa aktar</button>
            </div>
            <div className="table-wrap">
              <table>
                <thead>
                  <tr><th>Alıcı</th><th>Kod</th><th>Miktar</th><th>İletişim</th><th>Durum</th><th>Aksiyon</th></tr>
                </thead>
                <tbody>
                  {result.markets.map((market) => {
                    const done = confirmed.has(market.market_code);
                    return (
                      <tr key={market.market_code}>
                        <td><b>{market.market_name}</b></td>
                        <td className="id">{market.market_code}</td>
                        <td>{market.qty} paket</td>
                        <td>{market.contact}</td>
                        <td>{done ? <span className="status ok">Teyit edildi</span> : <span className="status warn">Teyit bekliyor</span>}</td>
                        <td>
                          <button className="btn btn-secondary" type="button" disabled={done} onClick={() => setConfirmed((prev) => new Set(prev).add(market.market_code))}>
                            {done ? "✓ Teyit edildi" : "Teyit alındı"}
                          </button>
                        </td>
                      </tr>
                    );
                  })}
                  {result.markets.length === 0 && <tr><td colSpan={6} className="muted">Etkilenen market bulunamadı.</td></tr>}
                </tbody>
              </table>
            </div>
          </div>

          <p className="muted" style={{ marginTop: 12 }}>{result.message}</p>
        </>
      )}
    </section>
  );
}
