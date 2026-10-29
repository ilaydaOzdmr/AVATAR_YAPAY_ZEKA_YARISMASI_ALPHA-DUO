import { useEffect, useState } from "react";
import { get } from "../../api";

interface Lot {
  id: number;
  item_type: string;
  qty: number;
}

interface Product {
  id: number;
}

export default function Reports() {
  const [materialCount, setMaterialCount] = useState(0);
  const [productCount, setProductCount] = useState(0);
  const [lotCount, setLotCount] = useState(0);

  useEffect(() => {
    get<Lot[]>("/lots").then((lots) => {
      setLotCount(lots.length);
      setMaterialCount(lots.filter((l) => l.item_type === "material").length);
    }).catch(() => {});
    get<Product[]>("/products").then((p) => setProductCount(p.length)).catch(() => {});
  }, []);

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
        <button className="btn btn-primary" type="button">⇩ Rapor oluştur</button>
      </div>

      <div className="metrics">
        <div className="metric">
          <div className="metric-top">Hammadde lotu <span className="metric-icon">▤</span></div>
          <strong>{materialCount}</strong>
          <div className="metric-foot">Mevcut kayıt</div>
        </div>
        <div className="metric">
          <div className="metric-top">Mamul / ürün kartı <span className="metric-icon">▣</span></div>
          <strong>{productCount}</strong>
          <div className="metric-foot">Kayıtlı ürün</div>
        </div>
        <div className="metric">
          <div className="metric-top">Toplam lot <span className="metric-icon">↔</span></div>
          <strong>{lotCount}</strong>
          <div className="metric-foot">Giriş, çıkış ve transfer</div>
        </div>
        <div className="metric">
          <div className="metric-top">Kalite olayı <span className="metric-icon" style={{ background: "var(--amber-bg)", color: "var(--amber)" }}>⚠</span></div>
          <strong>01</strong>
          <div className="metric-foot warning">Takipte</div>
        </div>
      </div>

      <div className="panel">
        <div className="panel-head">
          <h2>Rapor kataloğu</h2>
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Rapor</th>
                <th>Kapsam</th>
                <th>Format</th>
                <th></th>
              </tr>
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
