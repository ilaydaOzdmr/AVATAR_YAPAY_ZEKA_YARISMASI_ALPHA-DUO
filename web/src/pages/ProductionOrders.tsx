import { useState, type FormEvent } from "react";
import { get } from "../api";

interface CostSummary {
  input_cost: number;
  service_cost: number;
  total_cost: number;
  finished_qty: number;
  ekp: number;
  vkp: number;
  gross_profit: number;
  margin: number;
}

export default function ProductionOrders() {
  const [orderID, setOrderID] = useState("");
  const [summary, setSummary] = useState<CostSummary | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function loadCost(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setSummary(null);

    const id = Number.parseInt(orderID, 10);
    if (Number.isNaN(id)) {
      setError("Enter a numeric production order ID.");
      return;
    }

    try {
      setSummary(await get<CostSummary>(`/production-orders/${id}`));
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  return (
    <section>
      <h2>Production Order Cost</h2>
      <form className="inline-form" onSubmit={loadCost}>
        <input
          value={orderID}
          onChange={(event) => setOrderID(event.target.value)}
          placeholder="Production order ID (e.g. 1)"
        />
        <button type="submit">Load cost</button>
      </form>
      {error && <p className="error">{error}</p>}
      {summary && (
        <table>
          <tbody>
            <tr>
              <th>Input cost</th>
              <td>{summary.input_cost.toFixed(2)}</td>
            </tr>
            <tr>
              <th>Service cost</th>
              <td>{summary.service_cost.toFixed(2)}</td>
            </tr>
            <tr>
              <th>Total cost</th>
              <td>{summary.total_cost.toFixed(2)}</td>
            </tr>
            <tr>
              <th>Finished quantity</th>
              <td>{summary.finished_qty}</td>
            </tr>
            <tr>
              <th>EKP</th>
              <td>{summary.ekp.toFixed(2)}</td>
            </tr>
            <tr>
              <th>VKP</th>
              <td>{summary.vkp.toFixed(2)}</td>
            </tr>
            <tr>
              <th>Gross profit</th>
              <td>{summary.gross_profit.toFixed(2)}</td>
            </tr>
            <tr>
              <th>Margin (%)</th>
              <td>{summary.margin.toFixed(2)}</td>
            </tr>
          </tbody>
        </table>
      )}
    </section>
  );
}
