# imavara-server değişiklik geçmişi

Kayıt kapsamı ve sürüm çıkarma adımları için [AGENTS.md](AGENTS.md) içindeki değişiklik ve sürüm kurallarını izle.

## Yayımlanmamış

- Agent sonuç açıklamalarını Markdown olarak düzenler. Araç çağrıları arasındaki metinler ayrı paragrafta gösterilir.

- RabbitMQ yönetim paneli Pi üzerinde localhost:15672 adresinde etkinleştirildi.

- API ve MCP katalog aramasında isteğe bağlı `top_k` parametresi 1–10 sonuç sınırı seçer. Varsayılan 5’tir.

- Aramada benzerlik skoru ve güçlü eşleşme bulunmadığında tek yakın aday desteği eklendi.

- Üretim ve varyasyonlar teslimle birlikte ortak katalog değerlendirmesine girer. Eski özel durumdaki üretimler aynı işlem hattına alınır.

- Yerel katalog 16 shard ve replika faktörü 1 ile tek Qdrant node’unda çalışır. Compose yeni `qdrant_data` volume’ünü kullanır. Eski katalog için geçiş notu README’dedir.

- Repo çalışma rehberi ve değişiklik geçmişi eklendi.
- Geliştirme rehberine servis teslim komutları ve yeniden teslimde metriklerin nasıl doğrulanacağı eklendi.
- Yerel Go API ve PostgreSQL kurulumu, anonim tarayıcı oturumu, başlangıç bakiyesi, tarife ve kredi hareketleri eklendi.
- Sohbet geçmişi, ses transkripsiyonu, iç MCP araçları ve LangChain agent için servis girişleri eklendi.
- Katalog araması, dosya erişimi, edinim, üretim, katkı ve kuyruk işleme kodu ile servis panoları eklendi. Canlı model kabul kontrolleri sürüyor.
- WebP orijinaller, gerçekleşen/istenen oran, özel referans yükleme, Türkçe transkripsiyon dil ipucu ve hata kuyruğu tekrar komutu eklendi.
