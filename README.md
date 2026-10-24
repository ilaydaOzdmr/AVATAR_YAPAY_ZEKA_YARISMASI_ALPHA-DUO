# AVATAR Yapay Zeka Yarışması — ALPHA-DUO

Gıda işletmesi için uçtan uca izlenebilirlik + maliyet + yapay zekâ sistemi.
Hammaddeden nihai satışa kadar maliyet (EKP), stok, parçalama, lot (parti)
izlenebilirliği ve geri çağırma; iki zorunlu KI tahmini.

## Bileşenler
- `backend/` — Go REST API (doğruluk kaynağı)
- `web/` — React yönetim + raporlama
- `mobile/` — Kotlin Android satış (offline SQLite + sync)
- `ml/` — Python KI servisi (FastAPI + scikit-learn)
- `infra/` — Docker Compose + SonarQube

## Hızlı başlangıç (tamamlanacak)
```bash
cp .env.example .env
docker compose -f infra/docker-compose.yml up -d
```

## Dokümantasyon
- Tasarım: `docs/superpowers/specs/2026-10-24-avatar-ai-competition-design.md`
- Plan: `docs/superpowers/plans/2026-10-24-avatar-ai-competition.md`
- Veri modeli: `docs/data-model.md`
- API: `docs/api.md`
