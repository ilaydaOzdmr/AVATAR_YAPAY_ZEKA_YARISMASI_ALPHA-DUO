import { useEffect, useState, type FormEvent } from "react";
import { get } from "../../api";

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
    code: string;
    animal_type: string;
    input_weight: number;
    unit_cost: number;
    total_cost: number;
  };
  items: ButcheryItemCost[];
  weight_balance_ok: boolean;
  fire_weight: number;
  cost_mismatch: boolean;
  expected_cost: number;
}

export default function Butchery() {
  const [batchId, setBatchId] = useState("1");
  const [result, setResult] = useState<ButcheryResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [productNames, setProductNames] = useState<Map<number, string>>(new Map());

  useEffect(() => {
    get<Product[]>("/products")
      .then((products) => setProductNames(new Map(products.map((p) => [p.id, p.name]))))
      .catch(() => {});
  }, []);

  async function load(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setResult(null);
    const id = Number.parseInt(batchId, 10);
    if (Number.isNaN(id)) {
      setError("Geçerli bir parçalama kaydı numarası girin.");
      return;
    }
    try {
      setResult(await get<ButcheryResult>(`/butchery-batches/${id}`));
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  const label = (productId: number) => {
    const name = productNames.get(productId);
    return name ? `${name} (#${productId})` : `#${productId}`;
  };

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">KESİM &amp; VERİM ANALİZİ</div>
          <h1>Kuzu Parçalama</h1>
          <p>Giriş ağırlığından parça ürünlere kadar maliyet dağıtımını ve ağırlık dengesini izleyin.</p>
        </div>
        <button className="btn btn-primary" type="button">＋ Parçalama kaydı</button>
      </div>

      <div className="metrics">
        <div className="metric">
          <div className="metric-top">Giriş ağırlığı <span className="metric-icon">♧</span></div>
          <strong>{result ? `${result.batch.input_weight} kg` : "—"}</strong>
          <div className="metric-foot">{result ? result.batch.animal_type : ""}</div>
        </div>
        <div className="metric">
          <div className="metric-top">Dağıtılacak maliyet <span className="metric-icon">€</span></div>
          <strong>€ {result ? result.batch.total_cost.toFixed(2) : "—"}</strong>
          <div className="metric-foot">{result ? `${result.batch.input_weight} kg × €${result.batch.unit_cost}` : ""}</div>
        </div>
        <div className="metric">
          <div className="metric-top">Fire <span className="metric-icon" style={{ background: "var(--amber-bg)", color: "var(--amber)" }}>↓</span></div>
          <strong>{result && !result.weight_balance_ok ? `${result.fire_weight} kg` : "0 kg"}</strong>
          <div className="metric-foot">Ağırlık dengesi</div>
        </div>
        <div className="metric">
          <div className="metric-top">Maliyet kontrolü <span className="metric-icon">✓</span></div>
          <strong>{result && result.cost_mismatch ? "Uyuşmazlık" : "Tutarlı"}</strong>
          <div className="metric-foot">Beklenen € {result ? result.expected_cost.toFixed(2) : "—"}</div>
        </div>
      </div>

      {result && !result.weight_balance_ok && (
        <div className="recall-banner">
          <span>ⓘ</span>
          <div>
            <strong>Ağırlık dengesi kontrolü</strong>
            <p>
              Parça ağırlıkları toplamı {result.batch.input_weight} kg giriş ağırlığından
              {result.fire_weight} kg eksik. Fire / yan ürün kaydı doğrulanmalı.
            </p>
          </div>
        </div>
      )}

      <div className="panel">
        <div className="panel-head">
          <h2>Parça ve maliyet dağıtımı</h2>
          <form className="inline-form" onSubmit={load}>
            <input
              value={batchId}
              onChange={(event) => setBatchId(event.target.value)}
              placeholder="Parti ID"
              style={{ width: 120 }}
            />
            <button className="btn btn-secondary" type="submit">Yükle</button>
          </form>
        </div>
        {error && <p className="muted" style={{ padding: "0 20px" }}>{error}</p>}
        {result && (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Ürün</th>
                  <th>Adet</th>
                  <th>Ağırlık</th>
                  <th>Yan ürün</th>
                  <th>Dağıtılan maliyet</th>
                </tr>
              </thead>
              <tbody>
                {result.items.map((item, index) => (
                  <tr key={`${item.product_id}-${index}`}>
                    <td>{label(item.product_id)}</td>
                    <td>{item.quantity}</td>
                    <td>{item.weight} kg</td>
                    <td>{item.is_waste ? "Evet" : "Hayır"}</td>
                    <td>€ {item.allocated_cost.toFixed(2)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </section>
  );
}
