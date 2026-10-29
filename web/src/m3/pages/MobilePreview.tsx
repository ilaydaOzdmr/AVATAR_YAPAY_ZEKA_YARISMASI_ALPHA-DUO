export default function MobilePreview() {
  return (
    <section>
      <div className="pagehead">
        <div>
          <div className="eyebrow">ANDROID · KOTLIN</div>
          <h1>Mobil Operasyon</h1>
          <p>Saha ve satış ekipleri için tek elle kullanılabilen, hızlı ve güvenli akış.</p>
        </div>
        <span className="status ok">Mobil tasarım önizlemesi</span>
      </div>

      <div className="mobile-preview">
        <div className="mobile-status">
          <span>09:41</span>
          <span>●●● ▰</span>
        </div>
        <div className="mobile-head">
          <small>AVATAR OPERASYON</small>
          <h2>Merhaba, Ayşe 👋</h2>
          <p style={{ fontSize: 11, color: "var(--muted)", margin: "4px 0 0" }}>Bugünkü operasyonların hazır.</p>
        </div>
        <div className="mobile-body">
          <div className="mobile-card" style={{ background: "var(--amber-bg)", borderColor: "var(--amber-line)" }}>
            <div style={{ display: "flex", gap: 10 }}>
              <span style={{ fontSize: 22 }}>⚠</span>
              <div>
                <h3>Kalite uyarısı</h3>
                <p>ET-2026-001 lotu bloke edildi. İlişkili ürünlerin satışı engellendi.</p>
              </div>
            </div>
          </div>
          <div className="mobile-quick">
            <button className="quick" type="button">
              <span>▣</span>
              <b>Hızlı satış</b>
              <small>Lot seçerek satış yap</small>
            </button>
            <button className="quick" type="button">
              <span>⌗</span>
              <b>Lot okut</b>
              <small>Barkod / QR tara</small>
            </button>
            <button className="quick" type="button">
              <span>▤</span>
              <b>Stok sorgula</b>
              <small>Anlık stok durumu</small>
            </button>
            <button className="quick" type="button">
              <span>↗</span>
              <b>Sevkiyatlar</b>
              <small>Hazır ve gönderilenler</small>
            </button>
          </div>
          <div className="mobile-card">
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <h3>Son satış</h3>
              <span className="status ok">Tamamlandı</span>
            </div>
            <p style={{ marginTop: 8 }}>SUC-2026-052 · 4 paket · € 96,00</p>
            <div style={{ height: 1, background: "var(--line)", margin: "12px 0" }} />
            <div style={{ display: "flex", justifyContent: "space-between" }}>
              <span className="muted" style={{ fontSize: 10 }}>Lot kontrolü</span>
              <b style={{ fontSize: 11, color: "var(--green-4)" }}>✓ Satışa uygun</b>
            </div>
          </div>
        </div>
      </div>

      <p className="muted" style={{ marginTop: 16, textAlign: "center" }}>
        Bu ekran, Kotlin/Compose mobil uygulamasının tasarım önizlemesidir. Gerçek işlevler
        Android uygulamasında çalışır.
      </p>
    </section>
  );
}
