# 2026-11-01 — Geri çağırma alıcı teyitlerinin kalıcılığı

## Sorun
Web'de "Teyit alındı" yalnızca tarayıcıdaki geçici bir `Set` idi. Sayfa
yenilenince veya aynı lot için yeni bir geri çağırma açılınca teyitler
kayboluyor, alıcılar tekrar "Teyit bekliyor" görünüyordu. Aynı lot için
yinelenen recall kayıtları da oluşabiliyordu.

## Çözüm
- `recall_acknowledgements` tablosu eklendi (migration
  `0004_recall_acknowledgements`, geri alma dosyası `down/` altında).
- Teyitler **kaynak lot** bazında saklanır; aynı lotun tüm recall kayıtlarında
  ortaktır.
- `POST /recalls/{id}/acknowledge` endpoint'i eklendi.
- `GET /recalls/{id}` ve `POST /recalls` yanıtlarına `confirmed_markets` eklendi.
- Aynı lot için tekrar geri çağırma başlatılırsa yeni kayıt açılmaz; mevcut
  kayıt ve teyitleri döner.
- Web Recall sayfası teyitleri API'den okur/yazar; "Teyit edildi" durumu
  yenilemede korunur.
- Kullanıcı kararı gereği **satış blokajı kaldırılmaz**. Tüm teyitler
  tamamlandığında sayfa "Tüm alıcılar teyit etti · Satış blokajı aktif kalıyor"
  mesajını gösterir.

## Doğrulama
- Go birim + entegrasyon testleri: teyit ekleme, idempotent tekrar, aynı lotun
  farklı recall kayıtlarında ortak teyit, bilinmeyen market reddi.
- Canlı: `SUC-2026-045` için 3/3 teyit kaydedildi; aynı lotun eski recall
  kaydında da 3/3 göründü; lot blokajı ve `blocked_lot_codes` değişmedi.
- Web build + 8 Node testi başarılı.
