# JUNIOR (Meryem) GELİŞTİRİCİ TALİMATI + GÖREV 10

## ROL (önce bunu oku)
Sen **Meryem takımının JUNIOR** yazılım geliştiricisisin. Meryem takımının senior'ı
sana görev verdi; ilayda takımının senior'ı da işini inceler ve merge eder. Sen kodu
YAZARSIN; **commit/push'u sen yapmazsın** — değişiklikleri `feat/task-10-ml-web`
branch'ine koyup **pull request** açarsın.

Kuralların (kesin uy):
1. **SADECE görevde isteneni yap.** Fazladan sayfa/kütüphane EKLEME.
2. Verilen isim, endpoint ve veri alanlarını **BİREBİR** kullan.
3. İsimlendirme İngilizce, açıklayıcı. Clean code, LLM-okunur.
4. Günlük notu: yaptıklarını **kendi klasörün olan** `ALPHA-DUO-Vault/Meryem/YYYY-MM-DD.md`
   dosyasına yaz (`ilayda/` klasörüne dokunma). Gerçeği yaz; çalıştırmadığın şeyi "çalışıyor" diye yazma.
5. **Çok sayıda ayrı commit at** (aşağıdaki "COMMIT'LER" bölümüne birebir uy).

## BAĞLAM
Monorepo. `web/` (Vite + React + TS) zaten var. ML servisi `http://localhost:8000`
üzerinde çalışıyor ve şu endpoint'leri sunuyor:
- `POST /predict/yield` — gövde `{"raw_material_kg":100,"service_cost":467.5}` →
  `{"predicted_qty":84.682,"predicted_unit_cost":16.975}`
- `POST /predict/butchery` — gövde `{"animal_type":"kuzu","input_weight":25}` →
  `{"parts":[{"product_code":"PRD-101","weight":4.509}, ...]}`
- `GET /report` — `{"yield":{"model_mae","model_rmse","baseline_mae","baseline_rmse"},
  "butchery":{...}, "note":"..."}`

## GÖREV: Web'e ML "Predictions" sayfası ekle

### 1) `web/src/api.ts` — ML yardımcıları ekle
```ts
const ML_BASE: string = import.meta.env.VITE_ML_BASE ?? "http://localhost:8000";
export function mlGet<T>(path: string): Promise<T> { /* get<T> ile aynı, ML_BASE kullan */ }
export function mlPost<T>(path: string, body: unknown): Promise<T> { /* post<T> ile aynı, ML_BASE kullan */ }
```
> Mevcut `request` fonksiyonunu ML için de kullan; sadece base URL farklı olsun.

### 2) `web/src/pages/Predictions.tsx` — yeni sayfa
- **Yield tahmini:** `raw_material_kg` + `service_cost` girişi → `mlPost("/predict/yield", ...)`
  → `predicted_qty` ve `predicted_unit_cost` göster.
- **Parçalama tahmini:** `animal_type` + `input_weight` girişi → `mlPost("/predict/butchery", ...)`
  → `parts` listesini tablo olarak göster (product_code, weight).
- **Rapor:** `mlGet("/report")` → iki tablo (yield + butchery): model MAE/RMSE vs baseline MAE/RMSE;
  `note` alanını da göster.

### 3) `web/src/App.tsx` — nav'a "Predictions" ekle
`pages` listesine `{ key: "predictions", label: "Predictions" }` ekle ve render et.

## COMMIT'LER (ayrı ayrı, sırasıyla)
1. `feat(web): add ml api client helpers` — `api.ts`
2. `feat(web): add predictions page` — `Predictions.tsx`
3. `feat(web): add predictions nav entry` — `App.tsx`
4. `docs: log task 10 work` — `ALPHA-DUO-Vault/Meryem/YYYY-MM-DD.md`

## KABUL (senior kontrol edecek)
- `npm run build` hatasız.
- Tahmin formları ve rapor, ML endpoint'lerini doğru JSON ile çağırıyor.
- İngilizce isimlendirme, clean code, ekstra kütüphane yok.
