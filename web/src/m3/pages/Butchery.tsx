import { useCallback, useEffect, useState, type FormEvent } from "react";
import { get, post } from "../../api";

interface Product {
  id: number;
  name: string;
}

interface ButcheryBatch {
  id: number;
  code: string;
  animal_type: string;
  input_weight: number;
  unit_cost: number;
  total_cost: number;
  batch_date: string;
  status: string;
}

interface ButcheryItemCost {
  product_id: number;
  weight: number;
  quantity: number;
  is_waste: boolean;
  allocated_cost: number;
  allocated_cost_value: number;
}

interface ButcheryResult {
  batch: ButcheryBatch;
  items: ButcheryItemCost[];
  weight_balance_ok: boolean;
  fire_weight: number;
  cost_mismatch: boolean;
  expected_cost: number;
}

interface ItemDraft {
  product_id: string;
  weight: string;
  is_waste: boolean;
}

export default function Butchery() {
  const [batches, setBatches] = useState<ButcheryBatch[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [selectedId, setSelectedId] = useState("");
  const [result, setResult] = useState<ButcheryResult | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [feedback, setFeedback] = useState<string | null>(null);
  const [allocationKey, setAllocationKey] = useState<"weight" | "value">("weight");

  const [code, setCode] = useState("");
  const [animalType, setAnimalType] = useState("kuzu");
  const [inputWeight, setInputWeight] = useState("");
  const [unitCost, setUnitCost] = useState("");
  const [items, setItems] = useState<ItemDraft[]>([]);

  const load = useCallback(() => {
    get<ButcheryBatch[]>("/butchery-batches").then(setBatches).catch(() => {});
    get<Product[]>("/products").then(setProducts).catch(() => {});
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (!selectedId) {
      setResult(null);
      return;
    }
    get<ButcheryResult>(`/butchery-batches/${selectedId}`).then(setResult).catch(() => setResult(null));
  }, [selectedId]);

  const productName = (id: number) => products.find((p) => p.id === id)?.name ?? `#${id}`;

  async function createBatch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFeedback(null);
    const weight = Number.parseFloat(inputWeight);
    const cost = Number.parseFloat(unitCost);
    if (!code || Number.isNaN(weight) || Number.isNaN(cost)) {
      setFeedback("Kod, giriş ağırlığı ve birim maliyet zorunlu.");
      return;
    }
    const parsedItems = items
      .map((item) => ({
        product_id: Number.parseInt(item.product_id, 10),
        weight: Number.parseFloat(item.weight) || 0,
        is_waste: item.is_waste,
      }))
      .filter((item) => !Number.isNaN(item.product_id));
    try {
      await post("/butchery-batches", {
        batch: {
          code,
          animal_type: animalType,
          input_weight: weight,
          unit_cost: cost,
          total_cost: weight * cost,
          batch_date: new Date().toISOString().slice(0, 10),
          status: "draft",
        },
        items: parsedItems,
      });
      setCode("");
      setInputWeight("");
      setUnitCost("");
      setItems([]);
      setShowForm(false);
      setFeedback("Parçalama kaydı oluşturuldu.");
      load();
    } catch (err) {
      setFeedback(err instanceof Error ? err.message : String(err));
    }
  }

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">KESİM &amp; VERİM ANALİZİ</div>
          <h1>Kuzu Parçalama</h1>
          <p>Giriş ağırlığından parça ürünlere kadar maliyet dağıtımını ve ağırlık dengesini izleyin.</p>
        </div>
        <button className="btn btn-primary" type="button" onClick={() => setShowForm((v) => !v)}>＋ Parçalama kaydı</button>
      </div>

      {showForm && (
        <div className="panel" style={{ marginBottom: 16 }}>
          <div className="panel-head"><h2>Yeni parçalama kaydı</h2></div>
          <div style={{ padding: 20 }}>
            <form className="stack-form" onSubmit={createBatch}>
              <div className="inline-form">
                <input value={code} onChange={(e) => setCode(e.target.value)} placeholder="Kod (KUZ-...)" required />
                <input value={animalType} onChange={(e) => setAnimalType(e.target.value)} placeholder="Hayvan türü" style={{ width: 120 }} />
                <input value={inputWeight} onChange={(e) => setInputWeight(e.target.value)} placeholder="Giriş ağırlığı" style={{ width: 120 }} />
                <input value={unitCost} onChange={(e) => setUnitCost(e.target.value)} placeholder="Birim maliyet" style={{ width: 120 }} />
              </div>
              <div>
                {items.map((item, index) => (
                  <div key={index} className="inline-form" style={{ marginBottom: 8 }}>
                    <select
                      value={item.product_id}
                      onChange={(e) => setItems((prev) => prev.map((it, i) => (i === index ? { ...it, product_id: e.target.value } : it)))}
                    >
                      <option value="">Ürün...</option>
                      {products.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
                    </select>
                    <input
                      value={item.weight}
                      onChange={(e) => setItems((prev) => prev.map((it, i) => (i === index ? { ...it, weight: e.target.value } : it)))}
                      placeholder="Ağırlık"
                      style={{ width: 100 }}
                    />
                    <label style={{ flexDirection: "row", alignItems: "center", gap: 6, fontSize: 12 }}>
                      <input
                        type="checkbox"
                        checked={item.is_waste}
                        onChange={(e) => setItems((prev) => prev.map((it, i) => (i === index ? { ...it, is_waste: e.target.checked } : it)))}
                        style={{ width: "auto", height: "auto" }}
                      />
                      Yan ürün
                    </label>
                    <button className="btn btn-secondary" type="button" onClick={() => setItems((prev) => prev.filter((_, i) => i !== index))}>✕</button>
                  </div>
                ))}
                <button className="btn btn-secondary" type="button" onClick={() => setItems((prev) => [...prev, { product_id: "", weight: "", is_waste: false }])}>
                  ＋ Parça ekle
                </button>
              </div>
              <button className="btn btn-primary" type="submit">Kaydet</button>
            </form>
          </div>
        </div>
      )}

      {feedback && <p className="muted" style={{ marginBottom: 12 }}>{feedback}</p>}

      <div className="two-col">
        <div className="panel">
          <div className="panel-head"><h2>Parçalama kayıtları</h2></div>
          <div className="table-wrap">
            <table>
              <thead>
                <tr><th>Parti</th><th>Hayvan</th><th>Giriş</th><th>Maliyet</th><th>Durum</th></tr>
              </thead>
              <tbody>
                {batches.map((batch) => (
                  <tr key={batch.id} style={{ cursor: "pointer" }} onClick={() => setSelectedId(String(batch.id))}>
                    <td className="id">{batch.code}</td>
                    <td>{batch.animal_type}</td>
                    <td>{batch.input_weight} kg</td>
                    <td>€ {batch.total_cost.toFixed(2)}</td>
                    <td><span className="status progress">{batch.status}</span></td>
                  </tr>
                ))}
                {batches.length === 0 && <tr><td colSpan={5} className="muted">Parçalama kaydı bulunamadı.</td></tr>}
              </tbody>
            </table>
          </div>
        </div>

        <div className="panel">
          <div className="panel-head">
            <h2>Parça ve maliyet dağıtımı</h2>
            <select className="tiny-select" value={allocationKey} onChange={(e) => setAllocationKey(e.target.value as "weight" | "value")}>
              <option value="weight">Miktar (ağırlık) bazlı</option>
              <option value="value">Değer ağırlıklı</option>
            </select>
          </div>
          {result ? (
            <>
              {!result.weight_balance_ok && (
                <div className="recall-banner" style={{ margin: 16 }}>
                  <span>ⓘ</span>
                  <div>
                    <strong>Ağırlık dengesi</strong>
                    <p>{result.fire_weight} kg fire tespit edildi.</p>
                  </div>
                </div>
              )}
              {result.cost_mismatch && (
                <div className="recall-banner" style={{ margin: 16 }}>
                  <span>⚠</span>
                  <div>
                    <strong>Maliyet uyuşmazlığı</strong>
                    <p>Beklenen € {result.expected_cost.toFixed(2)} ≠ € {result.batch.total_cost.toFixed(2)}</p>
                  </div>
                </div>
              )}
              <div className="table-wrap">
                <table>
                  <thead>
                    <tr><th>Ürün</th><th>Adet</th><th>Ağırlık</th><th>Yan ürün</th><th>Maliyet</th></tr>
                  </thead>
                  <tbody>
                    {result.items.map((item, index) => (
                      <tr key={`${item.product_id}-${index}`}>
                        <td>{productName(item.product_id)}</td>
                        <td>{item.quantity}</td>
                        <td>{item.weight} kg</td>
                        <td>{item.is_waste ? "Evet" : "Hayır"}</td>
                        <td>€ {(allocationKey === "value" ? item.allocated_cost_value : item.allocated_cost).toFixed(2)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </>
          ) : (
            <p className="muted" style={{ padding: 20 }}>Detay için bir kayıt seçin.</p>
          )}
        </div>
      </div>
    </section>
  );
}
