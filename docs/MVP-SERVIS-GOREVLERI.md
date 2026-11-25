# Imavara MVP — servis görevleri ve uygulama yönergesi

MVP’nin beş backend süreci Go API, Go MCP, TypeScript web agent, görsel analiz worker’ı ve indeks worker’ıdır. React uygulaması ayrı repoda çalışır. Aşağıdaki dizinler hedef kod yerleşimini, boş kutular yapılacak işleri gösterir. K numaraları [teslim planındaki kabul kontrollerine](MVP-TESLIM-PLANI.md) karşılık gelir. Endpoint’ler, model profilleri ve filtre tipleri için [ürün ve mimari planını](IMAVARA-MVP.md) kullan.

## Hangi sırayla ilerlenir?

| Adım | Görevler | Sonraki adıma geçerken denenebilir sonuç |
|---|---|---|
| 1. Ortamı hazırla | ORT-01–05 | Boş diskle Compose başlar. Koleksiyon ilk yüklemeye hazırdır. Yedi modelin gerekli girdiyi kabul ettiği küçük gerçek örneklerle görülür. |
| 2. Oturumu ve dosya erişimini kur | API-01–03 | İki tarayıcı oturumunda bakiye, sahiplik ve yeniden indirme API üzerinden sınanabilir. |
| 3. İlk katkıyı katalogda bul | API-04, EXT-01–03, IDX-01–04, API-05 | İzinli bir görsel yüklenir, işlenir ve ikinci tarayıcıda filtreli aramayla bulunur. Aynı dosyayı tekrar yükleyerek ödülü kontrol et. |
| 4. Kataloğu web’e bağla | API-06, WEB-01–02 | Keşfet üzerinden arama, edinim, toplu katkı ve popülerlik verisi denenebilir. |
| 5. Üretimi tamamla | API-07, WEB-03 | Tek HTTP isteğinde yeni görsel veya seçili kartın varyasyonu teslim edilir. |
| 6. Sohbeti ve sesi bağla | API-08–09, MCP-01–03, AGT-01–03, WEB-04–05 | Aynı görsel ihtiyacı yazıyla ve sesle anlatılabilir. Agent gerçek araçları çağırır. |
| 7. Teslimi doğrula | SON-01–03 | Arama kalitesi, maliyet, hata davranışı ve son demo için aynı sürüme ait sonuçlar elde edilir. |

Her görevi kendi bölümündeki sırayla uygula. Bir servisin kodunu bitirince çıktısını tüketen servis üzerinden de dene. Kutuyu işaretlerken ilgili K maddesine çalıştırılan komutu veya kullanıcı adımını, gözlenen sonucu ve gerekirse istek ya da görsel kimliğini yaz.

**Görevi kapatırken** etkilenen reponun AGENTS.md dosyasındaki kontrolleri uygula. Örneğin bir WebSocket kart alanı eklenirse server tarafında eski istemcinin mesajı okuyabildiğini, web tarafında da alanı içermeyen kayıtlı bir sohbetin açıldığını doğrula. İşlem sayısı veya hata davranışı değişiyorsa ilgili Prometheus sorgusunu ve Grafana panelini dene, Sentry etkin olduğunda hata bağlamını kontrol et. Kullanıcıya, diğer servise veya kurulumu yöneten kişiye yansıyan değişikliği o reponun yayımlanmamış kayıtlarına ekle. [Server changelog](../CHANGELOG.md) · [Web changelog](https://github.com/ekremrn/imavara-web/blob/main/CHANGELOG.md)

## Ortak altyapı ve sözleşmeler

Hedef dizinler `imavara-server/contracts`, `internal/platform`, `migrations` ve `deploy`. Bu işler beş servisin ortak bağımlılıklarını hazırlar.

- [ ] **ORT-01 · Repo ve yerel ağ.** Web ve server repolarını hazırla. Repo köklerindeki [web çalışma rehberini](https://github.com/ekremrn/imavara-web/blob/main/AGENTS.md) ve [server çalışma rehberini](../AGENTS.md) izle. Server’da tek Go modülü, agent için ayrı TypeScript paketi ve beş süreç için ayrı container girişleri oluştur. Compose’a PostgreSQL, Qdrant, RabbitMQ, Garage, Prometheus ve Grafana ekle. Başlatma komutlarını, ortam değişkeni örneğini, sağlık kontrollerini ve kalıcı diskleri tanımla. Kullanıcı trafiğini web/API üzerinden geçir. Sentry DSN ve OpenRouter anahtarını sunucu yapılandırmasında tut. Yerel başlangıcın boş DSN ile çalıştığını dene.
- [ ] **ORT-02 · Sözleşmeler.** API kartı, hata yanıtı, MCP araç girdileri/çıktıları, WebSocket olayları ve iki kuyruk mesajının alanlarını sürümle. İstek, ücretli işlem, sohbet mesajı ve görsel kimliklerinin hangilerinin tekrar denemede korunacağını belirt. Kuyrukta `asset_id`, nesne anahtarı ve sürüm taşı. Web ile agent’ın aynı örnek kartı ve olay sırasını okuyabildiğini doğrula.
- [ ] **ORT-03 · Veri servisleri.** PostgreSQL migration’larını ve tekillik kurallarını hazırla. Garage’da orijinal, önizleme ve özel analiz erişimlerini ayır. RabbitMQ’da analiz, indeks ve hata kuyruklarını kur. Qdrant ana profilinde 16 shard, tek node üzerinde replika 1, bir point üzerinde Cosine `text[1024]` ve `image[768]`, binary quantization ve yeniden puanlama tanımla. Mimari planın 11. bölümündeki payload indekslerini ilk dosyadan önce oluştur. Başlatmayı ikinci kez çalıştırıp mevcut verinin korunduğunu dene.
- [ ] **ORT-04 · Ortak istemciler ve ölçüm.** Go süreçlerine ortak S3, OpenRouter ve kuyruk erişimi, yapılandırma, yapılandırılmış log ve istek kimliği aktarımı ekle. Her süreçten Prometheus metriği al. TypeScript agent’ta aynı kimliklerle log üret. Sorgu ve indekslemede kullanılacak embedding boyut doğrulaması ve normalizasyonunu ortak tanımla.
- [ ] **ORT-05 · Sağlayıcı denemeleri.** Mimari planın 12. bölümündeki sabit profillerle yedi model rotasını dene. Ember’dan bir araç çağrısı, Muse Image’dan yeni üretim ve referanslı düzenleme, Luna’dan karar, Glimmer’dan geçerli JSON, Qwen ve Gemini’den doğru boyutta vektör ve Transcribe’dan Türkçe metin al. Sonuç, maliyet ve varsa uyumsuzluğu model kimliğiyle kaydet. **K1, K2, K11**

## Go API

Hedef dizin `imavara-server/cmd/api`. Web’den HTTP/WebSocket isteklerini ve MCP’nin ilettiği ürün işlemlerini alır. Oturum, erişim, bakiye ve dosya tesliminde karar burada verilir.

- [ ] **API-01 · Tarayıcı oturumu.** Rastgele HttpOnly çerezi PostgreSQL oturumuna bağla. İlk oturum ile +100 kredi hareketini birlikte kaydet. `/v1/me`, tarife ve hareket geçmişini aç. Aynı oturumu ikinci sekmede, ayrı oturumu ikinci tarayıcıda doğrula. **K7, K8**
- [ ] **API-02 · Kredi işlemleri.** Rezervasyon, kesinleştirme, serbest bırakma ve tekil edinim işlemlerini kur. Ücretli istek tekrarında önceki sonuca dön. Son üç kredi için eşzamanlı iki rezervasyondan yalnız birinin geçtiğini dene. Yayım sırasında kullanılacak tekil katkı ödülünü aynı kredi modülünde tanımla. **K7**
- [ ] **API-03 · Dosya ve haklar.** Dosyanın gerçek biçimini, boyutlarını ve SHA-256 değerini oku. Asset, önizleme ve orijinal anahtarlarını kaydet. Detay, kitaplık, edinim, kalıcı orijinal yolu, paylaşım ve kaldırma işlemlerinde tarayıcı hakkını kontrol et. İlk edinim −1, tekrar indirme 0 olmalı. Normal geri çekme sonrası önceki edinimle indirmeyi ve yeni edinimin kapanmasını dene. **K5, K7, K8**
- [ ] **API-04 · Katkı kabulü.** Dosya başına paylaşım iznini ve kullanım koşullarını al. Qdrant’ta içerikten türetilen point kimliğini ve tam hash’i karşılaştır. Eşzamanlı kabulü aynı içerik için tekilleştir. Yakın görsel adaylarını kesin kopyadan ayrı işle. Kabul adayını `openai/gpt-6-luna-decisions` ile değerlendir. Ret, inceleme, sağlayıcı hatası ve kabul sonucunu ayrı döndür. Kabul ve analiz işini birlikte kalıcılaştır, RabbitMQ yayını kesilirse aynı işi sürdürebil. **K9, K10**
- [ ] **API-05 · Arama.** Metinde Qwen’i, görsel referansta Gemini’yi veya erişilebilir kayıtlı vektörü kullan. Doğru named vector, görünürlük, lisans, kesin kullanıcı filtreleri ve alana özgü skoru uygula. Varsayılan beş kart ve 1–10 aralığında `top_k` desteği sun. Sonuçsuz ve başarılı aramada kredi defterini kontrol et. **K3, K11, K12**
- [ ] **API-06 · Galeri ve istatistik.** Yeni görselleri sayfalı listele. Arama başına dönen asset ve sıraları bir kez yaz. İlk edinimlerden ve arama olaylarından dönem bazında iki ayrı sıralama üret. İki ayrı arama, bir edinim ve iki indirmeyle API sonucunun iki dönüş ve bir edinim verdiğini doğrula. **K5, K13**
- [ ] **API-07 · Üretim ve varyasyon.** Açık talepte 3 kredi ayır. Erişim denetiminden geçen referansı, istenen değişikliği ve oranı `meta/muse-image` çağrısına ekle. Orijinali Garage’a yaz, gerçek ölçüyü ve özel asset’i kaydet, ücret ve teslimi sonuçlandır. Teslimle aynı işlemde katalog işini oluştur. Worker üretimi otomatik kabul değerlendirmesine alır. Kesin hata, geçici hata, kesilmiş yanıt ve aynı isteğin tekrarını dene. **K6, K7, K8**
- [ ] **API-08 · Sohbet.** Konuşma oluşturma, listeleme, açma, silme ve WebSocket uçlarını ekle. Tarayıcı bağlamını sunucuda çöz, agent’a mesaj ve erişilebilir referansları ilet. Mesajları, gerçek araç sonuçlarını ve kart sırasını kaydet. Yeniden bağlantıda eksik durumu tamamla. Aynı mesaj kimliği tek agent turu başlatmalı. **K3, K5, K8**
- [ ] **API-09 · Ses.** `POST /v1/audio/transcriptions` ile kaydın biçimini ve boyutunu doğrula. `openai/gpt-transcribe` yanıtından metni döndür. Ham sesi yalnız işlem süresince tut. Boş transkript ve sağlayıcı hatasını normal mesajdan ayır. **K2, K4**

## Görsel analiz worker’ı

Hedef dizin `imavara-server/cmd/image-extract-worker`. Girdi RabbitMQ analiz işi ve Garage’daki orijinaldir. Çıktı doğrulanmış analiz dosyası ve indeks işidir.

- [ ] **EXT-01 · İşi al.** Mesaj sürümünü kontrol et. Aynı asset ve sürüm için tamamlanmış analiz varsa onu kullan. Gerekliyse görseli nesne anahtarından oku ve `meta/muse-glimmer-30b` modeline sabit low effort ile gönder.
- [ ] **EXT-02 · Analizi doğrula.** Mevcut JSON şemasını ve alanlar arası kuralları uygula. Okunamayan yazı, bilinmeyen nitelik ve boş alan yönü için örnekleri kontrol et. Geçerli sonucu asset, prompt, model ve şema sürümüyle özel Garage alanına yaz. [Glimmer analiz şeması](../contracts/glimmer-analysis.v1.schema.json)
- [ ] **EXT-03 · İndeks işini devret.** İndeks mesajını yayımla, broker onayından sonra analiz mesajını ACK et. Geçici hatayı sınırlı tekrar, sürekli geçersiz çıktıyı hata kuyruğuyla işle. Dosya yazımı ile ACK arasında worker’ı yeniden başlat. Doğru analiz sürümünün indekslemeye ulaşmasını doğrula. **K2, K10, K13**

## İndeks worker’ı

Hedef dizin `imavara-server/cmd/index-worker`. Girdi doğrulanmış analiz, orijinal dosya ve güncel paylaşım kaydıdır. Çıktı aranabilir katalog point’i, yayım durumu ve uygunsa dış katkı ödülüdür.

- [ ] **IDX-01 · Arama metni ve filtreler.** Açıklama, ana konu, görünür ilişkiler, kompozisyon ve etiketleri sabit sırada birleştir. Stil, renk ve boş alanı payload’a eşle. Genişlik, oran, dosya biçimi ve hash için dosya kaydını kullan. Belirsiz boolean alanını eksik bırak.
- [ ] **IDX-02 · İki vektör.** Metni `qwen/qwen3-embedding-8b`, görseli `google/gemini-embedding-2` ile işle. API ile aynı dönüşümü uygula. Yanlış uzunluk veya sonlu olmayan değer gelirse işi başarısız olarak işle.
- [ ] **IDX-03 · Yayım.** Dosyanın güncel paylaşım sürümünü kontrol et. Sabit point kimliğiyle iki vektörü ve filtreleri yaz. Paylaşım geri çekmeyle yarışırsa eski işin görseli yeniden görünür yapmasını önle. Yayım kaydı ve uygun dış katkı ödülünü tekilleştir. Kalıcı sonuçtan sonra ACK et.
- [ ] **IDX-04 · Kesinti deneyi.** Qdrant yazımı ile ödül kaydı arasında ve ödül ile ACK arasında worker’ı kesip işi yeniden teslim et. API üzerinden tek görünür asset, tek ödül ve iki arama türünde doğru sonuç görülmeli. İşlenirken geri çekilen görseli ikinci tarayıcıdan ara. **K7, K8, K9, K10, K11**

## Go MCP

Hedef dizin `imavara-server/cmd/mcp`. İç ağdaki agent’tan araç çağrısı alır. API sonucunu aynı tarayıcı ve işlem bağlamında agent’a döndürür.

- [ ] **MCP-01 · Araçları tanımla.** `get_balance`, `search_images`, `get_image`, `acquire_image`, `generate_image` ve `contribute_image` için girdi ve sonuç şemalarını kaydet. Agent bağlantısından araç listesini ve şemaları oku.
- [ ] **MCP-02 · API’ye bağla.** Her aracı karşılık gelen API işlemine yönlendir. Oturum bağlamını güvenilen çağrıdan al. Modelin araç argümanlarıyla tarayıcı seçmesini engelle. Referans asset kimliğini, filtreleri ve ücretli işlem anahtarını koru. Kart sırası, gerçekleşen ücret ve hata kodunu API’den taşı.
- [ ] **MCP-03 · Aynı işlemi iki yoldan dene.** Aynı oturumla API ve MCP’den filtreli arama yap, sıralamayı karşılaştır. Aynı anahtarla yapılan ücretli tekrarın tek harekete karşılık geldiğini, ikinci tarayıcının özel asset’ini araçla açamadığını ve MCP’nin yalnız iç ağda olduğunu doğrula. **K1, K3, K7, K8**

## TypeScript web agent

Hedef dizin `imavara-server/agent`. API’den konuşmayı ve referansları alır. LangChain üzerinden iç MCP araçlarını kullanır, olayları API’ye iletir.

- [ ] **AGT-01 · Model ve araçlar.** Mevcut Türkçe sistem prompt’unu yükle. MCP istemcisini ve `meta/muse-ember` profilini bağla. Temperature 0.3, low effort, otomatik araç seçimi ve streaming ayarlarını doğrula. [Agent sistem prompt’u](../agent/prompts/system.tr.txt)
- [ ] **AGT-02 · Araç kararları.** Belirsiz tarifte bir kısa soru sorulmasını, açık aramada ücretsiz aracı ve açık üretim talebinde ücretli aracı dene. “İkincisine benzer” ile “ikincideki fincanı kaldır” örneklerinde aynı kartın kimliğini ve iki farklı araç seçimini izle. Boş sonuçta üretim önerisi kullanıcı kararını beklemeli.
- [ ] **AGT-03 · Olay akışı.** Gerçek araç başlangıcı, kartlar, metin parçaları, teslim ve hata olaylarını mesaj kimliği ve sırasıyla API’ye aktar. İşlem tekrarı ve araç hatasında kullanıcıya sunulan ücret ile API kaydını karşılaştır. **K2, K3, K5, K6, K7**

## React web uygulaması

Hedef repo `imavara-web`. Tarayıcı yalnız ürün API’si ve sohbet WebSocket’iyle çalışır.

- [ ] **WEB-01 · Oturum ve gezinme.** Ortak HTTP istemcisiyle oturum, bakiye, tarife ve hata durumlarını bağla. Minimal açılış alanını, Meta Muse ifadesini, Keşfet’i ve kitaplığı hazırla.
- [ ] **WEB-02 · Keşfet ve katkı.** Metin/referans araması, filtreler, orijinal oranlı kart, sayfalama, son yüklenenler, popülerler ve edinim eylemini bağla. Toplu katkıda dosya başına sonucu ve güncel bakiyeyi göster. Reddedilen dosyayla birlikte kabul edilen bir dosyayı aynı seçimde dene.
- [ ] **WEB-03 · Üretim ve kitaplık.** Karttan varyasyon iste, yeni üretim yap, gerçek dosya ölçüsünü göster. Teslimi, otomatik katalog değerlendirmesini, geri çekmeyi ve daha önce edinilen dosyanın yeniden indirilmesini bağla.
- [ ] **WEB-04 · Sohbet.** İlk mesajla açılıştan sohbete geç. WebSocket olaylarını kart sırasıyla göster. Yeni sohbet, geçmiş ve yeniden bağlantıyı bağla. Sohbetin silinmesinden sonra edinilmiş dosyayı kitaplıkta doğrula.
- [ ] **WEB-05 · Sesle anlat.** Hem açılış hem sohbet girişine mikrofon ekle. Kayıt süresini, Gönder ve İptal eylemlerini göster. Gönder’de transkripsiyonu başlat, bekleme durumunu göster ve dönen metni taslak/referansla aynı konuşmaya otomatik olarak bir kez gönder. Çift dokunma, mikrofon izni reddi, boş kayıt ve ağ hatasında taslak/mesaj durumlarını sınayarak K4’ü kapat. **K3, K4, K5, K6, K8, K13**

## Servisleri birlikte doğrulama

- [ ] **SON-01 · Arama ve model ölçümü.** İzinli başlangıç setini katkı hattından geçir. Yaklaşık 100 etiketli Türkçe/İngilizce sorgu ve görsel referansları kalibrasyon ve kabul örneklerine ayır. İki arama türü için ayrı eşik seç. Ayrılmış sette ilk beşte isabet ve doğru boş dönüş oranını teslim hedefleriyle karşılaştır. Kesin arama karşılaştırmasını, p95 süreyi ve ilk 100 temsili analizde geçerli Glimmer çıktısı başına maliyeti kaydet. **K2, K12**
- [ ] **SON-02 · İşletim.** Beş uygulama servisi için Grafana panolarını aç. Bir yüklemenin istek kimliğini API’den worker’lara kadar izle. Model hatası, kuyruk yaşı, yayıma kadar süre, maliyet ve boş arama ölçümlerini işlemlerle değiştirip kontrol et. Qdrant snapshot’ını ayrı koleksiyona geri yükle ve bilinen bir görseli ara. **K1, K10, K11, K13**
- [ ] **SON-03 · Teslim demosu.** Teslimde iki reponun CHANGELOG kayıtlarını gerçekleşen değişikliklerle karşılaştır. Önemli sürümde tarih, sürüm ve gerekli geçiş adımlarını ekle. Belgelenmiş komutlarla temiz kurulum yap. İki tarayıcıyla sesli arama, edinim, tekrar indirme, varyasyon, yeni üretim ve benzersiz dış katkı senaryosunu yürüt. Bakiye sırası 100 → 99 → 96 → 93 → 94 olmalı. İkinci tarayıcıda özel referans erişimi ile yayımlanmış üretim ve katkı aramasını dene. K1–K13 kanıtlarını iki repo sürümüyle teslim planına işle. **K1–K13**
