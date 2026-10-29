import { useCallback, useEffect, useState } from "react";
import { get } from "../../api";

interface Lot {
  id: number;
  item_type: string;
}

interface Anomaly {
  ref: string;
}

export default function Reports() {
  const [productionCount, setProductionCount] = useState(0);
  const [saleCount, setSaleCount] = useState(0);
  const [lotCount, setLotCount] = useState(0);
  const [materialCount, setMaterialCount] = useState(0);
  const [productCount, setProductCount] = useState(0);
  const [movementCount, setMovementCount] = useState(0);
  const [anomalyCount, setAnomalyCount] = useState(0);

  const load = useCallback(() => {
    get<unknown[]>("/production-orders").then((r) => setProductionCount(r.length)).catch(() => {});
    get<unknown[]>("/sales").then((r) => setSaleCount(r.length)).catch(() => {});
    get<unknown[]>("/products").then((r) => setProductCount(r.length)).catch(() => {});
    get<unknown[]>("/anomalies").then((r) => setAnomalyCount(r.length)).catch(() => {});
    get<Lot[]>("/lots").then((lots) => {
      setLotCount(lots.length);
      setMaterialCount(lots.filter((l) => l.item_type === "material").length);
      Promise.all(
        lots.map((lot) =>
          get<unknown[]>(`/lots/${lot.id}/movements`)
            .then((m) => m.length)
            .catch(() => 0),
        ),
      ).then((counts) => setMovementCount(counts.reduce((a, b) => a + b, 0)));
    }).catch(() => {});
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const catalog = [
    { name: "Gerçek üretim maliyeti", scope: "Hammadde + hizmet + EKP", format: "PDF / CSV" },
    { name: "Lot hareket geçmişi", scope: "Giriş, transfer, üretim, satış", format: "CSV" },
    { name: "Geri çağırma alıcı listesi", scope: "Market ve miktar bazında", format: "CSV" },
    { name: "AI tahmin hata raporu", scope: "Gerçek / tahmin / MAE / RMSE", format: "PDF" },
  ];

  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">İŞLETME ANALİTİĞİ</div>
          <h1>Raporlar</h1>
          <p>Alış, üretim, parçalama, lot hareketleri, satış ve geri çağırma kayıtları.</p>
        </div>
        <button className="btn btn-primary" type="button" onClick={load}>↻ Yenile</button>
      </div>

      <div className="metrics">
        <div className="metric">
          <div className="metric-top">Üretim emri <span className="metric-icon">⚙</span></div>
          <strong>{productionCount}</strong>
          <div className="metric-foot">Kayıtlı emir</div>
        </div>
        <div className="metric">
          <div className="metric-top">Stok hareketi <span className="metric-icon">↔</span></div>
          <strong>{movementCount}</strong>
          <div className="metric-foot">Giriş, çıkış ve transfer</div>
        </div>
        <div className="metric">
          <div className="metric-top">Satış <span className="metric-icon">↗</span></div>
          <strong>{saleCount}</strong>
          <div className="metric-foot">Lot bazında kayıtlı</div>
        </div>
        <div className="metric">
          <div className="metric-top">Kalite olayı <span className="metric-icon" style={{ background: "var(--amber-bg)", color: "var(--amber)" }}>⚠</span></div>
          <strong>{anomalyCount}</strong>
          <div className="metric-foot">Takipte</div>
        </div>
      </div>

      <div className="metrics">
        <div className="metric">
          <div className="metric-top">Toplam lot <span className="metric-icon">▤</span></div>
          <strong>{lotCount}</strong>
          <div className="metric-foot">{materialCount} hammadde</div>
        </div>
        <div className="metric">
          <div className="metric-top">Ürün kartı <span className="metric-icon">▣</span></div>
          <strong>{productCount}</strong>
          <div className="metric-foot">Kayıtlı ürün</div>
        </div>
      </div>

      <div className="panel">
        <div className="panel-head">
          <h2>Rapor kataloğu</h2>
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr><th>Rapor</th><th>Kapsam</th><th>Format</th><th></th></tr>
            </thead>
            <tbody>
              {catalog.map((report) => (
                <tr key={report.name}>
                  <td><b>{report.name}</b></td>
                  <td>{report.scope}</td>
                  <td>{report.format}</td>
                  <td><button className="btn btn-secondary" type="button">Oluştur</button></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </section>
  );
}
