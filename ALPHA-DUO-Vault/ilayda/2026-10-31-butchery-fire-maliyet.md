# 2026-10-31 — Kuzu parçalama fire ve maliyet tutarlılığı

## Yapılanlar

### Sabit uyarı kaldırıldı
- `web/src/m3/pages/Butchery.tsx` üstündeki "PDF örnek tutarsızlığı" banner'ı
  ve `KUZ-2026-DEMO` açıklaması tamamen kaldırıldı.

### Otomatik analiz
- Seçili kaydın detayında, fire tespitinin altına iki blok eklendi:
  - **Fire ve Tutarlılık Analizi**: giriş ağırlığı, parça toplamı, hesaplanan
    fire ve veriden üretilen açıklama.
  - **Maliyet Kontrolü**: beklenen maliyet, kayıtlı maliyet, maliyet farkı ve
    veriden üretilen açıklama.
- Tutarlı sonuç yeşil/nötr, tutarsız sonuç amber uyarı olarak gösterilir.
- Sayılar `tr-TR` biçiminde (287,50 € / 3 kg) gösterilir.

### Backend
- Yeni `backend/internal/butchery/check.go`: ağırlık ve maliyet kontrollerinin
  tek kaynağı.
- `ButcheryResult` yeni alanları: `part_weight`, `weight_check_status`,
  `cost_check_status`, `recorded_cost`, `cost_difference`.
- `ButcheryBatch` yeni alanları: `expected_cost`, `cost_difference`,
  `cost_check_status`, `weight_check_status`.
- Create sırasında hesaplanıp kaydedilir; GET ve liste yanıtlarında döner.

### Migration
- `backend/migrations/0003_butchery_consistency.up.sql`
  - dört kolonu ekler ve mevcut kayıtları backfill eder.
- `backend/migrations/down/0003_butchery_consistency.down.sql`
  - geri alınabilir; MySQL init dizini alt klasörleri çalıştırmadığı için
    `down/` altında tutuldu.
- `db.EnsureSchema` eski Docker hacimleri için aynı kolonları ekler ve
  backfill'i her açılışta idempotent çalıştırır.

## Hesap mantığı
- `part_weight` = `is_waste = FALSE` parçaların toplamı.
- `fire_weight` = `input_weight - part_weight` (3 ondalık).
- `expected_cost` = `input_weight × unit_cost` (2 ondalık).
- `cost_difference` = `total_cost - expected_cost` (2 ondalık).
- Ağırlık kontrolü: fire negatifse (parçalar girişi aşıyorsa) `mismatch`.
  Fire tek başına hata değildir.
- Maliyet kontrolü: yuvarlanmış kayıtlı maliyet beklenenden farklıysa
  `mismatch`. İki kontrol bağımsızdır.

## Doğrulama
- `KUZ-2026-DEMO`: fire 3 kg, parça 22 kg, beklenen 287,50 €, kayıtlı 351,00 €,
  fark 63,50 €, `cost_check_status = mismatch`, `weight_check_status = ok`.
- `PRC-2026-002` (normal): fire 3,08 kg, beklenen = kayıtlı = 253,00 €,
  `ok/ok` — yanlış uyarı yok.
- Servis testleri:
  - demo değerleri,
  - kontrollerin bağımsızlığı,
  - kayıtlı fire satırının ağırlığı dengelemesi,
  - tutarlı kayıtta uyarı olmaması,
  - 2/3 ondalık toleransları.
- Repo entegrasyon testleri: create + get + list kalıcılığı ve normal kayıt.
- `db.EnsureSchema` entegrasyon testi: dört kolon + demo backfill.
- `go test ./...`, `go vet ./...`, web `npm run build` temiz.

## Ek: azami fire sınırı (2026-11-01)
- Ağırlık kontrolü artık negatif fire yanında **girişin %20'sini aşan fireyi**
  de `mismatch` sayıyor. 50 kg girişte azami 10 kg fire kabul edilir.
- 50 kg giriş / 27 kg parça / 23 kg fire kaydı artık "Uyarı" gösteriyor;
  maliyet kontrolü bağımsız kalıyor.
- `EnsureSchema` backfill'i ve `0003_butchery_consistency` migration'ı aynı
  %20 kuralını kullanıyor.
