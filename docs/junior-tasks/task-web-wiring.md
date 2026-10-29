# JUNIOR (Meryem) GELİŞTİRİCİ TALİMATI + GÖREV: Web arayüzünü backend'e tam bağla

## ROL (önce bunu oku)
Sen **Meryem takımının JUNIOR** yazılım geliştiricisisin. ilayda takımının senior'ı işini inceler,
test eder ve merge eder. Sen kodu YAZARSIN; **commit/push'u sen yapmazsın** — değişiklikleri
`feat/web-wiring` branch'ine koyup **pull request** açarsın.

Kuralların (kesin uy):
1. **SADECE görevde isteneni yap.** Yeni sayfa/özellik uydurma.
2. Mevcut `web/src/m3/` yapısını ve `api.ts` içindeki `get`/`post` yardımcılarını kullan.
3. Tüm veri **backend'den** gelecek; sabit (hardcoded) örnek sayı yazma.
4. İngilizce olmayan arayüz metinleri Türkçe kalacak (prototip gibi); kod/alan adları İngilizce.
5. **Çok sayıda ayrı commit at** (aşağıdaki "COMMIT'LER"e uy).
6. Günlük: `ALPHA-DUO-Vault/Meryem/YYYY-MM-DD.md` (kendi klasörün). Gerçeği yaz.

## BAĞLAM
- `web/src/m3/` → Food Intelligence arayüzü (Vite multi-page, `m3.html`).
- `web/src/api.ts` → `get<T>(path)` ve `post<T>(path, body)` (base `http://localhost:8080`).
- Backend uçları: `GET/POST /materials`, `GET/POST /products`, `GET/POST /lots`,
  `POST /stock-movements`, `GET/POST /production-orders`, `GET /production-orders/{id}`,
  `GET/POST /butchery-batches`, `GET /butchery-batches/{id}`, `GET /trace/forward?lot=`,
  `GET /trace/backward?lot=`, `POST /recalls`, `GET /recalls/{id}`, `POST /sales`,
  `GET /sync/pull`, `GET /markets`.

## GÖREV: Her ekranı ve her aksiyonu gerçek backend'e bağla

### 1) `Inventory.tsx` (Stok & Hammaddeler)
- `GET /materials` + `GET /lots?item_type=material` ile gerçek listeyi göster (sabit örnek satırları kaldır).
- "＋ Hammadde ekle" → form → `POST /materials`.
- Stok girişi → form → `POST /stock-movements` (movement_type `inbound`).
- Metrikleri (aktif lot sayısı, toplam kg, stok değeri) API verisinden hesapla.

### 2) `Production.tsx` (Üretim Emirleri)
- `GET /production-orders` listesi (gerçek). Metrikleri (açık emir, tamamlanan, ort. EKP) hesapla.
- "＋ Yeni üretim emri" → form → `POST /production-orders` (code, service_unit_cost, service_qty, inputs[], outputs[]).
- Maliyet detayı: seçili emir için `GET /production-orders/{id}` → EKP/VKP/brüt kâr/marj.

### 3) `Butchery.tsx` (Kuzu Parçalama)
- `GET /butchery-batches` listesi (gerçek). Ağırlık dengesi + fire uyarılarını `GET /butchery-batches/{id}` ile göster.
- "＋ Parçalama kaydı" → form → `POST /butchery-batches` (batch + items[]).

### 4) `Sales.tsx` (Satış & Sevkiyat) — çoğu hazır; şunu ekle
- Market seçimi: `GET /markets` → satış formuna market dropdown ekle (şu an sabit `market_id: 1`).
- Sevkiyat listesi: `GET /lots/{code}/movements` veya `GET /stock-movements` ile gerçek sevkiyatları göster.

### 5) `Traceability.tsx` (Lot İzlenebilirliği)
- İleri/geri izlemeyi `GET /trace/forward?lot=` / `GET /trace/backward?lot=` ile gerçek veriyle göster (sabit akışı kaldır).

### 6) `Recall.tsx` (Geri Çağırma)
- `POST /recalls` (kaynak lot + sebep) → sonucu (ürün adı + marketler + mesaj) göster.
- Alıcı teyit butonlarını ve CSV indirmeyi `affected_lots[].markets` verisinden üret (sabit 3 market yazma).

### 7) `Predictions.tsx` (Yapay Zekâ)
- Yield + parçalama tahminini `POST /predict/yield` / `POST /predict/butchery` (ML, `http://localhost:8000`) ile bağla; raporu `GET /report` ile göster. (Mevcut api.ts'te `mlGet`/`mlPost` var.)

### 8) `Reports.tsx` (Raporlar)
- Gerçek sayıları backend'den hesapla: üretim emri sayısı (`/production-orders`), stok hareketi sayısı, satış satırı sayısı (`/sales`), lot sayısı (`/lots`). Sabit "12 / 486 / 1.284" gibi sayıları kaldır.

### 9) `Dashboard.tsx` (Genel Bakış) — çoğu hazır; şunu ekle
- "Son işlemler" ve "Dikkat gerektirenler" bölümünü gerçek veriye bağla (ör. son sevkiyatlar, açık geri çağırmalar).

## COMMIT'LER (ayrı ayrı, sırasıyla)
1. `feat(web): wire inventory and production screens`
2. `feat(web): wire butchery and sales screens`
3. `feat(web): wire traceability and recall screens`
4. `feat(web): wire predictions and reports screens`
5. `feat(web): wire dashboard metrics and recent activity`
6. `docs: log web wiring work`

## KABUL
- `npm run build` hatasız.
- Her ekran backend'den canlı veri çekiyor; create formları gerçekten kaydediyor.
- Sabit örnek sayı yok; metrikler API verisinden hesaplanıyor.
