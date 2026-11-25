# Imavara server

Yerel MVP için Docker Engine ve Compose kullanılır. Aynı makinede iki Qdrant düğümü, PostgreSQL, Garage, RabbitMQ, API, MCP, agent, analiz/indeks worker’ları, Prometheus ve Grafana çalışır.

1. `.env.example` dosyasını `.env` olarak kopyala. Gerçek model çağrıları için `OPENROUTER_API_KEY` gir. Ağ ve depolama için yerel demo değerleri Compose dosyasında bulunur.
2. Repo kökünde `docker compose up -d --build` çalıştır. Başlatma adımları veri şemasını, Garage alanlarını ve Qdrant koleksiyonunu hazırlar. `docker compose ps -a` ile durumu kontrol et. Başarıyla tamamlanan garage-init ve catalog-init süreçlerinin çıkış kodu sıfırdır.
3. API sağlık kontrolü için `curl http://127.0.0.1:18080/healthz` kullan. Arayüzü web reposundaki yönergeyle aç. Grafana yerel adresi `http://127.0.0.1:13000`, başlangıç girişi Grafana’nın yerel admin hesabıdır.
4. Servis loglarını `docker compose logs -f api image-extract-worker index-worker` ile izle. `docker compose down` kalıcı diskleri korur.

Go paket kontrolleri için `go test ./...` çalıştır. Kredi eşzamanlılığı testi için boş, geçici PostgreSQL bağlantısını `IMAVARA_TEST_DATABASE_URL` ile belirt. Test bu veritabanına migration uygular ve test kayıtları yazar.

Sürekli hatalı işler üç denemeden sonra ilgili hata kuyruğunda kalır. Nedeni düzelttikten sonra `docker compose exec image-extract-worker replay -queue assets.extract.v1 -limit 1` ile tek bir işi yeniden dene. İndeks işleri için `assets.index.v1` kullan. Komut yalnız broker onayından sonra hata kuyruğundaki kopyayı kapatır.

Qdrant yedeği için indeks yazımlarının durduğu bir anda `docker compose exec index-worker snapshot` çalıştır. Tek node’un snapshot dosyası ve manifest Garage’ın özel snapshots bucket’ına yazılır. Çıkan kimliği `docker compose exec index-worker snapshot -id KIMLIK -restore-to YENI_KOLEKSIYON` ile kullan. Komut mevcut koleksiyonu korur, kurtarılan point sayısını ve bilinen bir referansla aramayı kontrol eder.

Eski iki node’lu kurulumdan geçerken yeni `qdrant_data` volume’ü boş bir 16 shard koleksiyonuyla başlar. Mevcut katalog ve PostgreSQL kayıtlarını birlikte kullanmaya devam etmek için geçişten önce indeks yazımlarını durdur, eski koleksiyonun point kimliklerini, iki vektörünü ve payload’larını dışa aktar. Yeni koleksiyona aynı kimliklerle yükle, point sayısını ve bilinen aramaları doğrula. Eski volume’leri geçiş doğrulanana kadar sakla. Önceki iki parçalı snapshot eski topolojiyle geri yüklenir.

[Ürün ve mimari planı](docs/IMAVARA-MVP.md) · [Servis görevleri](docs/MVP-SERVIS-GOREVLERI.md) · [Teslim kriterleri](docs/MVP-TESLIM-PLANI.md)
