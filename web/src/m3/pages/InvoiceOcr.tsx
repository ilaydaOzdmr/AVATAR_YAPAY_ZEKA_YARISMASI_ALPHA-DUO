import { useCallback, useEffect, useState, type FormEvent } from "react";
import { get, post } from "../../api";

interface Material {
  id: number;
  name: string;
}

interface InvoiceItem {
  product_name: string;
  qty: number;
  unit: string;
  lot_code: string;
  unit_cost: number;
}

export default function InvoiceOcr() {
  const [text, setText] = useState("");
  const [items, setItems] = useState<InvoiceItem[]>([]);
  const [materials, setMaterials] = useState<Material[]>([]);
  const [feedback, setFeedback] = useState<string | null>(null);
  const [saved, setSaved] = useState<Set<number>>(new Set());

  const loadMaterials = useCallback(() => {
    get<Material[]>("/materials").then(setMaterials).catch(() => {});
  }, []);

  useEffect(() => {
    loadMaterials();
  }, [loadMaterials]);

  async function parse(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFeedback(null);
    setItems([]);
    setSaved(new Set());
    if (!text.trim()) {
      setFeedback("Fatura/irsaliye metni girin.");
      return;
    }
    try {
      const result = await post<{ items: InvoiceItem[] }>("/ocr/invoice", { text });
      setItems(result.items);
    } catch (err) {
      setFeedback(err instanceof Error ? err.message : String(err));
    }
  }

  async function approve(index: number) {
    const item = items[index];
    setFeedback(null);
    // Resolve or create the material card, then create a material lot.
    let materialId = materials.find((m) =>
      m.name.toLowerCase().includes(item.product_name.toLowerCase()) || item.product_name.toLowerCase().includes(m.name.toLowerCase()),
    )?.id;

    if (!materialId) {
      try {
        const created = await post<{ id: number }>("/materials", {
          code: `OCR-${Date.now()}`,
          name: item.product_name,
          unit: item.unit,
          is_raw: true,
          default_purchase_price: item.unit_cost,
        });
        materialId = created.id;
        loadMaterials();
      } catch (err) {
        setFeedback(err instanceof Error ? err.message : String(err));
        return;
      }
    }

    try {
      await post("/lots", {
        code: item.lot_code || `OCR-LOT-${Date.now()}`,
        item_type: "material",
        item_id: materialId,
        qty: item.qty,
        unit: item.unit,
        purchase_price: item.unit_cost,
        unit_cost: item.unit_cost,
        status: "active",
      });
      setSaved((prev) => new Set(prev).add(index));
      setFeedback(`"${item.product_name}" stok kaydına işlendi.`);
    } catch (err) {
      setFeedback(err instanceof Error ? err.message : String(err));
    }
  }

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">BELGE OKUMA</div>
          <h1>Fatura / İrsaliye OCR</h1>
          <p>Fatura veya irsaliye metnini yapıştırın; ürün, miktar, lot ve maliyet önerilerini onaylayarak stok kaydı oluşturun.</p>
        </div>
      </div>

      <div className="panel" style={{ marginBottom: 16 }}>
        <div className="panel-head"><h2>Metin girişi</h2></div>
        <div style={{ padding: 20 }}>
          <form className="stack-form" onSubmit={parse}>
            <textarea
              value={text}
              onChange={(e) => setText(e.target.value)}
              rows={5}
              placeholder={"Örn:\nEt (dana) 100 kg ET-2026-099 9,70 TL/kg\nBaharat karışımı 2 kg BAH-2026-024 24,00 TL/kg"}
              style={{ width: "100%", minHeight: 110, padding: 12, border: "1px solid var(--line)", borderRadius: 10, font: "inherit" }}
            />
            <button className="btn btn-primary" type="submit">Metni ayrıştır</button>
          </form>
        </div>
      </div>

      {feedback && <p className="muted" style={{ marginBottom: 12 }}>{feedback}</p>}

      {items.length > 0 && (
        <div className="panel">
          <div className="panel-head"><h2>Önerilen kalemler</h2></div>
          <div className="table-wrap">
            <table>
              <thead>
                <tr><th>Ürün</th><th>Miktar</th><th>Lot</th><th>Birim maliyet</th><th>Aksiyon</th></tr>
              </thead>
              <tbody>
                {items.map((item, index) => (
                  <tr key={index}>
                    <td><b>{item.product_name || "—"}</b></td>
                    <td>{item.qty} {item.unit}</td>
                    <td className="id">{item.lot_code || "—"}</td>
                    <td>€ {item.unit_cost.toFixed(2)}</td>
                    <td>
                      <button
                        className="btn btn-secondary"
                        type="button"
                        disabled={saved.has(index)}
                        onClick={() => approve(index)}
                      >
                        {saved.has(index) ? "✓ Kaydedildi" : "Onayla ve kaydet"}
                      </button>
                    </td>
                  </tr>
                ))}
                {items.length === 0 && <tr><td colSpan={5} className="muted">Kalem bulunamadı.</td></tr>}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </section>
  );
}
