# SonarQube Taraması

## Sunucu ve kapsam
```powershell
docker compose -f infra/docker-compose.yml up -d sonarqube
```

Sunucu: [localhost:9000](http://localhost:9000).
Proje: [AVATAR ALPHA-DUO](http://localhost:9000/dashboard?id=avatar-alpha-duo).
Hesabın `My Account → Security` bölümünden tarama token'ı oluşturun ve
`SONAR_TOKEN` ortam değişkenine atayın. Token'ı Git'e veya komut dosyalarına yazmayın.

`infra/sonar-project.properties` backend, web, mobile, ml ve infra kaynaklarını
kapsar. Test dosyaları ayrı sınıflandırılır. Bağımlılıklar, sanal ortamlar,
derleme çıktıları, veri setleri, model dosyaları ve yerel raporlar hariçtir.

Kurulu SonarQube 9.9.8 sunucusunun `/api/languages/list` yanıtında Go desteği
mevcuttur ve Go sensor'ı çalışır. Önceki "Community Edition Go içermez" notu
bu kurulum için geçerli değildir. `go vet` gibi dil araçları ayrıca kullanılabilir.

## Docker ile tarama

Repo kökünde PowerShell ile çalıştırın. Aşağıdaki resmi scanner image'ının
özeti, yerelde doğrulanan SonarScanner CLI 8.1.0 image'ını sabitler.

```powershell
New-Item -ItemType Directory -Force ml/reports/local/sonar | Out-Null
$sonarScannerImage = "sonarsource/sonar-scanner-cli@sha256:47956b669bf486bdad919e1cbd0533223515aa0b6d1d31763e5a3754942df6d7"
docker run --rm --user 0 `
  --mount "type=volume,source=avatar-sonar-scanner-work,target=/sonar-work" `
  --entrypoint chown $sonarScannerImage 1000:1000 /sonar-work
docker run --rm `
  --env SONAR_TOKEN `
  --env SONAR_HOST_URL=http://host.docker.internal:9000 `
  --mount "type=bind,source=$($PWD.Path),target=/usr/src,readonly" `
  --mount "type=bind,source=$($PWD.Path)\ml\reports\local\sonar,target=/sonar-output" `
  --mount "type=volume,source=avatar-sonar-scanner-cache,target=/opt/sonar-scanner/.sonar/cache" `
  --mount "type=volume,source=avatar-sonar-scanner-work,target=/sonar-work" `
  $sonarScannerImage `
  -Dproject.settings=infra/sonar-project.properties `
  -Dsonar.working.directory=/sonar-work `
  -Dsonar.scanner.metadataFilePath=/sonar-output/report-task.txt `
  -Dsonar.qualitygate.wait=true
```

`project.settings` zorunludur: yapılandırma dosyası repo kökünde değildir.
Kaynaklar salt okunur bağlanır. Geçici analiz ve önbellek dosyaları Docker
volume'larında tutulur; Windows bağlamasındaki binlerce geçici dosyanın yazılması
önlenir. İşlem bilgisi ve dışa aktarılan sonuçlar Git dışındaki
`ml/reports/local/sonar/` dizinindedir.

Eski TypeScript analizörü uygulamanın `moduleResolution=bundler` ve
`allowImportingTsExtensions` seçeneklerini tanımadığı için yalnız analizde
`web/tsconfig.sonar.json` kullanılır. `web/tsconfig.json` ve Vite derleme ayarları
değiştirilmez. Analiz çıktısında TypeScript dosyalarının atlanmadığını doğrulayın.

Yerel CLI kullanılıyorsa:

```powershell
sonar-scanner -Dproject.settings=infra/sonar-project.properties -Dsonar.host.url=http://localhost:9000 -Dsonar.working.directory=ml/reports/local/sonar/work
```

## Kapsam raporlarını yeniden üretme

Go, Python, Node, Android SDK/JDK ve yerel MySQL/ML servisleri gerekir.
Go test sarmalayıcısı yalnız `avatar_sonar_test_*` adlı yeni bir geçici veritabanı
oluşturur ve temizler; işletme veritabanını kullanmaz. Web testleri seed kayıtları
olan yerel backend'i kullanır ve işletme verisine yazmaz.

Repo kökünden:

```powershell
python -m pip install -r ml/requirements-dev.txt
python -m coverage run --rcfile=ml/.coveragerc -m pytest ml/tests ml/test_model.py -q
python -m coverage xml --rcfile=ml/.coveragerc
python backend/tests/run_coverage.py
npm.cmd --prefix web ci
npm.cmd --prefix web run test:coverage:all
```

Chrome farklı bir konumdaysa `CHROME_PATH` ortam değişkenine executable yolunu
verin. Tarayıcı V8 çıktısı QA derlemesinin source map'leriyle gerçek kaynaklara
eşlenir ve Node ölçümüyle LCOV olarak birleştirilir. Normal web derlemesi source
map yayınlamaz; QA çıktıları yalnız `coverage/` altındadır.

Android için `mobile/` içinde JDK ile:

```powershell
.\gradlew.bat :app:testDebugUnitTest :app:jacocoTestReport :app:assembleDebug
```

JaCoCo 0.8.16, Kotlin/Compose için resmi sentetik bytecode filtrelerini içerir;
uygulama dosyaları test kapsamından çıkarılmaz. [JaCoCo sürüm notları](https://www.jacoco.org/jacoco/trunk/doc/changes.html).
Test JVM'inde İngilizce locale kullanılır: Türkçe Windows'ta native test
kütüphanelerinin `Windows` adını yanlış küçültmesini önler. Uygulamanın Türkçe
tarih/para biçimi testlerle korunur.

Son olarak repo kökünde:

```powershell
python infra/normalize_coverage.py
```

Bu komut yalnız dosya yollarını eşler; satır/branch çalışma sayaçlarını değiştirmez.
Ardından yukarıdaki scanner komutunu çalıştırın. Genel kapsam ile yeni kod
kapsamı ayrıdır; kalite kapısındaki yeni kod kapsam eşiği %80 olarak korunur.

## Sonucu okuma

Scanner'ın başarılı olması, kalite kapısının geçtiği anlamına gelmez.
`report-task.txt` içindeki işlem tamamlandıktan sonra dashboard ve Quality Gate
sonucunu kontrol edin. İlk taramadaki yeni kod kapsamı ile tüm proje bulguları
farklı olabilir.

Test kapsamı raporları içeri alınmadığında görünen %0 coverage, testlerin
çalışmadığını veya başarısız olduğunu kanıtlamaz. Gerçek kapsam için ilgili
test araçlarının coverage raporları ayrıca üretilip yapılandırılmalıdır.

SonarQube 9.9.8 Python analizörü, projenin Python 3.12 hedefini desteklemediğini
bildirip 3.10 kurallarıyla analiz eder. Bu tarama Python 3.12'ye özgü tam
uyumluluk doğrulaması sayılmaz.
Kurulu web analizörü scanner image'ındaki Node 22 için tavsiye edilmeyen sürüm
uyarısı da verir; bu sınırlılık raporda belirtilir.
