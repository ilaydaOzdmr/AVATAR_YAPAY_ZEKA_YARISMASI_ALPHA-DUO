# 2026-10-31 — Üretim emri “Tamamla” akışı

## Kök neden
- Web formu üretim emrini her zaman `status: "draft"` ile oluşturuyordu.
- Backend’de durum güncelleme endpoint’i yoktu; bu yüzden emir sonsuza kadar
  “İşlemde” görünüyordu.
- Ayrıca eski `CreateProductionOrder` draft emirde bile stok hareketi
  yazabiliyordu; “işlemde” emir stoğu tüketmiş olabilirdi.

## Yapılanlar
- `CreateProductionOrder`:
  - draft → yalnızca emir + girdi/çıktı satırları; **stok hareketi yok**,
  - completed → stok, lot bağı ve SKT aynı transaction’da.
- `CompleteProductionOrder` eklendi:
  - girdi/çıktı satırlarını okur, stok ve `lot_links` işler,
  - `status='completed'`, `finished_at` yazar,
  - tekrar çağrılırsa stok ikinci kez düşmez (idempotent),
  - eski hatalı draft’larda hareket varsa tekrar uygulamaz.
- `POST /production-orders/{id}/complete` endpoint’i eklendi.
- Web üretim tablosuna **Tamamla** butonu eklendi; tamamlanınca durum
  “Tamamlandı” olur ve maliyet özeti yenilenir.

## Ek: ret ve beklenen miktar
- Web üretim formunda **Beklenen mamul** alanı eklendi; `expected_qty` artık
  gönderiliyor (önceden hiç gönderilmiyordu, bu yüzden 0 görünüyordu).
- `POST /production-orders/{id}/reject` ve **Reddet** butonu eklendi.
  - Ret yalnızca tamamlanmamış ve stok hareketi uygulanmamış emirlerde çalışır.
  - Durum `rejected`; gerekçe `notes` alanına yazılır.
  - Stok/lot/SKT değişmez.
- Açık emir sayısı artık `completed` ve `rejected` dışındakileri sayıyor.
- Test: `TestRejectProductionOrderKeepsStockAndExpectedQty`; canlıda
  `expected=42`, `status=rejected`, ret sonrası giriş stoğu 10 kaldı.

## Doğrulama
- `TestCompleteProductionOrderAppliesStockOnce`:
  - draft: girdi 10, çıktı 0,
  - complete: girdi 0, çıktı 8, lot bağı 1,
  - ikinci complete: çıktı hâlâ 8.
- Canlı API:
  - draft `input=10 output=0`,
  - complete `status=completed input=0 output=8 expiry=2026-11-15`,
  - ikinci complete `output=8`.
- Go unit + integration + vet temiz; web build başarılı.
