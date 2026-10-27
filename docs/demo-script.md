# Demo Scripti — AVATAR ALPHA-DUO

Video/demo için adım adım senaryo. Tek komutla tüm sistem ayağa kalkar.

## 0) Başlat
```bash
docker compose -f infra/docker-compose.yml up -d --build
```
- Web: http://localhost:5173 · API: http://localhost:8080 · ML: http://localhost:8000

## 1) Katalog (Materials / Products)
- "Materials" ve "Products" sekmeleri → önceden tanımlı fiyat kataloğunu göster.
- Vurgu: tüm fiyatlar önceden tanımlı (PDF ile birebir).

## 2) Maliyet (Production Orders)
- "Production Orders" → ID `1` gir → **EKP 18.16 ₺**, VKP 24.90 ₺, brüt kâr 6.74 ₺, marj %27.07.
- Açıklama: 100 kg et (9.70) + 2 kg baharat (24.00) + 5 kg bağırsak (11.60) + 85×5.50 hizmet = 1543.50 ₺ → /85 = 18.16 ₺.

## 3) İzlenebilirlik (Traceability)
- İleri: `ET-2026-001` → **SUC-2026-045** + Market A(30) B(25) C(30).
- Geri: `SUC-2026-045` → ET / BAH / BAR + tedarikçiler.

## 4) Geri çağırma (Recall) — demo odağı
- Kaynak lot `ET-2026-001`, sebep "kalite sorunu" → **Recall oluştur**.
- Sonuç: etkilenen lot (SUC-2026-045), market tablosu (miktar + iletişim) + üretilen geri çağırma mesajı.
- Not: lotlar bloklandı; satış engellendi (aşağıda kanıtı).

## 5) Satış engeli (Android / API)
- Android'de (veya `POST /sales` ile) bloklu `SUC-2026-045` lotunu satmaya çalış → **409 (red)**.

## 6) Parçalama (Butchery)
- "Butchery" → ID `1` gir → 25 kg kuzu, parçalar 22 kg → **Fire: 3 kg** uyarısı;
  maliyet dağıtımı toplamı tam (örn. 130.68 + 104.55 + 52.27 = 287.50).

## 7) KI (Predictions)
- Yield tahmini (100 kg, 467.5 hizmet) → ~85 paket, ~18 ₺ birim maliyet.
- Parçalama tahmini (kuzu, 25 kg) → 7 parça ağırlıkları.
- Model raporu: model MAE/RMSE vs baseline (ortalama) — model daha iyi.

## Kapanış vurguları
- Doğru maliyet + eksiksiz lot ilişkileri + güvenilir stok + geri çağırma + ölçülebilir KI katkısı.
