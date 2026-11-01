import { useCallback, useEffect, useState } from "react";
import { get, post } from "../../api";

interface Anomaly {
  key: string;
  type: string;
  severity: string;
  message: string;
  ref: string;
  read: boolean;
}

const TYPE_LABELS: Record<string, string> = {
  fire: "Fire",
  cost: "Maliyet",
  yield: "Verim",
  stock: "Stok",
};

export default function Alerts({ onChanged }: { onChanged?: () => void }) {
  const [items, setItems] = useState<Anomaly[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const load = useCallback(() => {
    get<Anomaly[]>("/anomalies").then(setItems).catch((err) => setError(String(err)));
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const unread = items.filter((item) => !item.read);
  const critical = unread.filter((item) => item.severity === "critical").length;

  async function markRead(keys: string[]) {
    if (!keys.length || busy) return;
    setBusy(true);
    setError(null);
    try {
      await post("/anomalies/read", { keys });
      load();
      onChanged?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">KALİTE &amp; GÜVENLİK</div>
          <h1>Uyarılar</h1>
          <p>Fire, verim, maliyet ve stok kurallarından üretilen uyarıları inceleyin; okundu olarak işaretleyin.</p>
        </div>
        <button
          className="btn btn-primary"
          type="button"
          disabled={busy || unread.length === 0}
          onClick={() => void markRead(unread.map((item) => item.key))}
        >
          ✓ Tümünü okundu işaretle
        </button>
      </div>

      {error && <p className="muted" style={{ marginBottom: 12 }}>{error}</p>}

      <div className="metrics">
        <div className="metric">
          <div className="metric-top">Toplam uyarı <span className="metric-icon">⚠</span></div>
          <strong>{items.length}</strong>
          <div className="metric-foot">Kural tabanlı</div>
        </div>
        <div className="metric">
          <div className="metric-top">Okunmamış <span className="metric-icon" style={{ background: "var(--amber-bg)", color: "var(--amber)" }}>●</span></div>
          <strong>{unread.length}</strong>
          <div className="metric-foot">Bildirim rozetinde gösterilir</div>
        </div>
        <div className="metric">
          <div className="metric-top">Kritik <span className="metric-icon" style={{ background: "var(--red-bg)", color: "var(--red)" }}>!</span></div>
          <strong>{critical}</strong>
          <div className="metric-foot">Okunmamış kritik uyarı</div>
        </div>
      </div>

      <div className="panel">
        <div className="panel-head">
          <div>
            <h2>Uyarı listesi</h2>
            <p>Okundu olarak işaretlenen uyarılar listede soluk görünür ve rozet sayılmaz.</p>
          </div>
          <span className={unread.length ? "status warn" : "status ok"}>
            {unread.length ? `${unread.length} okunmamış` : "Tümü okundu"}
          </span>
        </div>
        <div className="alert-list">
          {items.map((item) => (
            <div className={item.read ? "alert read" : "alert"} key={item.key}>
              <div className={`alert-symbol ${item.severity === "critical" ? "red" : "amber"}`}>
                {item.severity === "critical" ? "⚠" : "↗"}
              </div>
              <div style={{ flex: 1 }}>
                <b>{item.message} · {item.ref}</b>
                <p>
                  {TYPE_LABELS[item.type] ?? item.type} · {item.severity === "critical" ? "Kritik" : "Uyarı"}
                  {item.read ? " · Okundu" : ""}
                </p>
              </div>
              {!item.read && (
                <button className="btn btn-secondary" type="button" disabled={busy} onClick={() => void markRead([item.key])}>
                  Okundu
                </button>
              )}
            </div>
          ))}
          {items.length === 0 && <p className="muted">Açık uyarı bulunmuyor.</p>}
        </div>
      </div>
    </section>
  );
}
