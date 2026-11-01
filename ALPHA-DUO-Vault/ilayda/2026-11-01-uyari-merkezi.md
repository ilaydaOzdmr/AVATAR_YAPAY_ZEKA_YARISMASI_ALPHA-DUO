# 2026-11-01 — Uyarı merkezi, okundu durumu ve stok kuralı

## Yapılanlar
- Stok anomalisinin mesajı negatif bakiye yerine **"Stok az olduğu için satış
  engellendi"** olarak değiştirildi; kural lot bakiyesi 1 birim veya altına
  düştüğünde çalışır (tarihsel hareket toplamı kullanılmaz).
- Lot bakiyesinden fazla satış backend'de `ErrInsufficientStock` ile
  reddediliyor; böylece yeni negatif stok oluşmuyor.
- `anomaly_reads` tablosu (migration `0005_anomaly_reads`) ve
  `POST /anomalies/read` eklendi. Uyarı anahtarı `<type>:<ref>`.
- Web'e **Uyarılar** sayfası eklendi:
  - tüm uyarılar listelenir; okundu olanlar soluk görünür,
  - tek tek veya tümü okundu işaretlenebilir,
  - okunmamış sayısı sidebar ve bildirim zilinde rozet olarak görünür.
- Zil ikonu ve Recall sayfasındaki uyarı banner'ı Uyarılar sayfasını açar.
- Dashboard yalnızca okunmamış uyarıları gösterir; "Tüm uyarılar" ile listeye
  geçilir.
- Recall sayfasındaki banner okunmuş ve aktif recall kapsamındaki lotları
  göstermez.
- Mobilde 409 yanıtı `insufficient stock` içeriyorsa satış `stock_low` olarak
  işaretlenir ve History'de **"Stok yetersiz: satış engellendi"** görünür;
  bloklu lottan ayrıdır. Yerel stok iadesi iki durumda da yapılır.
- Mobil tahmin ekranındaki model sürümü/`synthetic_only` uyarıları kaldırıldı;
  yalnızca tahmin çıktısına ilişkin kısa açıklama gösteriliyor.
- Mobil stok listesinde bloklu lotlar kırmızı kart olarak gösteriliyor.
  Üstteki "Bloklu lot" sayacına tıklanınca liste yalnızca bloklu lotlara
  filtreleniyor.
- Mobil satış formu, lot bakiyesinden fazla miktar girildiğinde butonu
  kapatıp **"Stok yetersiz: mevcut X"** uyarısı gösteriyor.

## Doğrulama
- Go birim + entegrasyon testleri: stoktan fazla satış reddi, lot miktarının
  değişmemesi, anomali okundu anahtarları.
- Canlı API: `stock:SUC-2026-045` okundu → unread 12'den 11'e düştü.
- Web: Uyarılar sayfası 12 kayıt / 11 okunmamış; okundu işaretleme rozeti
  düşürdü; stok mesajı doğru.
- Mobil: `InsufficientStockSaleTest`, `BlockedSaleSyncTest`, `SaleStatusTest`
  geçti; APK derlendi.
