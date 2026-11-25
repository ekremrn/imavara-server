# Imavara MVP teslim planı ve kabul kriterleri

**Hedef.** Temiz bir yerel kurulumda kullanıcı, Türkçe yazarak veya konuşarak uygun bir katalog görselini bulabilir, edinebilir, üretebilir veya değiştirebilir. İzin verdiği ve kabul edilen yeni bir görsel katalogda aranabilir hale gelir. Bakiye, görünürlük, sohbet geçmişi ve dosya teslimi bütün yolculuk boyunca tutarlı kalır.

Bu belge uygulama başladığında kullanılacak iş listesidir. Kutular bekleyen işleri gösterir. Ürün davranışının ve model ayarlarının kaynağı [yerel MVP planıdır](IMAVARA-MVP.md).

Servis sahipleri, bağımlılıklar ve adım adım uygulanacak işler [MVP servis görevlerinde](MVP-SERVIS-GOREVLERI.md) yer alır.

## Uygulama sırası

| Sıra | Yapılacak iş | Aşamanın çıktısı |
|---|---|---|
| 1 | İki repoyu, yapılandırma örneğini, Compose ağını, kalıcı diskleri, migration ve başlatma komutlarını hazırlamak. Yedi model rotasını küçük gerçek girdilerle erkenden sınamak. | Çalışan yerel ortam, model başına sonuç ve uyumsuzluk kaydı. Özellikle referansla üretim, iki embedding boyutu, ses biçimi ve Glimmer JSON’u görülür. |
| 2 | API’de tarayıcı oturumu, asset kimliği, dosya erişimi, bakiye, kredi defteri ve tekrarlanan işlem korumasını kurmak. | İki ayrı tarayıcıyla oturum, özel dosya, edinim ve indirme doğrulanır. |
| 3 | İlk yüklemeden önce Qdrant koleksiyonunu ve payload indekslerini kurmak. Dış katkı, kopya kontrolü, Luna, RabbitMQ ve iki worker’ı bağlamak. Kullanım izni kayıtlı başlangıç koleksiyonunu aynı hattan geçirmek. | Bir gerçek dosya yüklemeden yayıma kadar izlenebilir. Kopya, ret ve yeniden teslim sınanır. |
| 4 | Metin/görsel aramayı, kesin filtreleri, edinimi, popülerlik sorgularını ve Keşfet’i tamamlamak. | Keşfet üzerinden uygun görsel bulunur, edinilir ve orijinali indirilir. |
| 5 | Açık üretim ve varyasyon isteklerini Muse Image, Garage ve kredi defterine bağlamak. Özel kitaplık, paylaşım ve geri çekmeyi tamamlamak. | Aynı istek içinde dosya teslim edilir. Seçili referans ve kredi hareketi izlenebilir. |
| 6 | İç MCP araçlarını, LangChain agent’ını, WebSocket’i ve React sohbetini bağlamak. Mikrofonla giriş, eski konuşmalar ve referans kartlarını tamamlamak. | Yazılı ve sesli mesaj aynı ürün işlemlerini kullanır. |
| 7 | Servis panolarını tamamlamak. Hata ve tekrar senaryolarını, arama ölçümünü ve temiz kurulum demosunu yürütmek. | Aşağıdaki her madde için gerçek sonuç ve tekrarlanabilir kanıt bulunur. |

## Yüzde yüz geçmesi gereken kabul kontrolleri

Her kutu bir teslim koşuludur. Kontrolü gerçekleştiren kişi kullanılan komutu veya kullanıcı adımını, beklenen ve gerçekleşen sonucu, iki repo sürümünü ve varsa istek/asset kimliğini kaydeder. Sağlayıcı hataları kontrollü olarak üretilebilir. Canlı model davranışını gösteren kabul örnekleri gerçek OpenRouter çağrısı kullanır.

### Yerel ortam ve servisler

- [ ] **K1 · Temiz kurulum.** İki repo ve belgelenmiş önkoşullarla, boş veri disklerinden React ve Compose başlatılır. Go API, iç MCP, TypeScript agent, iki Go worker, PostgreSQL, Qdrant, RabbitMQ, Garage, Prometheus ve Grafana hazır hale gelir. İkinci başlatma mevcut veriyi ve migration durumunu korur. Ürün tarayıcıdan web/API üzerinden çalışır. MCP ile özel depolama ve veri servisleri iç ağdadır. OpenRouter anahtarı sunucuda kalır. Yerel kabul için Sentry anahtarı gerekmez.
- [ ] **K2 · Model sözleşmeleri.** Ana plandaki yedi tam model kimliğiyle en az bir başarılı gerçek işlem kaydedilir. Ember araç seçer ve metin akıtır, Transcribe Türkçe sesi yazıya çevirir, Muse Image hem sıfırdan hem referansla çıktı verir, Luna katalog kararı üretir, Glimmer sürümlü şemaya uyar, Qwen ve Gemini istenen uzunlukta sonlu vektör verir. Sunucu model profilleri uygulanır. Desteklenmeyen sağlayıcı özelliği varsa etkilenen madde açık kalır.

### Kullanıcının gördüğü ürün

- [ ] **K3 · Sohbet ve araçlar.** Minimal açılışta “imavara.”, tek giriş, mikrofon ve kararlaştırılan Meta Muse ifadesi görünür. İlk mesaj sohbeti açar. Belirsiz ihtiyaçta agent kısa bir soru sorar. Açık aramada iç MCP üzerinden gerçek API sonucu ve varsayılan olarak en fazla beş kart gelir. Açık adet isteğinde `top_k` uygulanır, üst sınır on karttır. Metin ve gerçek araç olayları WebSocket’te görünür. Boş arama üretim çağrısı ve kredi hareketi oluşturmaz. “İkincisine benzer” son gösterilen ikinci kartın kimliğiyle ücretsiz aramadır.
- [ ] **K4 · Sesli mesaj.** Açılışta ve açık sohbette “Sesle anlat” kaydı başlatır. “Gönder” tek dokunuşta bitirir. Beklerken durum görünür. Örnek Türkçe kayıttaki oran, yön ve seçilen kart anlamı korunur. Metin, varsa taslak ve referansla tek mesaj olarak otomatik gider ve tek agent turu başlar. Çift dokunma ve yeniden bağlantıyla sonuç aynıdır. İptal, izin reddi, boş transkript ve hata taslağı korur. Ham ses uygulamanın kalıcı depolarına yazılmaz.
- [ ] **K5 · Keşfet ve geçmiş.** Kullanıcı Keşfet’te agent turu olmadan filtreli metin veya referans araması yapar. Kartta önizleme, oran ve boyut görünür. Yeni sohbet açma, eski sohbeti kart sırasıyla açma, silme ve kitaplığa dönme tarayıcı yenilendikten sonra da çalışır. Sohbeti silmek edinilmiş görseli kitaplıktan kaldırmaz.
- [ ] **K6 · Üretim ve varyasyon.** Açık istekle üretilen tek orijinal önce Garage’a yazılır ve aynı istek içinde teslim edilir. “İkinci görseldeki fincanı kaldır” örneğinde sağlayıcıya giden referans ikinci karttır. Çıktıda istenen değişiklik ve korunması istenen kadraj örnek üzerinde incelenir. Gerçek dosya ölçüsü gösterilir. Üretim otomatik katalog değerlendirmesine girer. İndeksleme tamamlanınca ikinci tarayıcıdan aramayla bulunur.

### Kredi, erişim ve katalog

- [ ] **K7 · Bakiye.** Yeni tarayıcı bir kez 100 kredi alır. Sohbet, ses, arama ve önizleme 0, ilk edinim −1, tekrar indirme 0, teslim edilen üretim veya varyasyon −3, ilk yayımlanan benzersiz dış katkı +1’dir. Yetersiz bakiyede sağlayıcıya ücretli üretim gitmez. Çift tıklama, aynı işlem anahtarıyla HTTP tekrarı, son üç kredi için iki eşzamanlı istek ve kesin üretim hatası ayrı ayrı sınanır. Defter, rezervasyonlar ve kullanılabilir bakiye uyuşur. Belirsiz sağlayıcı yanıtı ikinci üretimi tetiklemez.
- [ ] **K8 · Erişim ve geri çekme.** İki bağımsız tarayıcıyla özel sohbet, görsel, orijinal dosya ve bakiye erişimi sınanır. İkinci tarayıcı birincinin özel içeriğini doğrudan kimlikle de alamaz. Normal katalogdan geri çekilen görsel yeni aramada çıkmaz ve yeni edinime kapanır. Önceden edinmiş tarayıcı kalıcı adresinden yeniden indirebilir. Üretimin katalog için reddi sahibinin özel teslimini bozmaz.
- [ ] **K9 · Katkı ve kabul.** Toplu yüklemeye özgün, aynı byte’lara sahip, görsel olarak yakın fakat farklı ve kabul kuralını ihlal eden örnekler konur. Her dosya için ayrı sonuç gösterilir. API tam hash ile kesin kopyayı ayırır. Yakınlık tek başına kesin kopya kararı verdirmez. Açık paylaşım, kullanım izni ve Luna kabulü yayım için gereklidir. Ret, inceleme ve servis hatası ayrı sonuçlardır. Eşzamanlı kopyalarda bir katalog kaydı ve en fazla bir katkı ödülü oluşur.
- [ ] **K10 · Analiz ve kuyruk.** RabbitMQ görseli asset kimliği, nesne anahtarı ve sürümle taşır. Extract-worker S3’ten aldığı görsel için doğrulanmış Glimmer JSON’unu özel alana sürümlü yazar. Index-worker bu JSON’dan arama metnini/filtreleri, görselden ikinci vektörü üretir. Worker yazımdan sonra fakat ACK öncesinde durdurulup yeniden başlatılır. İş tamamlanır, dosya kaybolmaz, ikinci point veya ödül oluşmaz. Kalıcı geçersiz analiz sınırlı tekrar sonunda hata kuyruğunda görülür. İş sürerken geri çekilmiş görsel yayımlanmaz.

### Arama ve işletim

- [ ] **K11 · Qdrant ve filtreler.** Aynı point üzerinde Cosine `text[1024]` ve `image[768]` bulunur. Yerel profilde 16 shard, tek node üzerinde replika 1, binary quantization ve orijinal vektörlerle yeniden puanlama koleksiyon ayarlarından doğrulanır. Ana plandaki payload indeksleri ilk yüklemeden önce vardır. Filtreli her sonuç görünürlük, lisans, oran ve belirtilen diğer kesin koşulları karşılar. Her iki arama türünde güçlü eşiğin altındaki sonuçlar items dizisinden elenir. Yakın aday istenirse güçlü eşleşme bulunmadığında alt eşiği geçen en fazla bir aday nearest alanında döner. Snapshot ayrı koleksiyona geri yüklenip bilinen sorguyla sınanır.
- [ ] **K12 · Arama kalitesi.** Yaklaşık 100 etiketli Türkçe/İngilizce metin sorgusu ve ayrıca görsel referanslar hazırlanır. Kalibrasyonda kullanılan örneklerle kabul örnekleri ayrılır. Aşağıdaki önerilen kalite hedefleri ayrılmış sette sağlanır, eşikler ve örnek sayıları koleksiyon/model sürümüyle kaydedilir.
- [ ] **K13 · İstatistikler ve teşhis.** Bilinen iki aramada aynı görsel dönüp bir kez edinilir. API dönem filtresinde iki arama dönüşü ve bir ilk edinim gösterir. Tekrar indirme ve popüler listeyi açma sayıları değiştirmez. Keşfet’te son yüklenenler ve en çok edinilenler doğru sırada görünür. Beş uygulama servisinin her biri için Prometheus ölçümleri ve Grafana panosu açılır. Örnek bir istek veya iş kimliği ortak loglardan izlenir. Arama gecikmesi, boş sonuç, sağlayıcı maliyeti/hatası, kuyruk yaşı ve yayıma kadar geçen süre ilgili panoda gözlenir.

## Model kalitesi nasıl kabul edilecek?

Önerilen MVP hedefi, katalogda karşılığı olan sorguların **en az %80’inde** ilk beşte elle etiketlenen uygun görsellerden en az birini bulmaktır. Katalogda karşılığı olmayanların **en az %90’ında** boş sonuç beklenir. İki oran metin ve referans araması için ayrı hesaplanır. Değerlendirmeye her grupta hem cevaplı hem cevapsız örnek girer. Türkçe/İngilizce dağılımı, filtre birleşimleri, örnek adetleri ve başarısız örnekler raporlanır. Bu yüzdeler önerilen ürün kabul hedefleridir. Qdrant cosine eşikleri ölçümle ayrıca seçilir.

Aynı değerlendirmede binary arama tam hassasiyetli aramayla karşılaştırılır. Arama p95 süresi, ilk 100 temsili görselde tekrarlar dahil geçerli Glimmer analizi başına maliyet, OCR ve kompozisyon hataları kaydedilir. Yerel donanım ve sağlayıcı için ilk ölçümler bütçe ve gecikme kararlarına temel olur.

## Son demo ve bitiş kararı

Temiz bir tarayıcıda 100 krediyle başla. “Solunda başlık için boşluk olan 4:3 bir kafe fotoğrafı var mı?” diye **sesle** sor. Gelen ikinci uygun kartı edin ve iki kez indir (**99**). Aynı karttan açıkça varyasyon iste (**96**). Yeni bir görsel üret (**93**). İkinci bir tarayıcıdan kabul edilen üretimi katalogda bul. Orijinal dosya için edinim hakkını sına. İlk tarayıcıdan ayrıca kullanım izni olan benzersiz bir dış fotoğrafı yükle. Yayımlandığında bakiye **94** olur. İkinci tarayıcıdan yeni katkıyı bul. İlk tarayıcıda sohbeti yeniden aç ve kartları, kitaplığı, bakiyeyi ve işlem geçmişini karşılaştır. Katalog için önceden en az iki uygun 4:3 kafe görseli hazırlanır.

**Bitiş kaydı.** K1–K13 sonuçları ve son demo aynı teslim sürümü için doğrulandığında MVP tamamlanır. Açık kalan madde varsa hangi davranışın eksik olduğu, nasıl tekrarlandığı ve etkisi yazılır. Kalite ve maliyet tahminleri ölçüm yapılıncaya kadar tahmin olarak kalır.

### Yerel doğrulama kaydı · 24 Kasım 2026

- Çalışan API ve PostgreSQL üzerinde iki ayrı çerezle tarayıcı başına 100 kredi, tek başlangıç hareketi, aynı çerezle oturum devamlılığı ve başka tarayıcının özel sohbetine erişimde HTTP 404 doğrulandı.
- Ayrı PostgreSQL test veritabanında `TestSessionAndConcurrentReservations` geçti. Son üç kredi için iki eşzamanlı rezervasyon denemesi tek rezervasyon oluşturdu. Yinelenen anahtar ikinci rezervasyon açmadı. Kesin başarısızlık kullanılabilir üç krediyi geri verdi.
- Önceki dağıtık profilde (güncel K11 kanıtı kapsamı dışında) Qdrant HTTP yanıtında iki peer, 32 yerel ve 32 uzak shard kopyası, replika 2, iki Cosine vektör ve 17 payload indeksi görüldü. Garage hazırlama süreci çıkış kodu 0 ile kapandı.
- Bu sonuçlar K1, K7, K8 ve K11 için kısmi kanıttır. Canlı model, uçtan uca ürün ve arama kalitesi doğrulaması açık.
