import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { get } from "../api";
import Dashboard from "./pages/Dashboard";
import Inventory from "./pages/Inventory";
import Production from "./pages/Production";
import Butchery from "./pages/Butchery";
import Sales from "./pages/Sales";
import Traceability from "./pages/Traceability";
import Recall from "./pages/Recall";
import Predictions from "./pages/Predictions";
import MobilePreview from "./pages/MobilePreview";
import Reports from "./pages/Reports";
import InvoiceOcr from "./pages/InvoiceOcr";
import Alerts from "./pages/Alerts";

interface NavItem {
  key: string;
  label: string;
  icon: string;
  badge?: number;
}

const groups: { label: string; items: NavItem[] }[] = [
  {
    label: "Genel",
    items: [
      { key: "dashboard", label: "Genel Bakış", icon: "◫" },
      { key: "inventory", label: "Stok & Hammaddeler", icon: "▤" },
      { key: "production", label: "Üretim Emirleri", icon: "⚙" },
      { key: "butchery", label: "Kuzu Parçalama", icon: "♧" },
      { key: "sales", label: "Satış & Sevkiyat", icon: "↗" },
    ],
  },
  {
    label: "Kontrol Merkezi",
    items: [
      { key: "trace", label: "Lot İzlenebilirliği", icon: "⌘" },
      { key: "recall", label: "Geri Çağırma", icon: "⚠" },
      { key: "alerts", label: "Uyarılar", icon: "▣" },
      { key: "ai", label: "Yapay Zekâ", icon: "✳" },
      { key: "mobile", label: "Mobil Uygulama", icon: "▯" },
      { key: "reports", label: "Raporlar", icon: "▥" },
      { key: "invoice-ocr", label: "Fatura OCR", icon: "⌗" },
    ],
  },
];

const titles: Record<string, string> = {
  dashboard: "Genel Bakış",
  inventory: "Stok & Hammaddeler",
  production: "Üretim Emirleri",
  butchery: "Kuzu Parçalama",
  sales: "Satış & Sevkiyat",
  trace: "Lot İzlenebilirliği",
  recall: "Geri Çağırma Merkezi",
  alerts: "Uyarılar",
  ai: "Yapay Zekâ Merkezi",
  mobile: "Mobil Uygulama",
  reports: "Raporlar",
  "invoice-ocr": "Fatura OCR",
};

function renderView(key: string, onRecallChanged: () => void, onOpenAlerts: () => void, onAlertsChanged: () => void) {
  switch (key) {
    case "inventory":
      return <Inventory />;
    case "production":
      return <Production />;
    case "butchery":
      return <Butchery />;
    case "sales":
      return <Sales />;
    case "trace":
      return <Traceability />;
    case "recall":
      return <Recall onChanged={onRecallChanged} onOpenAlerts={onOpenAlerts} />;
    case "alerts":
      return <Alerts onChanged={onAlertsChanged} />;
    case "ai":
      return <Predictions />;
    case "mobile":
      return <MobilePreview />;
    case "reports":
      return <Reports />;
    case "invoice-ocr":
      return <InvoiceOcr />;
    case "dashboard":
    default:
      return <Dashboard onOpenAlerts={onOpenAlerts} />;
  }
}

export default function App() {
  const [view, setView] = useState("dashboard");
  const [collapsed, setCollapsed] = useState(false);
  const [query, setQuery] = useState("");
  const [toast, setToast] = useState("");
  const [blockedLotCount, setBlockedLotCount] = useState(0);
  const [unreadAlertCount, setUnreadAlertCount] = useState(0);
  const toastTimer = useRef<number | undefined>(undefined);

  // The sidebar/bell badges must reflect the backend's canonical blocked-lot
  // list and the persisted unread warning state.
  const refreshAlerts = useCallback(() => {
    get<{ blocked_lot_codes?: string[] }>("/sync/pull")
      .then((data) => setBlockedLotCount(data.blocked_lot_codes?.length ?? 0))
      .catch(() => {});
    get<{ read: boolean }[]>("/anomalies")
      .then((list) => setUnreadAlertCount(list.filter((item) => !item.read).length))
      .catch(() => {});
  }, []);

  const openAlerts = useCallback(() => {
    setView("alerts");
    window.scrollTo({ top: 0, behavior: "smooth" });
  }, []);

  useEffect(() => {
    refreshAlerts();
  }, [refreshAlerts, view]);

  const activeLabel = titles[view] ?? "Genel Bakış";
  const viewEl = useMemo(
    () => renderView(view, refreshAlerts, openAlerts, refreshAlerts),
    [view, refreshAlerts, openAlerts],
  );

  function showToast(message: string) {
    setToast(message);
    window.clearTimeout(toastTimer.current);
    toastTimer.current = window.setTimeout(() => setToast(""), 2600);
  }

  function navigate(next: string) {
    setView(next);
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function handleSearch() {
    const q = query.toLowerCase();
    if (!q) return;
    if (q.includes("et-") || q.includes("suc-") || q.includes("lot")) navigate("trace");
    else if (q.includes("urt") || q.includes("üret")) navigate("production");
    else if (q.includes("market") || q.includes("sev")) navigate("sales");
    else navigate("inventory");
    showToast("Arama için ilgili bölüm açıldı.");
  }

  return (
    <div className="app">
      <aside className={collapsed ? "sidebar collapsed" : "sidebar"} id="sidebar">
        <div className="brand">
          <div className="logo">A</div>
          <div>
            <strong>AVATAR</strong>
            <small>FOOD INTELLIGENCE</small>
          </div>
        </div>
        {groups.map((group) => (
          <div key={group.label}>
            <div className="nav-label">{group.label}</div>
            <nav className="nav">
              {group.items.map((item) => (
                <button
                  key={item.key}
                  type="button"
                  className={view === item.key ? "active" : ""}
                  onClick={() => navigate(item.key)}
                >
                  <span className="ico">{item.icon}</span>
                  {item.label}
                  {item.key === "recall" && blockedLotCount > 0 ? <span className="badge">{blockedLotCount}</span> : null}
                  {item.key === "alerts" && unreadAlertCount > 0 ? <span className="badge">{unreadAlertCount}</span> : null}
                  {item.key !== "recall" && item.key !== "alerts" && item.badge ? <span className="badge">{item.badge}</span> : null}
                </button>
              ))}
            </nav>
          </div>
        ))}
        <div className="sidebar-bottom">
          <div className="profile">
            <div className="avatar">AY</div>
            <div>
              <b>Avatar Yönetici</b>
              <small>İşletme yöneticisi</small>
            </div>
            <span style={{ marginLeft: "auto" }}>⌄</span>
          </div>
        </div>
      </aside>

      <main className="main">
        <header className="topbar">
          <button
            className="mobile-menu"
            type="button"
            aria-label="Menü"
            onClick={() => setCollapsed((open) => !open)}
          >
            ☰
          </button>
          <div className="crumb">
            AVATAR <span style={{ padding: "0 8px", color: "#b8c1bb" }}>/</span>{" "}
            <b>{activeLabel}</b>
          </div>
          <div className="top-actions">
            <div className="search">
              <span>⌕</span>
              <input
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === "Enter") handleSearch();
                }}
                placeholder="Lot, ürün veya işlem ara..."
              />
            </div>
            <button
              className="icon-btn"
              type="button"
              title={unreadAlertCount > 0 ? `Bildirimler (${unreadAlertCount})` : "Bildirimler"}
              onClick={openAlerts}
            >
              ♧
              {unreadAlertCount > 0 && <span className="badge">{unreadAlertCount}</span>}
            </button>
            <button className="icon-btn" type="button" title="Yardım" onClick={() => showToast("Yardım merkezi.")}>
              ?
            </button>
          </div>
        </header>

        <div className="content">{viewEl}</div>
      </main>

      <div className={toast ? "toast show" : "toast"}>{toast}</div>
    </div>
  );
}
