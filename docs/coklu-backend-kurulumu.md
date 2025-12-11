# Çoklu S3 Backend Desteği

Bu doküman Console'da çoklu S3 backend desteği özelliğinin nasıl yapılandırılacağını ve kullanılacağını açıklar.

## Genel Bakış

Console artık aynı anda birden fazla S3-uyumlu backend'i desteklemektedir:
- MinIO
- Garage (cluster desteği ile)
- Versity Gateway
- Genel S3-uyumlu servisler (AWS S3, Wasabi, DigitalOcean Spaces, vb.)

## Özellikler

### 1. Çoklu Backend Yönetimi
- Tek bir console'dan birden fazla S3 backend'i yapılandırma ve yönetme
- Backend'ler arasında sorunsuz geçiş
- Tüm backend'ler için sağlık izleme
- Garage backend için cluster desteği

### 2. Merkezi Kullanıcı Yönetimi
- **Admin Kullanıcılar**: Console ve backend'leri yönetir
- **S3 Kullanıcılar**: Birden fazla backend'de nesne depolamaya erişir
- Kullanıcı-backend eşleştirme
- Detaylı izinler

### 3. Backend Türleri

#### MinIO
Tam özellik desteğine sahip standart MinIO backend'i.

#### Garage
Cluster desteği olan hafif S3-uyumlu depolama sistemi.
- Cluster düğümleri arasında round-robin yük dengeleme
- Otomatik yük devretme
- Düğümler arasında okuma dağıtımı

#### Versity Gateway
Çeşitli depolama backend'leri için S3-uyumlu arayüz.

#### Genel S3
Herhangi bir S3-uyumlu servis için destek (AWS S3, Wasabi, vb.).

## Yapılandırma

### Ortam Değişkenleri

#### Eski Tek Backend (Geriye Dönük Uyumlu)
```bash
export CONSOLE_MINIO_SERVER=http://localhost:9000
export CONSOLE_MINIO_REGION=us-east-1
```

#### Çoklu Backend Yapılandırması
```bash
export CONSOLE_ENABLE_MULTI_BACKEND=true
export CONSOLE_BACKENDS='[
  {
    "id": "minio-birincil",
    "name": "Birincil MinIO",
    "type": "minio",
    "endpoint": "http://minio1.ornek.com:9000",
    "region": "us-east-1",
    "accessKey": "minioadmin",
    "secretKey": "minioadmin",
    "enabled": true
  },
  {
    "id": "garage-cluster",
    "name": "Garage Cluster",
    "type": "garage",
    "endpoint": "http://garage1.ornek.com:3900",
    "region": "garage",
    "accessKey": "GK...",
    "secretKey": "...",
    "clusterMode": true,
    "clusterEndpoints": [
      "http://garage1.ornek.com:3900",
      "http://garage2.ornek.com:3900",
      "http://garage3.ornek.com:3900"
    ],
    "enabled": true
  },
  {
    "id": "versity-backend",
    "name": "Versity Gateway",
    "type": "versity",
    "endpoint": "http://versity.ornek.com:8080",
    "region": "us-west-1",
    "accessKey": "...",
    "secretKey": "...",
    "enabled": true
  }
]'
export CONSOLE_DEFAULT_BACKEND=minio-birincil
```

#### Admin Kullanıcı Yapılandırması
```bash
export CONSOLE_ADMIN_USERNAME=admin
export CONSOLE_ADMIN_PASSWORD=guvenliSifre123
```

### Backend Yapılandırma Şeması

```json
{
  "id": "benzersiz-backend-id",        // Benzersiz tanımlayıcı
  "name": "Görünen İsim",              // Okunabilir isim
  "type": "minio|garage|versity|s3",   // Backend türü
  "endpoint": "http://host:port",      // Backend uç noktası
  "region": "us-east-1",               // Bölge
  "accessKey": "erisim-anahtari",      // Erişim anahtarı (IAM için opsiyonel)
  "secretKey": "gizli-anahtar",        // Gizli anahtar (IAM için opsiyonel)
  "secure": true,                      // HTTPS kullan (endpoint'ten otomatik algılanır)
  "clusterMode": false,                // Cluster modunu etkinleştir (Garage)
  "clusterEndpoints": [],              // Cluster uç noktaları (Garage)
  "enabled": true,                     // Backend'i etkinleştir/devre dışı bırak
  "metadata": {                        // Opsiyonel metadata
    "konum": "veri-merkezi-1"
  }
}
```

## Kullanıcı Yönetimi

### Admin Kullanıcılar

Admin kullanıcıların console seviyesinde izinleri vardır:
- `manage:backends` - Backend'leri ekleme, kaldırma, yapılandırma
- `manage:users` - Kullanıcı oluşturma, güncelleme, silme
- `manage:s3users` - Backend'ler arası S3 kullanıcılarını yönetme
- `view:backends` - Backend bilgilerini görüntüleme
- `view:users` - Kullanıcı bilgilerini görüntüleme
- `access:backend` - Belirli backend'lere erişim

### S3 Kullanıcılar

S3 kullanıcılarının nesne depolama erişimi vardır:
- Birden fazla backend'e eşlenebilir
- Her backend eşleştirmesi S3 kimlik bilgilerini içerir
- Backend başına politikalar ve gruplar

### Başlangıçta Admin Kullanıcı Oluşturma

Console, yapılandırıldığında ilk başlatmada otomatik olarak bir admin kullanıcı oluşturur:

```bash
export CONSOLE_ADMIN_USERNAME=admin
export CONSOLE_ADMIN_PASSWORD=GuvenliSifreniz
```

## API Uç Noktaları

### Backend Yönetimi

- `GET /api/v1/backends` - Tüm backend'leri listele
- `GET /api/v1/backends/{id}` - Backend detaylarını al
- `POST /api/v1/backends` - Yeni backend ekle
- `DELETE /api/v1/backends/{id}` - Backend'i kaldır
- `GET /api/v1/backends/health` - Backend sağlığını kontrol et
- `PUT /api/v1/backends/{id}/default` - Varsayılan olarak ayarla

### Kullanıcı Yönetimi

- `POST /api/v1/users/admin` - Admin kullanıcı oluştur
- `POST /api/v1/users/s3` - S3 kullanıcı oluştur
- `GET /api/v1/users` - Kullanıcıları listele
- `GET /api/v1/users/{id}` - Kullanıcı detaylarını al
- `PUT /api/v1/users/{id}` - Kullanıcıyı güncelle
- `DELETE /api/v1/users/{id}` - Kullanıcıyı sil

### S3 Kullanıcı Yapılandırması

- `POST /api/v1/users/{id}/s3-config` - Backend için S3 yapılandırması oluştur
- `GET /api/v1/users/{id}/s3-config` - S3 yapılandırmalarını listele
- `GET /api/v1/users/{id}/s3-config/{backendId}` - Belirli yapılandırmayı al
- `PUT /api/v1/users/{id}/s3-config/{backendId}` - Yapılandırmayı güncelle
- `DELETE /api/v1/users/{id}/s3-config/{backendId}` - Yapılandırmayı sil

## Kullanım Örnekleri

### 1. Birden Fazla Backend Yapılandırma

```bash
# Ortam değişkenlerini ayarla
export CONSOLE_BACKENDS='[...]'  # JSON yapılandırması
export CONSOLE_DEFAULT_BACKEND=minio-birincil

# Console'u başlat
./console server
```

### 2. API ile Admin Kullanıcı Oluşturma

```bash
curl -X POST http://localhost:9090/api/v1/users/admin \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@ornek.com",
    "password": "GuvenliSifre123!",
    "permissions": ["manage:backends", "manage:users"]
  }'
```

### 3. API ile Backend Ekleme

```bash
curl -X POST http://localhost:9090/api/v1/backends \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "Yeni MinIO Sunucusu",
    "type": "minio",
    "endpoint": "http://minio2.ornek.com:9000",
    "region": "us-east-1",
    "accessKey": "admin",
    "secretKey": "sifre"
  }'
```

### 4. Garage Cluster Yapılandırma

```json
{
  "id": "garage-prod",
  "name": "Üretim Garage Cluster",
  "type": "garage",
  "endpoint": "http://garage-lb.ornek.com:3900",
  "region": "garage",
  "accessKey": "GK...",
  "secretKey": "...",
  "clusterMode": true,
  "clusterEndpoints": [
    "http://garage1.ornek.com:3900",
    "http://garage2.ornek.com:3900",
    "http://garage3.ornek.com:3900"
  ],
  "enabled": true,
  "metadata": {
    "ortam": "uretim",
    "konum": "us-west"
  }
}
```

## Tek Backend'den Geçiş

Şu anda tek bir MinIO backend kullanıyorsanız, Console geriye dönük uyumluluğu korur:

1. `CONSOLE_MINIO_SERVER` ve `CONSOLE_MINIO_REGION` kullanmaya devam edin
2. Hazır olduğunuzda, çoklu backend'e geçiş yapın:
   - `CONSOLE_ENABLE_MULTI_BACKEND=true` ayarlayın
   - Mevcut backend'inizi `CONSOLE_BACKENDS` ile yapılandırın
   - Gerektiğinde yeni backend'ler ekleyin

## Güvenlik Konuları

1. **Kimlik Bilgileri Depolama**: Backend kimlik bilgileri bellekte saklanır ve API yanıtlarında asla gösterilmez
2. **Admin Kimlik Doğrulama**: Admin kullanıcılar için güçlü şifreler kullanın
3. **Backend Erişimi**: Backend erişimini kontrol etmek için detaylı izinler kullanın
4. **TLS**: Üretimde her zaman HTTPS uç noktaları kullanın
5. **Gizli Yönetimi**: Ortam değişkenlerini veya gizli yönetim sistemlerini kullanmayı düşünün

## Sorun Giderme

### Backend Bağlanmıyor
1. Endpoint URL formatını kontrol edin (http:// veya https://)
2. Ağ bağlantısını doğrulayın
3. Kimlik bilgilerini doğrulayın
4. Backend sağlığını kontrol edin: `GET /api/v1/backends/health`

### Cluster Modu Sorunları (Garage)
1. Tüm cluster uç noktalarının erişilebilir olduğundan emin olun
2. Cluster düğümlerinin düzgün yapılandırıldığını kontrol edin
3. Cluster genelinde tutarlı kimlik bilgilerini doğrulayın
4. Bireysel düğümlerin sağlık durumunu izleyin

### Kullanıcı Kimlik Doğrulaması Başarısız
1. Kullanıcının var olduğunu ve aktif olduğunu doğrulayın
2. Kullanıcı türünü kontrol edin (admin vs S3)
3. Backend erişim izinlerini kontrol edin
4. Hedef backend için S3 kullanıcı yapılandırmasının var olduğunu doğrulayın

## Performans Konuları

### Garage Cluster Modu
- Yazma işlemleri için round-robin yük dengeleme kullanır
- Okuma işlemleri için rastgele seçim
- Sağlık kontrolleri periyodik olarak çalışır
- Başarısız düğümler otomatik olarak atlanır

### Bağlantı Havuzlama
- Her backend kendi bağlantı havuzunu tutar
- Bağlantılar istekler arasında yeniden kullanılır
- Boşta kalan bağlantılar otomatik olarak temizlenir

## Gelecek Geliştirmeler

- [ ] Kullanıcı verileri için kalıcı depolama (veritabanı)
- [ ] Backend'ler arası kullanıcı senkronizasyonu
- [ ] Backend yük devretme ve yedeklilik
- [ ] Gelişmiş backend yönlendirme (bölge tabanlı, performans tabanlı)
- [ ] Admin işlemleri için denetim günlüğü
- [ ] Backend ve kullanıcı yönetimi için UI
