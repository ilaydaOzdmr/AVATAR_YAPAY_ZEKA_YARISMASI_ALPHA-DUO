import { useState, type FormEvent } from "react";
import { post } from "../api";

interface RecallEvent {
  id: number;
  code: string;
  reason: string;
  source_lot_id: number;
  created_at: string;
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
  const [reason, setReason] = useState("");
  const [result, setResult] = useState<RecallResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function createRecall(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setResult(null);

    const code = sourceLotCode.trim();
    if (!code) {
      setError("Enter a source lot code.");
      return;
    }

    try {
      setResult(await post<RecallResult>("/recalls", { source_lot_code: code, reason }));
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  return (
    <section>
      <h2>Recall</h2>
      <form className="stack-form" onSubmit={createRecall}>
        <label>
          Source lot code
          <input
            value={sourceLotCode}
            onChange={(event) => setSourceLotCode(event.target.value)}
            placeholder="e.g. ET-2026-001"
          />
        </label>
        <label>
          Reason
          <input
            value={reason}
            onChange={(event) => setReason(event.target.value)}
            placeholder="Quality issue"
          />
        </label>
        <button type="submit">Create recall</button>
      </form>
      {error && <p className="error">{error}</p>}
      {result && (
        <>
          <p>
            Recall <strong>{result.event.code}</strong> for{" "}
            <strong>{result.source_lot_code}</strong>
          </p>
          <p>
            Affected lots:{" "}
            {result.affected_lots.length === 0 ? "-" : result.affected_lots.join(", ")}
          </p>
          <table>
            <thead>
              <tr>
                <th>Market code</th>
                <th>Market</th>
                <th>Quantity</th>
                <th>Contact</th>
              </tr>
            </thead>
            <tbody>
              {result.markets.map((market) => (
                <tr key={market.market_code}>
                  <td>{market.market_code}</td>
                  <td>{market.market_name}</td>
                  <td>{market.qty}</td>
                  <td>{market.contact}</td>
                </tr>
              ))}
            </tbody>
          </table>
          <p className="message">{result.message}</p>
        </>
      )}
    </section>
  );
}
