# Mobil senkronizasyon — SQLite ↔ MySQL

## 1. Temel ilke

**MySQL tek doğruluk kaynağıdır (source of truth).** Android’deki Room/SQLite
çevrimdışı önbellek ve gönderim kuyruğudur (outbox). Stok, lot, satış ve geri
çağırma kararları sunucuda verilir; mobil bunları yerelde tahmin edip resmi kayıt
gibi göstermez.

Bu nedenle “her iki taraf da birbirinin aynısı” bir çift yönlü çoklu-yazıcı
(multi-master) senkronizasyonu **bilinçli olarak kullanılmaz**. Böyle bir model
stokta çift düşüm ve geri çağırma yarışı üretir.

## 2. Ağ bağlantısı

Mobil önce `/health` ile erişilebilir backend’i bulur:

1. Yapılandırılmış adres (varsa)
2. `http://10.0.2.2:8080` — Android emülatöründe host
3. `http://localhost:8080` — fiziksel cihazda `adb reverse` sonrası

Bulunan adres bellekte önbelleğe alınır. Hiçbiri yoksa kullanıcıya denenen
adreslerle birlikte açık hata gösterilir.

Fiziksel cihaz:

```bash
adb reverse tcp:8080 tcp:8080
```

Emülatörde ek ayar gerekmez; `10.0.2.2` otomatik denenir.

## 3. Pull (MySQL → SQLite)

`GET /sync/pull` tek snapshot döndürür:

```json
{
  "server_time": "2026-10-31T...",
  "blocked_lot_codes": ["SUC-2026-045"],
  "products": [...],
  "lots": [...],
  "markets": [...]
}
```

Mobil:

1. Tüm snapshot’ı **önce ayrıştırır**.
2. Room içinde **tek transaction** ile `local_products`, `local_lots`,
   `local_markets` tablolarını değiştirir.
3. Herhangi bir ağ/parse hatasında **hiçbir tabloyu silmez**; eski önbellek
   korunur.
4. `blocked_lot_codes` sunucudan geldiği için bloklu lotlar yerelde de
   satılamaz.

Böylece “önce temizle, sonra yaz” kaynaklı yarım kalmış SQLite durumu oluşmaz.

## 4. Push (SQLite → MySQL)

Çevrimdışı satış:

1. `local_sales` + `local_sale_lines` + `outbox` aynı yerel işlemde yazılır.
2. Lot bakiyesi yerelde iyimser olarak düşülür (kullanıcıya anında geri bildirim).
3. Bağlantı gelince outbox **satış satış** gönderilir (`POST /sales`).

### Idempotency

Satış kodu (`code`) idempotency anahtarıdır. Aynı kod tekrar gönderilirse
sunucu yeni satış oluşturmaz; mevcut satışı döndürür ve stok **ikinci kez
düşülmez**.

### 409 (bloklu lot)

Sunucu bloklu lot nedeniyle satışı reddederse:

- yerel satış `status = rejected`,
- `synced = true`,
- outbox kaydı silinir.

Böylece aynı satış sonsuza kadar tekrar denenmez.

### Ağ hatası

Satış yerelde `synced = false` ve outbox’ta kalır; sonraki `push()` denenir.

## 5. Stok uzlaşması

- Satış başarılıysa sunucu stoğu düşer; sonraki pull sunucu bakiyesini getirir.
- Satış reddedildiyse sonraki pull bloklu lotu ve sunucu bakiyesini getirir;
  yerel iyimser düşüm sunucu değeriyle değiştirilir.
- Stokta “last write wins” uygulanmaz; sunucu değeri kazanır.

## 6. Raporlar

Mobil raporları yerel SQLite’tan değil, **MySQL’den** okur:

`GET /reports/summary`

Dönen altı bölüm:

| Bölüm | Kaynak |
|---|---|
| `purchases` | Hammadde lotları, tedarikçi, miktar, birim maliyet |
| `production` | Üretim emirleri, girdi kg, mamul, gerçek EKP |
| `butchery` | Parçalama partileri, parça/fire ağırlığı, maliyet |
| `lot_movements` | Tüm lot stok hareketleri |
| `sales` | Satışlar, market, satır/miktar/tutar |
| `recalls` | Geri çağırma olayları ve bloklu lotlar |

Mobil “Raporlar” sekmesi bu altı bölümü gösterir; bağlantı yoksa açık
“çevrimdışı” mesajı verir ve sahte rapor üretmez.

## 7. Çakışma kuralları

| Durum | Kazanan |
|---|---|
| Ürün/lot/market referansı | Sunucu |
| Stok miktarı | Sunucu |
| Bloklu lot | Sunucu (satış reddedilir) |
| Aynı satış kodu tekrar | Sunucudaki mevcut kayıt |
| Rapor içeriği | Sunucu |

## 8. Doğrulama kontrol listesi

- [ ] Emülatörde `10.0.2.2`, fiziksel cihazda `adb reverse` ile `/health` 200.
- [ ] `GET /sync/pull` ürün + lot + market + bloklu lotları birlikte döndürür.
- [ ] Ağ kesilince pull SQLite’ı silmez.
- [ ] Aynı satış iki kez gönderilince stok bir kez düşer.
- [ ] Bloklu lot satışı 409 alır ve outbox’tan temizlenir.
- [ ] `/reports/summary` altı bölümü döndürür.
- [ ] Mobil Raporlar sekmesi çevrimdışıyken açık hata gösterir.
