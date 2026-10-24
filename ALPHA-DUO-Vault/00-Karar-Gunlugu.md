---
type: karar-gunlugu
created: 2026-10-24
---

# Karar Günlüğü — ALPHA-DUO

## Proje özeti
Gıda işletmesi için uçtan uca izlenebilirlik + maliyet + KI sistemi.
Detay: `docs/superpowers/specs/2026-10-24-avatar-ai-competition-design.md`

## Alınan kararlar
- **Yaklaşım:** Omurga önce (dikey dilimler); kayarsak minimal prototipe kırp.
- **Süre:** 5 gün kod + 2 gün anlama/video; teslim 1 Kasım. Takım: 2 kişi.
- **Öncelik:** bitmiş + eksiksiz + canlı çalışır proje (demo: ET-2026-001 → marketler).
- **Teknik yığın:** Go (stdlib) + MySQL 8 + React (Vite+TS) + Kotlin (Compose+Room) + Python (FastAPI+sklearn).
- **Fiyatlar:** tümü önceden tanımlı (seed), PDF ile birebir realist katalog.
- **Maliyet yöntemi:** ağırlıklı ortalama; fire EKP'ye yedirilir.
- **Geri çağırma:** sunucu otoriter (bloke lot satışı 409 red).
- **Senkronizasyon:** outbox + last-write-wins (SQLite↔MySQL).
- **Süreç:** ayrık/sıralı commit (Conventional Commits), branch'ler, clean code + İngilizce isimlendirme, Docker + SonarQube.

## Plan
- Uygulama planı: `docs/superpowers/plans/2026-10-24-avatar-ai-competition.md` (11 görev).

## Durum
- [ ] Tasarım onayı (yapıldı: 2026-10-24)
- [ ] Plan onayı
- [ ] Uygulama (Task 1..11)
