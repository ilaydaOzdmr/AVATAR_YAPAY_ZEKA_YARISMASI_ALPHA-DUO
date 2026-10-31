import { useEffect, useState, type FormEvent } from "react";
import { rowKey } from "../rowIdentity";
import { get } from "../api";

interface Product {
  id: number;
  name: string;
}

interface ButcheryItemCost {
  product_id: number;
  weight: number;
  quantity: number;
  is_waste: boolean;
  allocated_cost: number;
}

interface ButcheryResult {
  batch: {
    id: number;
    code: string;
    animal_type: string;
    input_weight: number;
    unit_cost: number;
    total_cost: number;
    batch_date: string;
    status: string;
  };
  items: ButcheryItemCost[];
  weight_balance_ok: boolean;
  fire_weight: number;
  cost_mismatch: boolean;
  expected_cost: number;
}

export default function Butchery() {
  const [batchID, setBatchID] = useState("");
  const [result, setResult] = useState<ButcheryResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [productNames, setProductNames] = useState<Map<number, string>>(new Map());

  useEffect(() => {
    get<Product[]>("/products")
      .then((products) => {
        setProductNames(new Map(products.map((product) => [product.id, product.name])));
      })
      .catch(() => {
        // Product names are a display convenience; keep the raw ID on failure.
      });
  }, []);

  async function loadBatch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setResult(null);

    const id = Number.parseInt(batchID, 10);
    if (Number.isNaN(id)) {
      setError("Enter a numeric butchery batch ID.");
      return;
    }

    try {
      setResult(await get<ButcheryResult>(`/butchery-batches/${id}`));
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  function productLabel(productId: number): string {
    const name = productNames.get(productId);
    return name ? `${name} (#${productId})` : `#${productId}`;
  }

  return (
    <section>
      <h2>Butchery Batch</h2>
      <form className="inline-form" onSubmit={loadBatch}>
        <input
          value={batchID}
          onChange={(event) => setBatchID(event.target.value)}
          placeholder="Butchery batch ID (e.g. 1)"
        />
        <button type="submit">Load batch</button>
      </form>
      {error && <p className="error">{error}</p>}
      {result && (
        <>
          <p>
            Batch <strong>{result.batch.code}</strong> — {result.batch.animal_type},{" "}
            {result.batch.input_weight} kg, total cost {result.batch.total_cost.toFixed(2)}
          </p>
          {!result.weight_balance_ok && (
            <p className="warning">Fire: {result.fire_weight} kg</p>
          )}
          {result.cost_mismatch && (
            <p className="warning">
              Maliyet uyuşmazlığı (beklenen: {result.expected_cost.toFixed(2)})
            </p>
          )}
          <table>
            <thead>
              <tr>
                <th>Product</th>
                <th>Qty</th>
                <th>Weight</th>
                <th>Waste</th>
                <th>Allocated cost</th>
              </tr>
            </thead>
            <tbody>
              {result.items.map((item) => (
                <tr key={rowKey(item)}>
                  <td>{productLabel(item.product_id)}</td>
                  <td>{item.quantity}</td>
                  <td>{item.weight}</td>
                  <td>{item.is_waste ? "Yes" : "No"}</td>
                  <td>{item.allocated_cost.toFixed(2)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </>
      )}
    </section>
  );
}
