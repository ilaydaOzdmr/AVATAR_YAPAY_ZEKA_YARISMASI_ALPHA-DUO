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
  expected_cost: number;
  cost_difference: number;
  cost_check_status: string;
  weight_check_status: string;
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
  weight_check_status: string;
  part_weight: number;
  fire_weight: number;
  weight_diff: number;
  cost_mismatch: boolean;
  cost_check_status: string;
  expected_cost: number;
  recorded_cost: number;
  cost_difference: number;
  cost_balance_ok: boolean;
  cost_diff: number;
}

interface ItemDraft {
  product_id: string;
  weight: string;
  quantity: string;
  is_waste: boolean;
}

const weightFormatter = new Intl.NumberFormat("tr-TR", { maximumFractionDigits: 3 });
const moneyFormatter = new Intl.NumberFormat("tr-TR", {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
});

const formatWeight = (value: number) => weightFormatter.format(value);
const formatMoney = (value: number) => moneyFormatter.format(value);

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
  const [batchDate, setBatchDate] = useState(new Date().toISOString().slice(0, 10));
  const [totalCost, setTotalCost] = useState("");
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
        quantity: Number.parseInt(item.quantity, 10) || 0,
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
          total_cost: Number.parseFloat(totalCost) || weight * cost,
          batch_date: batchDate,
          status: "draft",
        },
        items: parsedItems,
      });
      setCode("");
      setInputWeight("");
      setUnitCost("");
      setTotalCost("");
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
                <input value={batchDate} onChange={(e) => setBatchDate(e.target.value)} type="date" placeholder="Tarih" style={{ width: 150 }} />
                <input value={totalCost} onChange={(e) => setTotalCost(e.target.value)} placeholder="Toplam maliyet (ops.)" style={{ width: 150 }} />
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
                      style={{ width: 90 }}
                    />
                    <input
                      value={item.quantity}
                      onChange={(e) => setItems((prev) => prev.map((it, i) => (i === index ? { ...it, quantity: e.target.value } : it)))}
                      placeholder="Adet"
                      style={{ width: 70 }}
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
                <button className="btn btn-secondary" type="button" onClick={() => setItems((prev) => [...prev, { product_id: "", weight: "", quantity: "", is_waste: false }])}>
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
              {result.fire_weight > 0 && (
                <div className="recall-banner" style={{ margin: 16 }}>
                  <span>ⓘ</span>
                  <div>
                    <strong>Fire tespiti</strong>
                    <p>{formatWeight(result.fire_weight)} kg fire tespit edildi.</p>
                  </div>
                </div>
              )}

              <div className={`analysis-card ${result.weight_check_status === "ok" ? "ok" : "warn"}`}>
                <div className="analysis-head">
                  <strong>Fire ve Tutarlılık Analizi</strong>
                  <span className={`status ${result.weight_check_status === "ok" ? "ok" : "warn"}`}>
                    {result.weight_check_status === "ok" ? "Tutarlı" : "Uyarı"}
                  </span>
                </div>
                <div className="analysis-grid">
                  <div>
                    <span>Giriş ağırlığı</span>
                    <b>{formatWeight(result.batch.input_weight)} kg</b>
                  </div>
                  <div>
                    <span>Parça toplamı</span>
                    <b>{formatWeight(result.part_weight)} kg</b>
                  </div>
                  <div>
                    <span>Hesaplanan fire</span>
                    <b>{formatWeight(result.fire_weight)} kg</b>
                  </div>
                </div>
                <p className="analysis-note">
                  Ürün giriş ağırlığı ile elde edilen parçaların toplamı karşılaştırılmıştır.{" "}
                  {formatWeight(result.batch.input_weight)} kg girişten {formatWeight(result.part_weight)} kg parça elde edilmiş,
                  aradaki {formatWeight(result.fire_weight)} kg fark fire olarak hesaplanmıştır.
                </p>
              </div>

              <div className={`analysis-card ${result.cost_check_status === "ok" ? "ok" : "warn"}`}>
                <div className="analysis-head">
                  <strong>Maliyet Kontrolü</strong>
                  <span className={`status ${result.cost_check_status === "ok" ? "ok" : "warn"}`}>
                    {result.cost_check_status === "ok" ? "Tutarlı" : "Tutarsız"}
                  </span>
                </div>
                <div className="analysis-grid">
                  <div>
                    <span>Beklenen maliyet</span>
                    <b>€ {formatMoney(result.expected_cost)}</b>
                  </div>
                  <div>
                    <span>Kayıtlı maliyet</span>
                    <b>€ {formatMoney(result.recorded_cost)}</b>
                  </div>
                  <div>
                    <span>Maliyet farkı</span>
                    <b>€ {formatMoney(result.cost_difference)}</b>
                  </div>
                </div>
                <p className="analysis-note">
                  Maliyet kontrolünde ise {formatWeight(result.batch.input_weight)} kg × {formatMoney(result.batch.unit_cost)} € ={" "}
                  {formatMoney(result.expected_cost)} € beklenirken veritabanında {formatMoney(result.recorded_cost)} € kayıtlı olduğu görülmüştür.{" "}
                  {result.cost_check_status === "ok"
                    ? "Beklenen ve kayıtlı maliyet tutarlıdır."
                    : `Aradaki ${formatMoney(result.cost_difference)} € fark maliyet tutarsızlığı olarak belirlenmiştir.`}
                </p>
              </div>

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
