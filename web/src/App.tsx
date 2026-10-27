import { useState } from "react";
import Dashboard from "./pages/Dashboard";
import Materials from "./pages/Materials";
import Products from "./pages/Products";
import ProductionOrders from "./pages/ProductionOrders";
import Butchery from "./pages/Butchery";
import Traceability from "./pages/Traceability";
import Recall from "./pages/Recall";
import Predictions from "./pages/Predictions";

const pages = [
  { key: "dashboard", label: "Dashboard" },
  { key: "materials", label: "Materials" },
  { key: "products", label: "Products" },
  { key: "production-orders", label: "Production Orders" },
  { key: "butchery", label: "Butchery" },
  { key: "traceability", label: "Traceability" },
  { key: "recall", label: "Recall" },
  { key: "predictions", label: "Predictions" },
] as const;

type PageKey = (typeof pages)[number]["key"];

export default function App() {
  const [page, setPage] = useState<PageKey>("dashboard");

  return (
    <div className="app">
      <header className="app-header">
        <h1>AVATAR Admin</h1>
        <nav className="nav">
          {pages.map((item) => (
            <button
              key={item.key}
              type="button"
              className={page === item.key ? "nav-button active" : "nav-button"}
              onClick={() => setPage(item.key)}
            >
              {item.label}
            </button>
          ))}
        </nav>
      </header>
      <main className="content">
        {page === "dashboard" && <Dashboard />}
        {page === "materials" && <Materials />}
        {page === "products" && <Products />}
        {page === "production-orders" && <ProductionOrders />}
        {page === "butchery" && <Butchery />}
        {page === "traceability" && <Traceability />}
        {page === "recall" && <Recall />}
        {page === "predictions" && <Predictions />}
      </main>
    </div>
  );
}
