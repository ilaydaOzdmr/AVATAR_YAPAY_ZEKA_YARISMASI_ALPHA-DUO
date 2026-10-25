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
- [x] Tasarım onayı (2026-10-24)
- [x] Plan onayı
- [ ] Uygulama (Task 1..11)

## İlerleme
- [x] **Task 1** — monorepo iskeleti + Docker (MySQL, SonarQube) → `chore: scaffold...`
- [x] **Task 2** — MySQL şema + seed → `feat: add mysql schema and scenario seed data`
  - Doğrulandı (Docker/MySQL): 17 tablo; EKP 18,16 ₺ (girdi 1.076,00 + hizmet 467,50); marketler 30/25/30; UTF-8 OK.
- [x] **Task 3** — Go backend iskeleti + CRUD + stok → `feat: add go backend skeleton...`
  - Doğrulandı: `go build` + `go vet` + `go test` geçti; Docker'da `/health`, `/materials`, `/lots/{code}`, hareketler çalışıyor.
- [ ] **Task 4** — Maliyet motoru (EKP/VKP/brüt kâr/marj)
