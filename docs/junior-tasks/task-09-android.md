# JUNIOR (Meryem) GELİŞTİRİCİ TALİMATI + GÖREV 9

## ROL (önce bunu oku)
Sen **Meryem takımının JUNIOR** yazılım geliştiricisisin. Meryem takımının senior'ı
sana görev verdi; ilayda takımının senior'ı da işini inceler ve merge eder. Sen kodu
YAZARSIN; **commit/push'u sen yapmazsın** — değişiklikleri `feat/task-09-android`
branch'ine koyup **pull request** açarsın.

Kuralların (kesin uy):
1. **SADECE görevde isteneni yap.** Fazladan ekran/kütüphane EKLEME.
2. Verilen isim, endpoint ve veri alanlarını **BİREBİR** kullan.
3. İsimlendirme İngilizce, açıklayıcı. Clean code, LLM-okunur.
4. Günlük notu: yaptıklarını **kendi klasörün olan** `ALPHA-DUO-Vault/Meryem/YYYY-MM-DD.md`
   dosyasına yaz (`ilayda/` klasörüne dokunma). Gerçeği yaz; derlemediğin şeyi "derlendi" diye yazma.
5. Teslimat: `feat/task-09-android` branch'i + PR. Dosyaların TAM içeriği.

## BAĞLAM
Monorepo. Yeni `mobile/` dizini: Kotlin + Jetpack Compose + Room (SQLite).
Backend Go API `http://localhost:8080` üzerinde:
- `GET /sync/pull` → `{"blocked_lot_codes": ["SUC-2026-045"], "products": [{id, code, name, unit, list_price, ...}]}`
- `POST /sales` → `{"code","market_id","sold_at","status","client_id","lines":[{"lot_id","qty","unit_price","total_price"}]}` → 201 veya bloklu lot için **409**
- `POST /sync/push` → `{"sales":[...]}` → 200 veya 409

## GÖREV: Kotlin Android satış + offline SQLite + senkronizasyon

### 1) `mobile/` Gradle iskeleti (Kotlin DSL)
- `settings.gradle.kts`, `build.gradle.kts`, `app/build.gradle.kts` (Compose + Room + Ktor/OkHttp).
- Paket: `com.avatar.app`. minSdk 26, targetSdk 34.

### 2) `data/AppDatabase.kt` + `data/entities/` (Room)
- `LocalProduct(id, code, name, unit, list_price)`
- `LocalLot(id, code, item_type, qty, unit, blocked: Boolean)` (blok bilgisi çekme sonrası işaretlenir)
- `LocalSale(id, code, market_id, sold_at, status, synced: Boolean)`
- `LocalSaleLine(id, sale_id, lot_id, qty, unit_price, total_price)`
- `Outbox(id, entity, entity_id, operation, payload, created_at)`

### 3) `data/SyncRepository.kt`
- `pull()`: `GET /sync/pull` → ürünleri ve bloke lot kodlarını Room'a yaz (lotlar `blocked` bayrağıyla).
- `push()`: `Outbox` içindeki satışları `POST /sync/push` ile gönder; 409 dönerse kullanıcıya
  "bloklu lot satılamaz" hatası ilet, satışı `synced=false` bırak.
- LWW: çekme sırasında mevcut kayıtları `updated_at`/son çekme ile güncelle (basit: pull ile ez).

### 4) `ui/screens/`
- `ProductList.kt` — ürünleri listele; bloklu lotları devre dışı göster.
- `Sale.kt` — ürün + lot + miktar seç → yerel satış oluştur (Outbox'a yaz, `synced=false`).
- `History.kt` — yerel satışları listele (senkron durumuyla).

### 5) `MainActivity.kt` — Compose ile 3 ekran arası basit geçiş + `SyncRepository` çağrıları.

## KABUL (senior kontrol edecek)
- Gradle projesi derlenebilir yapıda (Kotlin/Compose/Room tutarlı).
- `GET /sync/pull` / `POST /sync/push` / `POST /sales` endpoint'lerini doğru JSON ile kullanır.
- Bloklu lotların satışı engellenir (UI + 409 yakalama).
- İngilizce isimlendirme, clean code, ekstra gereksiz bağımlılık yok.
