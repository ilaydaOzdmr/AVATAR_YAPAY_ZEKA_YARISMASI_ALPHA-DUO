# SonarQube Taraması

## 1) Sunucuyu başlat
```bash
docker compose -f infra/docker-compose.yml up -d sonarqube
```
- `http://localhost:9000` (varsayılan kullanıcı/şifre: `admin`/`admin`).
- `admin` ile gir → `My Account → Security` altından bir token üret.

## 2) Scanner çalıştır
`sonar-scanner` CLI'yi [sonarsource.com](https://docs.sonarsource.com/sonarqube/latest/analyzing-source-code/scanners/sonarscanner/) adresinden indir ve PATH'e ekle. Sonra repo kökünde:

```bash
sonar-scanner \
  -Dsonar.host.url=http://localhost:9000 \
  -Dsonar.token=<TOKEN> \
  -Dsonar.projectKey=avatar-alpha-duo
```

`infra/sonar-project.properties` proje ayarlarını içerir.

## 3) Dil desteği notu
- **Web (JS/TS)**, **ML (Python)**, **Android (Kotlin)** — SonarQube Community Edition ile taranır.
- **Go** — Community Edition Go analizini **içermez**. Go için kalite kapısı olarak
  `go vet ./...` (geçiyor) ve `golangci-lint run` kullanılır:
  ```bash
  cd backend && go vet ./... && golangci-lint run
  ```
