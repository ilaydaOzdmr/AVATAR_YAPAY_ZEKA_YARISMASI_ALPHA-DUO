# SonarQube güvenlik incelemesi

## Giderilen noktalar

- Backend ve ML çalışma süreçleri ayrı, yetkisiz kullanıcılarla çalışır.
- Web image'ları yetkisiz Nginx kullanıcısıyla 8080 portunda çalışır; dış portlar
  yine 5173/5174'tür.
- `COPY . .` yerine derleme için gereken dosya/dizinler açıkça listelenmiştir.
  Yerel token, yedek, veri indirmeleri ve raporlar image'a kopyalanmaz.
- Android'in ana manifestinde veri yedekleme ve açık HTTP kapatılmıştır.
  Debug manifestinde de yedekleme kapalıdır.

## İncelenerek güvenli kabul edilen geliştirme ayarları

| Nokta | İnceleme gerekçesi |
| --- | --- |
| Debug manifestindeki `usesCleartextTraffic=true` | Yalnız debug varyantı, operatörün yerel Go backend'ine ve `adb reverse` bağlantısına HTTP ile erişir. Release manifesti bu dosyayı içermez ve HTTP'ye izin vermez. Üretim bağlantısı HTTPS olmalıdır. |
| MainActivity içindeki eski emulator IP sabiti | Eski otomatik sunucu tercihini ayırt etmek için karşılaştırılan Android emulator host alias'ıdır. Burada ağ isteği gönderilmez; kullanıcı verisi sabit bir dış sunucuya yönlendirilmez. |

Bu iki SonarQube hotspot'u gerekçeleriyle `REVIEWED / SAFE` olarak işaretlenir.
Diğer bulgular kaynak kod düzeltmesiyle giderilir; kurallar kapatılmaz ve
uygulama kaynakları tarama dışına alınmaz.
