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
| 11 | Kapanış (test/SonarQube/demo) | ilayda | senior | ⏳ |

## Backend ekleri (senior — ilayda)
Görev 5'in kalan kısmı senior tarafından öne alındı:
- Geri çağırma (`POST /recalls`, `GET /recalls/{id}`) + lot bloklama → ✅
- Satış (`POST /sales`) + bloklu lot engeli (409) → ✅
- Senkronizasyon (`POST /sync/push`, `GET /sync/pull`) → ✅

## Kurallar
- Meryem takımı günlüğünü `Meryem/`, ilayda takımı `ilayda/` klasörüne yazar.
- Commit'leri ilayda senior atar; junior'lar commit/push yapmaz.
- Junior'lar kendi adlarına (Meryem/ilayda) PR açar; senior inceler, doğrular, merge eder.
