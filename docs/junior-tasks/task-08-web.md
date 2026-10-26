# JUNIOR (Meryem) GELİŞTİRİCİ TALİMATI + GÖREV 8

## ROL (önce bunu oku)
Sen **Meryem takımının JUNIOR** yazılım geliştiricisisin. Meryem takımının senior'ı
sana görev verdi; ilayda takımının senior'ı da işini inceler, test eder ve merge eder.
Sen kodu YAZARSIN; **commit/push'u sen yapmazsın** — değişiklikleri `feat/task-08-web`
branch'ine koyup **pull request** açarsın.

Kuralların (kesin uy):
1. **SADECE görevde isteneni yap.** Fazladan sayfa/özellik/kütüphane EKLEME.
2. Verilen isim, endpoint ve veri alanlarını **BİREBİR** kullan.
3. İsimlendirme İngilizce, açıklayıcı. Clean code, LLM-okunur.
4. Günlük notu: yaptıklarını **kendi klasörün olan** `ALPHA-DUO-Vault/Meryem/YYYY-MM-DD.md`
   dosyasına yaz (`ilayda/` klasörüne dokunma). Gerçeği yaz; çalıştırmadığın şeyi "çalışıyor" diye yazma.
5. Teslimat: `feat/task-08-web` branch'i + PR. Dosyaların TAM içeriği.

## BAĞLAM
Monorepo. Yeni `web/` dizini: Vite + React + TypeScript. Backend Go API şu anda
`http://localhost:8080` üzerinde çalışıyor ve şu endpoint'leri sunuyor:
- `GET /materials`, `GET /products`
- `GET /production-orders/{id}` → `CostSummary` (input_cost, service_cost, total_cost,
  finished_qty, ekp, vkp, gross_profit, margin)
- `GET /butchery-batches/{id}` → `ButcheryResult` (batch, items, weight_balance_ok,
  fire_weight, cost_mismatch, expected_cost)
- `GET /trace/forward?lot=CODE` → `[AffectedLot]`; `GET /trace/backward?lot=CODE` → `[RawMaterialSource]`
- `POST /recalls` (`{"source_lot_code","reason"}`) → `RecallResult`; `GET /recalls/{id}` → `RecallResult`
  (`{event, source_lot_code, affected_lots, markets:[{market_code,market_name,contact,qty}], message}`)

## GÖREV: React web (yönetim + raporlama ekranları)

### 1) `web/` iskeleti (Vite + React + TS)
- `npm create vite@latest web -- --template react-ts` (veya elle eşdeğeri).
- `web/src/api.ts` — `fetch` sarmalayıcı: `const API_BASE = import.meta.env.VITE_API_BASE ?? "http://localhost:8080"`;
  `get<T>(path): Promise<T>` ve `post<T>(path, body): Promise<T>` yardımcıları.

### 2) Sayfalar (`web/src/pages/`)
- `Dashboard.tsx` — basit başlık + nav (sayfalar arası geçiş).
- `Materials.tsx` — `GET /materials` listesini tablo olarak göster.
- `Products.tsx` — `GET /products` listesini tablo olarak göster.
- `ProductionOrders.tsx` — lot kodu gir → `GET /production-orders/{id}` → maliyet tablosu
  (EKP, VKP, brüt kâr, marj).
- `Butchery.tsx` — lot kodu gir → `GET /butchery-batches/{id}` → parça/maliyet tablosu;
  `weight_balance_ok=false` ise "Fire: X kg" uyarısı, `cost_mismatch=true` ise
  "Maliyet uyuşmazlığı (beklenen: X)" uyarısı göster.
- `Traceability.tsx` — iki sekme: İleri (`/trace/forward`) ve Geri (`/trace/backward`);
  lot kodu gir → sonuç listesi (ileri: lot + marketler; geri: hammadde + tedarikçi).
- `Recall.tsx` — kaynak lot kodu + sebep gir → `POST /recalls` → sonuç: etkilenen lotlar,
  market tablosu (market, miktar, iletişim) + üretilen geri çağırma mesajı.

### 3) `web/src/App.tsx` — basit routing (React Router yerine state ile sekme geçişi yeter).

## KABUL (senior kontrol edecek)
- `npm install` + `npm run build` hatasız.
- Sayfalar yukarıdaki endpoint'leri kullanıp veriyi tablo olarak gösteriyor.
- İngilizce isimlendirme, clean code, ekstra kütüphane yok (fetch + state yeterli).
