---
type: gorev-dagilimi
---

# Görev Dağılımı

İki takım çalışıyor: **ilayda** (senior + junior) ve **Meryem** (senior + junior).
ilayda'nın senior'ı her iki takımın işini inceler ve commit'leri atar.

| Görev | Konu | Sahip | Yapan | Durum |
|-------|------|-------|-------|-------|
| 1 | İskelet + Docker + SonarQube | ilayda | senior | ✅ |
| 2 | MySQL şema + seed | ilayda | senior | ✅ |
| 3 | Go backend iskeleti + CRUD + stok | ilayda | senior | ✅ |
| 4 | Maliyet motoru (EKP/VKP/marj) | Meryem | junior | ✅ (senior onayladı) |
| 5 | Lot izlenebilirliği (ileri/geri) | ilayda | junior | ✅ (senior onayladı) |
| 6 | Parçalama + maliyet dağıtımı | Meryem | junior | ✅ (senior onayladı) |
| 7 | ML servisi (2 model) | ilayda | junior | ✅ (senior onayladı) |
| 8 | React web (yönetim + rapor) | Meryem | junior | ✅ (senior onayladı) |
| 9 | Kotlin Android + senkronizasyon | Meryem | junior | ✅ (senior onayladı) |
| 10 | ML tahmin + rapor ekranı (web) | Meryem | junior | ✅ (senior onayladı) |
| 11 | Kapanış (test/SonarQube/demo) | ilayda | senior | ✅ (Docker orchestration + demo script + logo + README) |

## Backend ekleri (senior — ilayda)
Görev 5'in kalan kısmı senior tarafından öne alındı:
- Geri çağırma (`POST /recalls`, `GET /recalls/{id}`) + lot bloklama → ✅
- Satış (`POST /sales`) + bloklu lot engeli (409) → ✅
- Senkronizasyon (`POST /sync/push`, `GET /sync/pull`) → ✅

## Kurallar
- Meryem takımı günlüğünü `Meryem/`, ilayda takımı `ilayda/` klasörüne yazar.
- Commit'leri ilayda senior atar; junior'lar commit/push yapmaz.
- Junior'lar kendi adlarına (Meryem/ilayda) PR açar; senior inceler, doğrular, merge eder.

---

# PDF tam uyum — eksikleri giderme planı (2026-10-30 değerlendirmesi)

İki takıma bölündü. **Önce ilayda (backend), sonra Meryem (AI/arayüz)** — çünkü
Meryem'in form/rapor işleri ilayda'nın stok/lot davranışına bağlı.

## ilayda takımı — Backend veri bütünlüğü (P0, kritik)

| # | Görev | Detay |
|---|-------|-------|
| B1 | Stok defterini gerçek yap | Üretim/parçalama/satış/iade/transfer/rework/fire işlemlerinde `stock_movements` yaz; `lots.qty`'yi aynı transaction'da güncelle |
| B2 | Satışta maliyet alanını düzelt | `unit_cost`'a satış fiyatı değil **lot maliyeti** yaz |
| B3 | Runtime lot bağları | Üretimde `lot_links(parent=input, child=output, qty)`; parçalamada giriş lotu → her parça lotu `lot_links` |
| B4 | Butchery'de ürün lotu üret | `product_lot_id` istemciden beklenmesin; backend lot oluştursun + stok hareketiyle stoğa girsin |
| B5 | Recall'i sağlamlaştır | Transaction yap; **net** miktar (iade düş) + ikinci kademe dağıtımı sonuca ekle; `lots.status` bloke et; kod çakışmasını önle |
| B6 | Entegrasyon testleri | Stok, lot-link, recall, üretim/parçalama HTTP entegrasyon testleri |

## Meryem takımı — AI + arayüz + rapor + mobil + doküman

| # | Görev | Detay |
|---|-------|-------|
| M1 | AI tahmin–gerçek (P0) | Seed'deki 9 üretim/9 parçalama kaydını eğitim/değerlendirmede kullan; AI A'da model `predicted_qty`/`predicted_unit_cost` vs gerçek `finished_qty`/EKP; AI B'de parça ağırlıkları vs `butchery_items` fark tablosu |
| M2 | Web üretim formu | Hammadde lotları + miktarları + mamul çıktısı alanları |
| M3 | Web parçalama formu | Gerçek tarih, toplam maliyet, adet ve **fire satırı** |
| M4 | Raporlar | Alış, parçalama ve satış raporları (CSV) |
| M5 | Tutarsızlık gösterimi | 287,50 € / 351,00 € ve 22 kg örneğini batch + açıklama olarak göster |
| M6 | Mobil iyileştirmeler | Offline `IOException` yakala ("yerel kaydedildi"); market seçimi; lot–ürün uyumu; yerel stok hareketi |
| M7 | Dokümantasyon | `docs/api.md`, `data-model.md`, `demo-script.md`, README güncelle; LWW'yi ya uygula ya README'den kaldır |

## Operasyonel not (demo öncesi)
Eski image/volume'dan kurtulmak için: `docker compose -f infra/docker-compose.yml down -v && docker compose -f infra/docker-compose.yml up -d --build`
