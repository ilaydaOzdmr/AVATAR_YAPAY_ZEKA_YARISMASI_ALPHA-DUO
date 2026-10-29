import { useState, type ComponentType } from "react";
import Dashboard from "../pages/Dashboard";
import Materials from "../pages/Materials";
import Products from "../pages/Products";
import ProductionOrders from "../pages/ProductionOrders";
import Butchery from "../pages/Butchery";
import Traceability from "../pages/Traceability";
import Recall from "../pages/Recall";
import Predictions from "../pages/Predictions";

interface NavItem {
  key: string;
  label: string;
  Component: ComponentType;
}

const navItems: NavItem[] = [
  { key: "dashboard", label: "Dashboard", Component: Dashboard },
  { key: "materials", label: "Materials", Component: Materials },
  { key: "products", label: "Products", Component: Products },
  { key: "production-orders", label: "Production orders", Component: ProductionOrders },
  { key: "butchery", label: "Butchery", Component: Butchery },
  { key: "traceability", label: "Traceability", Component: Traceability },
  { key: "recall", label: "Recall", Component: Recall },
  { key: "predictions", label: "Predictions", Component: Predictions },
];

export default function App() {
  const [activeKey, setActiveKey] = useState("dashboard");
  const active = navItems.find((item) => item.key === activeKey) ?? navItems[0];
  const ActiveComponent = active.Component;

  return (
    <div className="m3-app">
      <header className="m3-top-bar">
        <span className="m3-top-bar__title">AVATAR Admin</span>
        <span className="m3-top-bar__subtitle">Material 3</span>
      </header>
      <div className="m3-body">
        <nav className="m3-nav" aria-label="Main navigation">
          {navItems.map((item) => (
            <button
              key={item.key}
              type="button"
              className={item.key === activeKey ? "m3-nav__item active" : "m3-nav__item"}
              onClick={() => setActiveKey(item.key)}
            >
              {item.label}
            </button>
          ))}
        </nav>
        <main className="m3-content">
          <div className="m3-page">
            <h1 className="m3-page__title">{active.label}</h1>
            <ActiveComponent />
          </div>
        </main>
      </div>
    </div>
  );
}
