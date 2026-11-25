# Imavara server çalışma rehberi

Bu monorepo yerel MVP’nin API, iç MCP, LangChain agent ve iki worker sürecini taşır. İşe ilgili servisin normal girişinden başla, çağırdığı ortak kodu ve çıktıyı tüketen servisi izle. Değişikliği bu akış üzerinden doğrula.

## Yerleşim ve sorumluluk

- Go giriş noktalarını `cmd/{api,mcp,image-extract-worker,index-worker}` altında tut. İş kuralları ilgili `internal` alanında, Garage, OpenRouter, kuyruk ve ortak loglama bağlantıları `internal/platform` altında yaşar.
- TypeScript agent ve prompt’lar `agent/`, sürümlü mesaj şemaları `contracts/`, veri geçişleri `migrations/`, yerel işletim dosyaları `deploy/` ve kökteki Compose dosyasında yer alır. Mevcut paketi genişletmeden önce çağıranlarını incele.
- API tarayıcı bağlamını, erişimi ve krediyi yönetir. Agent iç MCP üzerinden API işlemlerini çağırır. Analiz worker’ı doğrulanmış JSON üretir. İndeks worker’ı iki vektörü aynı point’e yazar ve yayımı tamamlar. Kuyruk mesajlarında nesne anahtarı ve sürümü taşı.

## Uyumluluk ve doğrulama

HTTP alan adları, durum kodları, WebSocket olayları, MCP araçları ve RabbitMQ mesajları bağımsız tüketicilerin sözleşmesidir. Ek alanı eski girdilerle çalışacak şekilde tanımla. Zorunlu değişiklikte önce her iki sürümü okuyabilen tüketiciyi çıkar, ardından üreticiyi geçir. Eski kuyruk mesajıyla ve önceki istemci isteğiyle sonucu dene. Migration’ı mevcut satırlarla doğrula. Embedding boyutu veya modeli değişirse yeni koleksiyonu doldurup doğruladıktan sonra arama alias’ını geçir.

Kredi, üretim, yayım veya ACK sırası değiştiğinde isteği tekrar gönder ve yazımdan sonraki kesintiyi dene. Dosya, bakiye, point ve ödülün tek işlemle tutarlı kaldığını doğrula.

## Gözlemlenebilirlik

Değişen işlemde istek veya iş kimliğinin servisler arasında taşındığını kontrol et. Süre, hata, sağlayıcı maliyeti veya kuyruk davranışı değiştiğinde ilgili Prometheus metriğini ve servisin Grafana panelini birlikte dene. Metrik adını, birimini veya etiketini değiştirmeden önce mevcut sorguları incele. İstek ve asset kimliklerini loglarda tut, metrik etiketi olarak kullanma. Sentry etkinse yeni hata yoluna servis ve sürüm bağlamı ekle. Yerel Compose boş DSN ile çalışmalı. Ham ses, özel görsel, oturum çerezi ve sağlayıcı anahtarını telemetriye yazma.

## Değişiklik ve sürüm kaydı

Teslimden önce repo kökünde `go test ./...` ve `docker compose config --quiet` çalıştır. Değişen servisi Compose’da yeniden oluşturup gerçek tüketicisinden çağır. Ölçüm değiştiyse `deploy/prometheus.yml` içindeki hedefi ve `deploy/grafana/dashboards/` altındaki ilgili panel sorgusunu kontrol et. Örneğin bir indeks işi yeniden teslim edildiğinde deneme sayısı artabilir, ancak aynı görsel için ikinci yayım veya ikinci katkı ödülü oluşmamalı.

Kullanıcı davranışı, entegrasyon sözleşmesi, veri geçişi veya işletim gereksinimi değiştiğinde aynı commit ya da değişiklik setinde [CHANGELOG.md](CHANGELOG.md) içindeki “Yayımlanmamış” bölümüne kısa bir kayıt ekle. Etkilenen kullanıcıyı veya servisi, görünen farkı ve gerekiyorsa geçiş adımını yaz. İki repoyu etkileyen işte her repo kendi etkisini kaydeder. Önemli bir sürüm yayımlandığında ilgili kayıtları gerçek sürüm ve tarih altında topla. Henüz planlanan özelliği teslim edilmiş gibi kaydetme.
