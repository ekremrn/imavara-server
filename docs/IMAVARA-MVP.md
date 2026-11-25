# Imavara — Yerel MVP ürün ve mimari planı

Bu belge ürün kapsamını ve hedef mimariyi tanımlar. Mevcut proje klasöründe tasarım şemaları ve agent sistem prompt’u bulunur. Aşağıdaki servisler, endpoint adları ve dizinler uygulama tasarımını tarif eder.

## 1. Ürün ve başarı ölçüsü

Imavara’nın ilkesi **önce ara, gerektiğinde üret**. Kullanıcı bir görsele neden ihtiyaç duyduğunu anlatır. Sistem mevcut AI görselleri arasından kullanıma uygun adayları getirir. Kullanıcı orijinali alabilir, yakın bir adaydan varyasyon isteyebilir veya yeni bir görsel üretebilir. Açıkça paylaşılmış ve kabul edilmiş görseller gelecekteki aramalarda başka ihtiyaçları karşılar.

İlk sürümün ana sorusu şudur: Kullanıcı gerçek bir işte ihtiyacını mevcut bir görselle karşılayabiliyor mu? Bunu aramadan edinime dönüşüm, boş sonuç oranı, üretime geçiş ve görsel edinme maliyetiyle izleyeceğiz. İlk katalog için paylaşım ve kullanım koşulları belli, oran ve kompozisyon çeşitliliği olan bir başlangıç koleksiyonu gerekir.

**Örnek yolculuk.** Kullanıcı “Kafemin web sitesi için sıcak ışıklı bir kapak istiyorum” yazar. Agent gerekliyse başlığın nereye konacağını sorar. Kullanıcı “Sola” dediğinde gerçek katalog sonuçları gelir. Kullanıcı ikinci fotoğrafı 1 krediye alabilir veya “İkincideki fincanı kaldır, kadrajı koru” diyerek 3 krediye varyasyon isteyebilir. Aynı konuşmayı ertesi gün kart sırasıyla birlikte yeniden açabilir.

## 2. Yerel çalışma sınırı

- React arayüzü yerelde çalışır. API, iç MCP, web agent, iki worker ve veri servisleri Docker Compose ile çalıştırılır. Model çağrıları internet üzerinden OpenRouter’a gider. OpenRouter anahtarı sunucu tarafında kalır.
- Kullanıcı giriş yapmadan tarayıcısıyla başlar. Kalıcı, rastgele bir `HttpOnly` çerez PostgreSQL’deki anonim tarayıcı kaydına bağlanır. HTTP ve WebSocket aynı tarayıcı bağlamını kullanır. Sohbet, bakiye, kitaplık ve indirme işlemlerinde API bu kaydı esas alır. Agent’a hesap veya tarayıcı seçtirilmez.
- İlk MCP istemcisi Imavara’nın kendi LangChain agent’ıdır. MCP’ye Docker’ın özel ağından erişilir. Başka bir özel ağ istemcisine erişim açılması daha sonraki bir kurulum tercihidir.
- Keycloak, OAuth, roller, SMTP/Mailpit, Kubernetes, Netlify, production dağıtımı ve kredi satışı yerel MVP kapsamı dışındadır. Üretim ortamı gündeme geldiğinde kimlik, dağıtım ve Sentry tekrar ele alınır. SVG/komponent kütüphanesi, video ve diğer medya türleri ileri fazlardır.

## 3. Arayüz ve konuşma

**Açılış.** İlk ekran arama motoru sadeliğindedir. Ortada küçük “imavara.” logosu ve tıklanıp yazılabilen tek bir giriş alanı bulunur. Alan referans görsel eklemeyi destekler. İlk mesaj gönderilince kullanıcı metni konuşmanın ilk mesajına dönüşür, yazma alanı kısa bir geçişle aşağı yerleşir ve sohbet açılır. Klavye odağı korunur. Azaltılmış hareket tercihinde geçiş doğrudan yapılabilir.

**Sesle anlat.** Açılış alanında ve sohbetin mesaj alanında mikrofon düğmesi bulunur. Kullanıcı “Sesle anlat”a dokununca kayıt başlar (ilk kullanımda mikrofon iznini verir). Kayıt süresiyle birlikte “Gönder” ve “İptal” gösterilir. “Gönder” kaydı bitirir ve yazıya çevirmeyi başlatır. Uzun kayıtlarda “Ses yazıya çevriliyor…” durumu görünür. Transkript hazır olunca varsa yazılı taslakla birleştirilir ve seçili referans görselle birlikte otomatik olarak konuşmada tek kullanıcı mesajı olarak görünür. Agent kataloğu aramaya başladığında arayüz “Katalogda aranıyor…” durumunu gerçek araç olayından gösterir. İptal, reddedilen mikrofon izni, boş transkript ve transkripsiyon hatasında mesaj oluşturulmaz, önceki taslak korunur.

**Sohbet.** Agent kullanım amacı, oran, stil, kadraj veya boş alan gibi kararı etkileyen bir eksik varsa kısa bir soru sorar. İstek yeterince açıksa kataloğu arar. Metinle birlikte gerçek görsel kartları gösterilir. Sonuçların sırası konuşmada saklanır. Kullanıcı yeni sohbet açabilir, eski sohbet ve görsellerine dönebilir. Sohbeti silme ile kitaplıktaki görseli kaldırma ayrı işlemlerdir. [Agent sistem prompt’u](../agent/prompts/system.tr.txt)

**Keşfet.** İkinci görünüm agent turu gerektirmeden doğrudan arama, referansla benzerlerini bulma ve filtreleme sunar. İlk görünümünde son eklenenler ve en çok edinilenler bulunur. Kart orijinal kadrajı, önizlemeyi, oranı ve boyutu gösterir. “Orijinali al · 1 kredi”, “Benzerlerini bul” ve “Varyasyon üret · 3 kredi” eylemleri ilgili görsele bağlanır. Katkı ekranı toplu dosya seçimini ve her dosyanın sonucunu gösterir.

**Görsel dil.** Sıcak arka plan `#F7F6F2`, beyaz yüzey `#FFFFFF`, ana yazı `#202522`, ikincil yazı `#58635D`, petrol yeşili vurgu `#176454`, seçili yüzey `#E3EFE9`. Logo küçük harfli ve noktalı. Arayüz için DM Sans, açılış sorusu kullanılırsa Instrument Serif önerildi. Yerleşim ve animasyon ayrıntıları arayüz görüşmesinde kesinleşecek. Açılış ekranının alt kısmında küçük, ikincil metin olarak “Sohbet, görsel üretimi ve analizde Meta Muse modellerini kullanıyoruz.” ifadesi yer alır.

## 4. Tarayıcı kredisi ve erişim

| İşlem | Demo kredisi | Kayıt davranışı |
|---|---:|---|
| İlk tarayıcı oturumu | +100 | Tek seferlik açılış hareketi |
| Sohbet, sesle yazma, tarif geliştirme, arama ve önizleme | 0 | Boş sonuç da ücretsiz |
| Mevcut orijinali ilk kez alma | −1 | Tarayıcı ve görsel çifti için tek edinim |
| Edinilmiş dosyayı tekrar indirme | 0 | Aynı kalıcı adres ve hak kullanılır |
| Yeni üretim veya varyasyon | −3 | Teslim edilebilen çıktı için tek ücret, kaynak görseli ayrıca edinme bedeli yok |
| İlk kez yayımlanan benzersiz dış katkı | +1 | İçerik başına tek ödül |

Kredi bir Imavara ürün birimidir. Sağlayıcı maliyeti ayrı izlenir. Sekmeler ve yeni sohbet aynı bakiyeyi paylaşır. Çerez silinirse, gizli pencere veya başka profil kullanılırsa yeni demo oturumu oluşur. Önceki sohbetler eski oturuma bağlı kalır. Bu davranış güvenilen yerel demoda kabul edilmiştir.

API bakiyeyi ve hareketi PostgreSQL transaction’ında yönetir. Üretimden önce 3 kredi ayrılır. Orijinal storage’a yazılıp teslim edilebilir olduğunda harcama kesinleşir. Kesin başarısızlıkta tutar serbest kalır. İstek veya WebSocket yeniden gönderilirse işlem anahtarı çift üretim ve harcamayı önler. Eşzamanlı iki istek aynı son kredileri kullanamaz.

## 5. Servisler ve modeller

| Süreç | Sorumluluk |
|---|---|
| Go API | Tarayıcı bağlamı, HTTP/WebSocket, ses kaydını yazıya çevirme, arama, edinim, üretim, katkı, istatistik ve kredilendirme |
| Go MCP | İç ağdaki araç çağrısını mevcut API işlemine aktarır. Ürün kuralları API’de uygulanır. |
| TypeScript web agent | LangChain ile konuşmayı ve MCP araçlarını yürütür. Gerçek kartları, araç olaylarını ve kullanıcıya yönelik metni API’ye döndürür. |
| Go image-extract-worker | Analiz kuyruğundan aldığı görselin açıklama ve özniteliklerini çıkarır. |
| Go index-worker | Analizi standartlaştırır, vektörleri ve filtre alanlarını Qdrant’a yazar, yayımlamayı tamamlar. |

| OpenRouter model kimliği | Görev |
|---|---|
| `meta/muse-ember` | Sohbet ve araç seçimi |
| `openai/gpt-transcribe` | Ses kaydını özgün dilinde sohbet mesajına çevirmek |
| `meta/muse-image` | Prompt’tan üretim ve referansla düzenleme |
| `openai/gpt-6-luna-decisions` | Ortak katalog kabul kurallarını değerlendirme |
| `meta/muse-glimmer-30b` | Görselden yapılandırılmış analiz |
| `qwen/qwen3-embedding-8b` | Açıklama ve metin sorgusu için embedding |
| `google/gemini-embedding-2` | Görsel ve referans görsel için embedding |

## 6. Uçtan uca işlemler

**Arama.** API, metin sorgusunu Qwen ile, dış referansı Gemini ile embed eder. Katalogdan seçilen bir referansın erişilebilir kayıtlı vektörü yeniden kullanılır. Kesin filtreler, görünürlük ve ilgili skor eşiği uygulanır. API ve MCP’de isteğe bağlı `top_k` 1–10 arasında bir tam sayıdır (varsayılan 5). En fazla istenen sayıda önizleme, dosya özellikleri ve kullanım bilgileri döner. API güçlü eşleşmeleri `items` içinde döndürür. İstek `include_nearest: true` içeriyorsa ve güçlü eşleşme yoksa tüm kesin filtreleri geçen tek aday, alt eşiği geçtiğinde `nearest` alanında gelir. MCP ve web bu seçeneği kullanır. Yanıtta `score_threshold`, `candidate_threshold`, `metric` ve `vector`, kartta `match.score`, `match.percent` ve `match.level` bulunur. Yüzde, kosinüs skorunun yuvarlanmış gösterimidir. Boş liste başarılı bir yanıttır. Otomatik ücretli üretim başlatılmaz.

**Orijinali alma.** API ilk edinim ve kredi hareketini birlikte kaydeder. Kalıcı bir Imavara adresi verir. İndirmede tarayıcı hakkını kontrol edip Garage/S3’ten dosyayı sunar. Normal katalogdan çekmede yeni arama ve edinimler kapanır, önceki edinimlerin yeniden indirme hakkı korunur. Hak ihlali nedeniyle kaldırma ayrı ele alınır.

**Üretim.** API prompt’u, oranı ve varsa erişilebilir referansı Muse Image’a gönderir. Orijinal önce S3’e yazılır. Kullanıcı aynı istekte çıktı alır. Her üretim ve varyasyon teslimle birlikte otomatik olarak ortak katalog kuyruğuna girer. Kabul edilen görsel indeksleme tamamlandığında herkesçe aranabilir. Katalog kalitesinden dolayı reddedilmesi özel teslime engel olmaz. Geçici ve güvenle tekrarlanabilen hatalarda aynı istek süresi içinde en fazla iki tekrar önerilir. Sağlayıcının işi alıp almadığı belirsizse körlemesine yeni üretim başlatılmaz.

**Dış katkı.** Web aynı seçimdeki kopyaları erkenden ayıklayabilir. API dosyayı doğrular, boyutları okur ve ham byte’lardan SHA-256 hesaplar. Qdrant’ta içerikten türetilen sabit point kimliğiyle birebir kopya kontrolü yapılır. Görsel embedding’iyle yakın adaylar bulunabilir. Küçük ama anlamlı bir varyasyon yalnızca yüksek benzerlik nedeniyle birebir kopya sayılmaz. Eşzamanlı yükleme ve tekrarlar çift katalog kaydı veya çift ödül oluşturmamalıdır.

**Kabul.** Dış katkılar ve otomatik sıraya alınan üretimler Luna değerlendirmesinden geçer. Aynı dosya katalogda tekilleştirilir. Sürümlü sorular örneğin belirgin filigranı, okunabilir özel bilgiyi, ağır bozulmayı ve kapsam dışı içeriği değerlendirir. API karar skorlarını kabul, ret veya inceleme sonucuna ve gerekçe kodlarına dönüştürür. Model erişim hatası katalog reddinden ayrı bildirilir. Paylaşım izni ve kullanım koşulları ayrıca kaydedilir.

**Analiz ve indeksleme.** Kabul ile RabbitMQ’ya yayınlanacak iş birlikte kalıcı olarak kaydedilir. Kuyruk mesajı görsel byte’ları yerine `asset_id`, `object_key` ve sürümü taşır. Analiz worker’ı S3’ten okur, Glimmer JSON’unu doğrular ve sürümlü saklar. İndeks mesajının yayın onayından sonra ilk mesajı ACK eder. İndeks worker’ı filtreleri ve embedding’leri hazırlar, güncel paylaşım durumunu kontrol eder, Qdrant’a yazar. Yayımlama ve varsa dış katkı ödülü kaydedildikten sonra ACK verir. Tekrarlar aynı görsel ve sürüm üzerinde güvenlidir. Sürekli başarısız işler hata kuyruğunda incelenir.

**Sesli giriş.** Tarayıcı “Gönder” ile kaydı kapatır ve mevcut oturum üzerinden Go API’ye gönderir. API ses biçimini ve istek boyutunu doğrular, OpenRouter’dan metni alıp tarayıcıya döndürür. Ham ses yalnızca transkripsiyon süresince işlenir. Tarayıcı boş olmayan metni aynı konuşmaya (açılışta yeni konuşmaya) mevcut WebSocket akışından otomatik gönderir. Gönderim kimliği çift dokunmada veya bağlantı tekrarında aynı kayıttan ikinci mesaj ve ikinci agent turu oluşmasını önler. Ürün kredisi ve araç çağrısı kuralları yazılı mesajla aynıdır.

**Canlı sohbet.** React ve Go API arasında WebSocket; Go API ile iç agent arasında olay akışı vardır. Web agent, MCP üzerinden aynı API işlemlerini kullanır. Arayüz gerçek araç başlangıcı, dönen kartlar, metin parçaları, teslim ve hata olaylarını gösterir. Mesaj kimliği, ücretli araç işlem anahtarı ve olay sıra numarası tekrarları ayırt eder. Yeniden bağlantıda API tamamlanmış mesaj, kart ve işlem durumunu PostgreSQL’den yükler.

## 7. Veri, filtreler ve ölçüm

- **Garage/S3.** Orijinaller, açık önizlemeler ve özel analiz dosyaları ayrı erişim alanlarında tutulur. Sürümlü JSON tekil analiz ve yeniden indeksleme içindir. Toplu analitik dışa aktarım için Parquet düşünülmüştür. Dosya referansı ve hash yeterlidir; bağımsız veri kümesi dışa aktarımında görsel byte’ları ayrıca eklenebilir. Qdrant snapshot’ları da saklanabilir.
- **Qdrant.** Her katalog görseli bir point’tir. Aynı point üzerinde iki ayrı named vector bulunur: `text` ve `image`. Model, boyut ve benzerlik eşiği alan başına seçilir. Float orijinallerle yeniden puanlama ve binary quantization hedeflenir. İlk toplu yüklemeden önce kullanılan payload filtrelerine indeks açılır.
- **Glimmer alanları.** Açıklama, ana konu, nesneler, insan varlığı, sahne, perspektif, ışık, atmosfer, stil, renk ailesi, ana konunun ve boş alanın konumu, okunan yazı, etiketler, olası kullanım alanları. Kontrollü filtrelerde belirsizlik için `unknown` desteklenir.
- **Kesin alanlar.** Genişlik, yükseklik, oran, dosya biçimi, alfa, byte boyutu ve hash dosyadan gelir. Görünürlük, kaynak, varsa `parent_asset_id`, model, paylaşım ve kullanım izinleri uygulama kaydından gelir. Qdrant filtreleri örneğin oran, asgari çözünürlük, stil, renk, boş alan, insan/yazı varlığı, görünürlük ve lisansı kapsar.
- **PostgreSQL.** Tarayıcı oturumu, kredi hareketleri, edinimler, konuşmalar, mesajlar, kart sırası, yayımlama ve arama olayları burada tutulur. Bir aramaya bir `search_id`, dönen her adaya bir `(search_id, asset_id, sıra)` kaydı yazılır. İlk edinim `(browser_id, asset_id)` çifti için tekildir. Tarih aralığında ilk sayaç “en sık dönenleri”, ikincisi “en çok edinilenleri” verir. Tekrar indirme ve popüler listeyi görüntüleme bu sayıları şişirmez.
- **Gözlemlenebilirlik.** Ortak Go logger’ı, istek ve işlem kimlikleri kullanılır. Prometheus ve servis başına Grafana panoları arama süresi, boş sonuç, model hatası, maliyet, kuyruk yaşı ve yayına kadar geçen süreyi izler. Yerelde Sentry anahtarı gerekmez.

## 8. Arayüz sözleşmeleri

Aşağıdaki yollar önerilen MVP sözleşmesidir. `/me` aktif tarayıcıyı ifade eder. API ve MCP aynı fiyat ve işlem kurallarını uygular.

| API | İşlev | MCP aracı |
|---|---|---|
| `GET /v1/me`, `GET /v1/pricing`, `GET /v1/me/credits/transactions` | Bakiye, tarife ve geçmiş | `get_balance` bakiye ve tarifeyi sunar |
| `POST /v1/search` | Metin veya referansla ücretsiz arama | `search_images` |
| `GET /v1/assets`, `GET /v1/assets/{id}` | Galeri ve görsel detayı | `get_image` detay için |
| `POST /v1/assets/{id}/acquire`, `GET /v1/assets/{id}/original` | İlk edinim ve orijinal dosya | `acquire_image` |
| `POST /v1/generations` | Üretim veya referansla varyasyon | `generate_image` |
| `POST /v1/contributions` | Dosya bazında katkı, toplu seçimde ayrı sonuçlar | `contribute_image` |
| `GET /v1/me/assets`, `PATCH /v1/assets/{id}`, `DELETE /v1/assets/{id}` | Kitaplık, paylaşım ve kaldırma | Web |
| `GET /v1/stats/assets` | `acquisitions` veya `search_returns` ve dönem filtresi | Web |
| `POST /v1/audio/transcriptions` | Tarayıcıdan multipart kaydı alır ve `{"text":"…"}` döner. Web bu metni sohbet kanalına otomatik gönderir. | Web |
| Konuşma oluşturma, listeleme, açma, silme ve WebSocket | Persist edilen geçmiş, yeni mesaj ve canlı olaylar | Web agent girişi |

## 9. İki repo ve kod sahipliği

| Repo / hedef dizin | İçerik |
|---|---|
| `imavara-web/src/app`, `features/{chat,discover,assets,contributions,library}` | Yönlendirme, açılıştan sohbete geçiş, Keşfet, kitaplık ve katkı |
| `imavara-web/src/components`, `src/api` | Ortak görsel kartları, HTTP ve WebSocket istemcileri |
| `imavara-server/cmd/{api,mcp,image-extract-worker,index-worker}` | Tek Go modülünden dört bağımsız süreç ve container |
| `imavara-server/internal/{catalog,search,credits,generation,chat,extraction,indexing}` | İş kuralları ve süreçlerin uygulaması |
| `imavara-server/internal/platform` | S3, RabbitMQ, OpenRouter, veritabanı bağlantıları ve gözlemlenebilirlik |
| `imavara-server/agent` | Kendi paket tanımı ve Docker imajı olan TypeScript/LangChain servisi, MCP istemcisi ve sistem prompt’u |
| `imavara-server/contracts`, `migrations`, `deploy`, `docker-compose.yml` | Sürümlü HTTP/olay/kuyruk şemaları, PostgreSQL migration’ları, Qdrant ve yerel servis ayarları |

API, kredilerin ve görsel erişiminin sahibidir. MCP araçları ve agent bu kuralları çoğaltmaz. Web ile backend arasındaki bağımsızlık sürümlü sözleşmelerle korunur. Yerel kurulumun odağı Compose’tur.

## 10. Ağ ve storage sözleşmesi

Tarayıcıdan erişilmesi gereken backend yüzeyi Go API'dir. Yerel öneri `127.0.0.1:8080`; React geliştirme sunucusu `/v1` ve WebSocket yolunu bu adrese proxy edebilir. Böylece web tek origin ve aynı çerezle çalışır. Bir private LAN demosunda tarayıcılara yalnızca web girişi ve gerektiğinde API adresi duyurulur. Docker ağında API ve worker'lar Garage'a örneğin `http://garage:3900` üzerinden erişir.

| Trafik | MVP'deki yol |
|---|---|
| Tarayıcıda önizleme / orijinal | API, kartta önizleme URL'si döndürür. Kalıcı orijinal yolu `GET /v1/assets/{id}/original` tarayıcı hakkını kontrol edip dosyayı aktarır. |
| Toplu katkı | Tarayıcı dosya başına API'ye gönderir. API gerçek biçimi doğrular, hash alır, S3'e kaydeder. |
| Araçlar ve kuyruk | MCP, agent, RabbitMQ, Qdrant ve PostgreSQL iç ağdadır. Yönetim portları ürün yüzeyine dahil edilmez. |
| OpenRouter'a görsel girişi | Yerel dosyayı API/worker iç storage'dan okur; ilgili endpoint'in desteklediği inline görsel biçiminde dışarı giden HTTPS çağrısına ekler. Bu taşıma yalnızca sağlayıcı sınırındadır; RabbitMQ'da nesne anahtarı kalır. Endpoint için inline desteği ilk entegrasyonda doğrulanır. Sağlayıcı yerel Docker URL'sini çözemez. |
| İsteğe bağlı ikinci adres | Doğrudan medya sunma ihtiyacı çıkarsa yalnızca açık önizlemeler için salt okunur bir medya URL'si tanımlanabilir. Sağlayıcının URL gerektirdiği özel dosya için dosyaya özel, dışarıdan erişilebilir süreli okuma URL'si ayrı seçenektir. S3 yönetim/yazma yüzeyi, özel orijinaller, Parquet ve yedekler iç erişimde kalır. |

## 11. Qdrant koleksiyon sözleşmesi

İlk yüklemeden önce tek seferlik bir kurulum adımı koleksiyonu, iki named vector alanını ve aşağıdaki payload indekslerini oluşturur. Filtreli HNSW grafiği bu indeksleri kullanarak inşa edilir. Servis başlatılırken mevcut şema doğrulanır. Sürüm değişiminde yeni koleksiyona yeniden indeksleme ve doğrulama ardından okuma alias'ını taşıma öngörülür. [Qdrant indeksleme](https://qdrant.tech/documentation/manage-data/indexing/)

| Alan | Tür | Kullanım |
|---|---|---|
| `owner_browser_id` | uuid | Tarayıcıya ait özel kitaplık indekslendiğinde sahiplik filtresi |
| `visibility`, `catalog_status`, `license_scope` | keyword | Her genel aramada yalnızca yayımlanmış, ortak ve kullanıma izinli kayıtlar |
| `aspect_ratio`, `orientation`, `mime_type` | keyword | Örn. 4:3, yatay, image/png |
| `style`, `theme`, `color_family`, `negative_space` | keyword, gerektiğinde değer dizisi | Kontrollü sözlükle estetik ve kompozisyon filtreleri |
| `width`, `height` | integer | Piksel cinsinden minimum veya aralık sorgusu |
| `has_people`, `has_text`, `has_alpha` | bool | Gerçekten belirlenmiş özellikler; analiz belirsizse alan boş bırakılır. |
| `published_at` | datetime | Yayın zamanı aralıkları ve yeni görseller |
 
Bir dosyanın tam SHA-256 değeri payload'da tutulur. Kesin eşleşme, hash'ten türetilen UUID ile point okuması ve tam hash karşılaştırmasıdır. Kopya kontrolü için yoğun vektör veya tüm açıklamaya full-text indeks gerekmez. Aynı özel dosyanın varlığı başka tarayıcıya açıklanmaz. Her görselin ortak katalog kimliği, tarayıcı bazlı edinim kayıtlarından ayrıdır.

**Sunucu arama profili.** `text` 1024 ve `image` 768 boyutlu float32, ikisi de Cosine. İlk deneme için orijinal vektörler diskte, 1-bit binary temsil bellekte; HNSW `m=16`, `ef_construct=100`, sorguda `hnsw_ef=128`, quantization `oversampling=3.0` ve `rescore=true`. Bunlar Imavara için başlangıç ayarlarıdır. Sunucu ilgili vektör alanını seçer, payload filtrelerini uygular, orijinal vektörlerle yeniden puanlanan adaylarda alana özgü minimum skoru uygular ve istenen `top_k` kadar aday dönebilir. [Qdrant quantization](https://qdrant.tech/documentation/manage-data/quantization/)

**Kalibrasyon.** Yaklaşık 100 Türkçe/İngilizce metin sorgusu ve ayrıca görsel referanslar hazırlanır. Katalogda karşılığı olmayan istekler, sol/sağ boşluk gibi zor ayrımlar ve sıkı filtre birleşimleri sete girer. Tam hassasiyetli kesin aramaya göre ilk beşte isabet, hatalı sonuç dönen boş sorgu oranı ve p95 gecikme ölçülür. Metin ve görsel için iki ayrı eşik ayarlanır, ayrılmış örneklerde doğrulanır ve koleksiyon/model sürümüyle kaydedilir. Arama eşiği, katalog kabulü ve yakın kopya tespiti bağımsız kararlardır.

**Yerleşim.** Yerel MVP tek node ve tek veri volume’ü kullanır. Koleksiyon 16 shard ve replika faktörü 1 ile oluşturulur. Snapshot geri yüklemesiyle yeni bir koleksiyonun aranabilirliği ayrıca sınanır.

## 12. Sunucu model profili v1

Aşağıdakiler uygulama için önerilen sabitlerdir. HTTP ve MCP sözleşmeleri yalnızca ürün girdilerini kabul eder: tarif, filtre, referans görsel ve istenen oran gibi. Model adı, sağlayıcı, sampling, reasoning, boyut ve ham sağlayıcı parametreleri için istemci override'ı sunulmaz. Profil sürümü, kullanılan model, istek maliyeti ve süre sunucuda izlenir.

| Model | Başlangıç profili ve gerekçesi |
|---|---|
| `meta/muse-ember` | `temperature=0.3`, `reasoning.effort=low`, `tool_choice=auto`, streaming açık. Araç çağrıları ve referans seçiminde tutarlılık, kısa yanıtta doğal dil hedeflenir. Sabit thinking token sayısı gönderilmez. Sunucu yalnızca gerçek araç olayları ve kullanıcıya yönelik yanıtı WebSocket'e aktarır. MVP boyunca effort sabit kalır. [OpenRouter endpoint özellikleri](https://openrouter.ai/api/v1/models/meta/muse-ember/endpoints) · [Ember reasoning seviyeleri](https://docs.cloud.google.com/gemini-enterprise-agent-platform/models/partner-models/meta/muse-ember-1-0) |
| `openai/gpt-transcribe` | Go API, tamamlanan kaydı OpenRouter `POST /api/v1/audio/transcriptions` yoluna `input_audio.data` ve `input_audio.format` ile gönderir. Kayıt süresi, transkripsiyon gecikmesi ve sağlayıcı maliyeti izlenir. İlk uçtan uca denemede Türkçe bir görsel tarifi, oran ve yön sözcükleriyle gerçek tarayıcı kaydından gönderilir. [OpenRouter model kartı](https://openrouter.ai/openai/gpt-transcribe) · [Transkripsiyon API’si](https://openrouter.ai/blog/tutorials/transcription-on-openrouter/) |
| `meta/muse-image` | İşlem başına bir çıktı. Sıfırdan üretimde referans yok; varyasyonda bir referans ve korunacak özellikler açıkça iletilir. Kullanıcı oran belirtmezse yeni üretimde hedef 1:1, varyasyonda kaynağın oranı. Temperature, top_p, seed, kalite ve thinking için özel değer gönderilmez. Referans ve oran desteği seçilen Images endpoint'inde doğrulanır; gerçek piksel boyutu dosyadan okunur. Kesin oran karşılanmıyorsa bunu kullanıcıya bildir. [OpenRouter Muse Image](https://openrouter.ai/meta/muse-image) · [Image API endpoint sözleşmesi](https://openrouter.ai/blog/announcements/image-api/) |
| `qwen/qwen3-embedding-8b` | Hedef 1024 boyut. Tam 4096'ya göre float vektör alanını dörtte bire indirir. Sorgu için sabit retrieval talimatı, doküman için Glimmer açıklaması + kompozisyon + kontrollü özellikler. Sağlayıcı istenen boyutu vermiyorsa modelin MRL desteğinden yararlanarak tam çıktının ilk 1024 bileşenini alıp L2 normalize etme yolu doğrulanabilir. API ve index-worker aynı dönüşümü kullanır. [Qwen model kartı](https://huggingface.co/Qwen/Qwen3-Embedding-8B) |
| `google/gemini-embedding-2` | Hedef 768 boyut; Google'ın önerdiği kompakt boyutlardan biri. Aynı model ve dönüşüm katalog ile dış referansa uygulanır. Dönen uzunluk ve sonlu değerler doğrulanır; yanlış boyut indekse yazılmaz. Rota uyumu gerçek bir görselle sınanır. [OpenRouter Gemini Embedding 2](https://openrouter.ai/google/gemini-embedding-2) |
| `meta/muse-glimmer-30b` | Önerilen sabit analiz profili: `reasoning.effort=low`, `reasoning.exclude=true`, `temperature=0.1`, `stream=false`, araç tanımı yok. Worker, v1 şemasını sistem talimatına ekler, `response_format.type=json_object` ile yanıt alır ve tam şemayı yerelde doğrular. Canlı denemede sağlayıcı tarafındaki strict şema, caption ve description alanlarında yalnızca üç nokta üretti. Açıklama kalitesi için yer tutucular reddedilir. Yönlendirmede `provider.require_parameters=true` kullanılır. [OpenRouter rota özellikleri](https://openrouter.ai/api/v1/models/meta/muse-glimmer-30b/endpoints) |
| `openai/gpt-6-luna-decisions` | Sürümlü adlandırılmış evet/hayır soruları; chat sampling ayarı gönderilmez. Her kural için skor ve gerekçe kodu kaydedilir. Eşikler etiketli katalog örnekleriyle belirlenecek, kararsız alan incelemeye gidecek. [Decisions model kartı](https://openrouter.ai/openai/gpt-6-luna-decisions) |

**Kabul örnekleri.** Tarayıcının aynı isteği tekrarlaması ikinci ücret oluşturmaz. Sohbette “ikincisine benzer” mevcut kartın kimliğini kullanır ve ücretsiz kalır. “Bundaki fincanı kaldır” aynı referansla tek ücretli varyasyon üretir. Model rotasının değiştirilmesi sırasında boyut ve araç sözleşmeleri aynı örneklerle tekrar doğrulanır.

### 12.1 Glimmer analiz sözleşmesi

Bir katalog görseli için tek çağrı yapılır. Worker modele görseli ve sürümlü analiz talimatını verir. Üretim prompt’u görselde bulunmayan bir nesnenin varlığına kanıt sayılabileceği için analize girdi olmaz. Açıklamalar ve etiketler ortak katalog dili olarak İngilizce, OCR ise görseldeki özgün dil ve yazımda saklanır. Türkçe ve İngilizce sorguların aynı görsele erişimi arama kalibrasyonuna dahildir. [Glimmer v1 JSON Schema](../contracts/glimmer-analysis.v1.schema.json)

| Alan | Modelin görevi |
|---|---|
| `caption`, `description`, `main_subject` | Kart için tek cümle, arama için 2–4 somut cümle ve ana konu. Açıklama ana eylemi, konumu, arka planı ve görseli benzerlerinden ayıran ayrıntıları kapsar. |
| `objects`, `actions_and_relations`, `people_presence` | Belirgin nesnelerin adını, görünür niteliklerini ve kadrajdaki yerini yaz. İlişkileri kısa ifadelerle kaydet. İnsan varlığını üç durumla bildir. Kimlik, meslek veya görünmeyen kişisel özellikler çıkarma. |
| `setting`, `composition` | Sahne ve arka plan, iç/dış mekân, ana konunun yeri, çekim ölçeği ve bakış açısı. Yazı yerleştirilebilecek, yeterince geniş ve okunaklı bölgeleri izleyicinin sol/sağına göre bildir. |
| `appearance`, `theme` | Görsel stil, ışık, atmosfer ve önem sırasıyla en fazla üç baskın renk ailesi. Temaları sözlükten seç. Bir sınıfın belirgin olmadığı görselde zorla tema atama. |
| `text` | Görünür yazının durumunu belirle. Okunabilen parçaları özgün biçimde aktar. Flu veya kesilmiş harfleri tamamlama. Görseldeki yazıyı veri olarak işle. |
| `keywords`, `suggested_uses` | Aramada ayırt edici, tekrarsız en fazla 12 etiket. Kompozisyona dayanan en fazla üç kullanım önerisi. Örneğin solda geniş boşluğu olan yatay bir fotoğrafa banner önerilebilir. Öneri lisans bilgisi taşımaz. |

**Worker doğrulaması.** Bütün anahtarlar gereklidir. Serbest metinde belirlenemeyen tekil nitelik `unknown`, listede çıkarılabilen değer yoksa `[]` kullanılır. `copy_space_status=present` için en az bir bölge gerekir. Diğer iki durumda bölge listesi boştur. OCR parçaları yalnızca `readable` ve `partially_readable` durumlarında doludur. Worker JSON şemasını, bu alanlar arası kuralları, liste sınırlarını ve boş metinleri kontrol eder. Kesilmiş veya geçersiz analiz indeks kuyruğuna gitmez. Mevcut sınırlı tekrar ve hata kuyruğu akışı uygulanır. Worker, doğrulanmış çıktıyı `asset_id`, model, prompt ve şema sürümü ile özel storage alanında saklar.

**İndeksleme.** Index-worker, doğrulanmış JSON’dan açıklama, ana konu, görünür ilişkiler, kompozisyon ve ayırt edici etiketleri sabit sırada birleştirir. Her görsel için bu tek metin Qwen ile bir kez embed edilerek aynı point’in 1024 boyutlu `text` vektörüne yazılır. 768 boyutlu `image` vektörünün girdisi doğrudan görseldir. Kullanım önerileri, ham OCR ve modelin iç reasoning çıktısı bu metne eklenmez. İndeks worker’ı stil, tema ve renkleri payload’a eşler. Boş alan bölgeleri `negative_space` alanına gider. İnsan ve yazı varlığı kesinleşince bool yazılır. Okunamayan ama varlığı görülen yazı `has_text=true`, belirsiz varlık ise eksik alan üretir. Piksel boyutları, hash, lisans ve katalog kararı mevcut sahiplerinden alınır.

**İlk deneme.** Sol/sağ boşluğu ters iki kompozisyon, yalnız eli görünen kişi, yazısız yüzey, flu yazı ve kalabalık sahne örnekleriyle şema geçerliliği, yön ve OCR doğruluğu, analiz başına maliyet ve süre ölçülür. MVP analizleri ve tekrarları sabit `low` kullanır. Chat Completions çağrısında `exclude=true` iç reasoning metnini yanıttan çıkarır. Üretilmiş reasoning tokenları yine çıktı tarifesiyle ücretlenir. İlk 100 temsili görselde tekrarlar dahil toplam Glimmer maliyeti ve geçerli analiz sayısı kaydedilir. Benzer bir katalog için 100.000 görselin tahmini, tamamlanan analiz başına ortalama maliyet üzerinden hesaplanır. Profil değişirse sürümlenir. [OpenRouter reasoning ücretlendirmesi](https://openrouter.ai/docs/guides/best-practices/reasoning-tokens)

## 13. Mimari çizimler ve doğrulanacak kararlar

[Yerel API](mimari/01-api.png) · [İç MCP](mimari/02-mcp.png) · [Görsel analiz worker’ı](mimari/03-extract.png) · [İndeks worker’ı](mimari/04-index.png) · [Web agent ve sohbet](mimari/06-agent.png) · [Yerel sistemin tamamı](mimari/05-genel.png)

| Konu | Bugünkü durum |
|---|---|
| Qdrant yerleşimi | Yerel varsayılan tek node, tek volume, 16 shard ve replika faktörü 1. |
| Arama kalitesi | Sunucu profilinde metinde 1.024, görselde 768 boyut ve cosine benzerliği. Sağlayıcının döndürdüğü boyutlar, binary quantization etkisi ve iki ayrı skor eşiği ilk katalogda doğrulanır. 0,80 sabit bir karar değildir. |
| Katalog kabulü | Luna kural listesi ve kabul/inceleme/ret sınırları örnek görsellerle ayarlanır. Yakın görsellerde otomatik kopya davranışı da aynı koleksiyonda denenir. |
| Özel kitaplık | Önceki konuşma ve görsellere erişim kapsamda. Paylaşılmayan tüm özel görseller için semantik indeksin hangi aşamada oluşturulacağı ayrıca seçilir. |
| Saklama ve başlangıç | Otomatik sohbet saklama süresi, başlangıç koleksiyonunun kaynağı, inline görsel desteğinin uçtan uca testi ve gerçek sağlayıcı bütçesi doğrulanacak ayrıntılardır. |

Uygulama sırası, tamamlanma koşulları ve son demo [MVP teslim planı ve kabul kriterlerinde](MVP-TESLIM-PLANI.md) yer alır.
