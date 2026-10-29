# Skills

Bu klasör, arayüz geliştirmede kullanılan ajan becerilerini (skills) projeye dahil eder.

## `material-3/`
Kaynak: <https://github.com/hamen/material-3-skill>

Google Material Design 3 (Material You) uygulama rehberi. `SKILL.md` + `references/`
(renk sistemi, bileşen kataloğu, tipografi/shape, layout/responsive, navigasyon,
tema/dinamik renk). Web arayüzü (5174) ve Kotlin/Compose mobil için kullanıldı.

## `impeccable/`
Kaynak: <https://github.com/pbakaus/impeccable>

Ajan için tasarım sözlüğü: `polish`, `bolder`, `critique`, `audit` vb. komutlar;
AI "slop" yakalayan 62 deterministik kural. `SKILL.md` + `reference/` (24 komut
playbook'u) + `agents/` (alt-ajan prompt'ları).

> Not: `impeccable` CLI/tool'u (`npx impeccable install`) bu klasörün dışında,
> ajan harness'ına kurulur. Buradaki dosyalar skill içeriğinin proje içindeki kopyasıdır.

## Kullanım
Beceriler, skill yükleyebilen bir ajan (OpenCode, Claude Code, Codex vb.) tarafından
okunur; arayüz tasarım/iyileştirme işlerinde rehber olarak kullanılır.
