import { useEffect, useState } from "react";
import { get } from "../../api";

interface Lot {
  id: number;
  code: string;
  item_type: string;
  qty: number;
  unit: string;
  unit_cost: number;
}

interface CostSummary {
  ekp: number;
  vkp: number;
  gross_profit: number;
  margin: number;
}

export default function Dashboard() {
  const [stockValue, setStockValue] = useState(0);
  const [ekp, setEkp] = useState(0);

  useEffect(() => {
    get<Lot[]>("/lots")
      .then((lots) => {
        const value = lots
          .filter((lot) => lot.item_type === "material")
          .reduce((sum, lot) => sum + lot.qty * lot.unit_cost, 0);
        setStockValue(value);
      })
      .catch(() => {
        /* keep 0 on failure */
      });

    get<CostSummary>("/production-orders/1")
      .then((summary) => setEkp(summary.ekp))
      .catch(() => {
        /* keep 0 on failure */
      });
  }, []);

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">İŞLETME ÖZETİ</div>
          <h1>Günaydın, yönetici 👋</h1>
          <p>İşletmenizin bugünkü operasyon özeti ve dikkat gerektiren noktalar.</p>
        </div>
        <button className="btn btn-primary" type="button">
          ＋ Yeni üretim emri
        </button>
      </div>

      <div className="metrics">
        <div className="metric">
          <div className="metric-top">
            Toplam stok değeri <span className="metric-icon">▤</span>
          </div>
          <strong>€ {stockValue.toLocaleString("tr-TR", { minimumFractionDigits: 2 })}</strong>
          <div className="metric-foot">Hammadde lotlarına göre</div>
        </div>
        <div className="metric">
          <div className="metric-top">
            Bugünkü üretim <span className="metric-icon">⚙</span>
          </div>
          <strong>
            85 <small style={{ fontSize: 13, fontWeight: 600 }}>paket</small>
          </strong>
          <div className="metric-foot">
            <span className="positive">URT-2026-045</span> · tamamlandı
          </div>
        </div>
        <div className="metric">
          <div className="metric-top">
            Birim gerçek maliyet <span className="metric-icon">€</span>
          </div>
          <strong>€ {ekp.toFixed(2)}</strong>
          <div className="metric-foot">Narin Sucuk · EKP</div>
        </div>
        <div className="metric">
          <div className="metric-top">
            Kritik uyarı <span className="metric-icon" style={{ background: "var(--amber-bg)", color: "var(--amber)" }}>⚠</span>
          </div>
          <strong>
            01 <small style={{ fontSize: 13, fontWeight: 600 }}>uyarı</small>
          </strong>
          <div className="metric-foot">
            <span className="warning">İnceleme gerekli</span> · ET-2026-001
          </div>
        </div>
      </div>

      <div className="dashboard-grid">
        <div className="panel">
          <div className="panel-head">
            <div>
              <h2>Üretim ve maliyet görünümü</h2>
              <p>Narin Sucuk · örnek gösterim verileri</p>
            </div>
            <select className="tiny-select" defaultValue="7">
              <option value="7">Son 7 gün</option>
              <option value="30">Son 30 gün</option>
            </select>
          </div>
          <div className="chart-wrap">
            <div className="chart-summary">
              <div>
                <strong>1.248 <span style={{ fontSize: 12, color: "var(--muted)", fontWeight: 500 }}>paket</span></strong>
                <span>Toplam üretim</span>
              </div>
              <div>
                <strong>€ {ekp.toFixed(2)}</strong>
                <span>Ort. EKP</span>
              </div>
              <div>
                <strong style={{ color: "var(--green-4)" }}>94,2%</strong>
                <span>Verim</span>
              </div>
            </div>
            <svg className="chart" viewBox="0 0 600 155" preserveAspectRatio="none" aria-label="Üretim grafiği">
              <defs>
                <linearGradient id="fill" x1="0" x2="0" y1="0" y2="1">
                  <stop offset="0%" stopColor="#8cb9a0" stopOpacity="0.28" />
                  <stop offset="100%" stopColor="#8cb9a0" stopOpacity="0" />
                </linearGradient>
              </defs>
              <g stroke="#edf0eb" strokeWidth="1">
                <path d="M0 20H600M0 58H600M0 96H600M0 134H600" />
              </g>
              <path
                d="M0 111 C35 98 45 113 78 83 S130 93 164 70 S215 91 248 62 S300 69 332 45 S386 72 420 50 S468 59 500 33 S552 46 600 18 L600 145 L0 145Z"
                fill="url(#fill)"
              />
              <path
                d="M0 111 C35 98 45 113 78 83 S130 93 164 70 S215 91 248 62 S300 69 332 45 S386 72 420 50 S468 59 500 33 S552 46 600 18"
                fill="none"
                stroke="#2c7058"
                strokeWidth="3"
                strokeLinecap="round"
              />
            </svg>
            <div className="chart-labels">
              <span>23 Eki</span>
              <span>24 Eki</span>
              <span>25 Eki</span>
              <span>26 Eki</span>
              <span>27 Eki</span>
              <span>28 Eki</span>
              <span>29 Eki</span>
            </div>
          </div>
        </div>

        <div className="panel">
          <div className="panel-head">
            <div>
              <h2>Dikkat gerektirenler</h2>
              <p>Önceliğe göre sıralandı</p>
            </div>
            <span className="status warn">2 bildirim</span>
          </div>
          <div className="alert-list">
            <div className="alert">
              <div className="alert-symbol red">⚠</div>
              <div>
                <b>Kalite uyarısı · ET-2026-001</b>
                <p>Bu lotla ilişkili 85 paket SUC-2026-045. Geri çağırma incelemesi başlatın.</p>
                <a href="#recall">Geri çağırmayı aç →</a>
              </div>
            </div>
            <div className="alert">
              <div className="alert-symbol amber">↗</div>
              <div>
                <b>Üretim verimi beklenenin altında</b>
                <p>URT-2026-046 için gerçekleşen miktar tahminden %6,2 düşük.</p>
                <a href="#ai">Tahminleri incele →</a>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="panel" style={{ marginTop: 16 }}>
        <div className="panel-head">
          <div>
            <h2>Son işlemler</h2>
            <p>Üretim, stok ve sevkiyat hareketleri</p>
          </div>
          <button className="btn btn-secondary" type="button">Tüm hareketler ↗</button>
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>İşlem / Referans</th>
                <th>Tür</th>
                <th>Miktar</th>
                <th>Durum</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>
                  <span className="id">URT-2026-045</span>
                  <br />
                  <span className="muted">Narin Sucuk · SUC-2026-045</span>
                </td>
                <td>Üretim kabulü</td>
                <td>85 paket</td>
                <td><span className="status ok">Tamamlandı</span></td>
              </tr>
              <tr>
                <td>
                  <span className="id">SEV-2026-118</span>
                  <br />
                  <span className="muted">Market A</span>
                </td>
                <td>Sevkiyat</td>
                <td>30 paket</td>
                <td><span className="status ok">Sevk edildi</span></td>
              </tr>
              <tr>
                <td>
                  <span className="id">URT-2026-046</span>
                  <br />
                  <span className="muted">Narin Sucuk</span>
                </td>
                <td>Dış üretim</td>
                <td>100 kg hammadde</td>
                <td><span className="status progress">İşlemde</span></td>
              </tr>
              <tr>
                <td>
                  <span className="id">LOT-2026-231</span>
                  <br />
                  <span className="muted">Kuzu but</span>
                </td>
                <td>Stok girişi</td>
                <td>8,4 kg</td>
                <td><span className="status warn">Kontrol bekliyor</span></td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>
  );
}
