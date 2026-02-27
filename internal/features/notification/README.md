# Feature: Notification (À Implémenter)

Cette feature gérera les **push notifications** pour informer les users des événements importants.

## 🎯 Objectifs

Notifier les users quand:
- ✅ **Song disponible:** Lyrics sont créés pour une song qu'ils ont demandée
- ✅ **Pending request créé:** Leur demande a été ajoutée à la PR list
- ✅ **Lyrics mis à jour:** Nouvelle version disponible pour une song suivie
- ✅ **Popular song:** Une song devient tendance (inciter à contribuer)

## 📋 Responsabilités

- Device registration (store FCM/APNs tokens)
- Send push notifications (Firebase Cloud Messaging, Apple Push Notification)
- Manage notification preferences per device
- Queue notifications (background worker)
- Track notification delivery (logs)
- Handle notification errors (retry logic)

## 🏗️ Architecture (Proposée)

```
notification/
├── handlers.go       # HTTP handlers (device registration, preferences)
├── service.go        # Notification logic, templates
├── repository.go     # Device CRUD, logs
├── pusher.go         # Push service abstraction (FCM, APNs)
├── worker.go         # Background job processor
├── templates.go      # Notification message templates
├── types.go          # DTOs (DeviceRegisterRequest, SendNotificationRequest)
└── README.md         # Ce fichier
```

## 🔔 Types de Notifications

### 1. Song Available
**Trigger:** Quand lyrics sont créés pour une song en pending_request

**Template:**
```
Title: "Lyrics disponibles! 🎵"
Body: "Les lyrics pour '{song_title}' de {artist} sont maintenant disponibles"
Data: {
  "type": "song_available",
  "song_id": "uuid",
  "lyrics_id": "uuid"
}
```

**Flow:**
```
1. Admin/Contributor crée lyrics pour song X
2. Background worker détecte pending_request avec status=completed
3. Query tous les devices qui ont demandé cette song (via request_history)
4. Filter devices avec notify_on_song_available=true
5. Send notification à chaque device
6. Log résultat dans notification_logs
```

### 2. Pending Request Created
**Trigger:** Quand user fait une requête et song not found

**Template:**
```
Title: "Demande enregistrée ✓"
Body: "Votre demande pour '{song_title}' a été ajoutée à notre liste"
Data: {
  "type": "pending_request_created",
  "pending_request_id": "uuid"
}
```

### 3. Lyrics Updated
**Trigger:** Quand nouvelle version de lyrics est publiée

**Template:**
```
Title: "Mise à jour disponible 🔄"
Body: "Nouvelle version des lyrics pour '{song_title}'"
Data: {
  "type": "lyrics_updated",
  "song_id": "uuid",
  "lyrics_id": "uuid",
  "version": 2
}
```

### 4. Popular Song
**Trigger:** Quand song atteint un seuil de demandes sans lyrics

**Template:**
```
Title: "Chanson tendance 🔥"
Body: "'{song_title}' est très demandée. Contribuez ses lyrics!"
Data: {
  "type": "popular_song",
  "pending_request_id": "uuid",
  "request_count": 50
}
```

## 📡 API Endpoints (À Implémenter)

### POST /api/v1/devices
**Register device pour notifications**

**Request:**
```json
{
  "token": "fcm_token_or_apns_token",
  "platform": "ios",
  "device_name": "iPhone 14",
  "app_version": "1.0.0",
  "os_version": "iOS 17.2",
  "language": "fr",
  "timezone": "Europe/Paris"
}
```

**Response 201:**
```json
{
  "id": "uuid",
  "token": "...",
  "platform": "ios",
  "active": true,
  "notifications_enabled": true,
  "created_at": "2026-02-27T10:00:00Z"
}
```

---

### PUT /api/v1/devices/:id
**Update device preferences**

**Request:**
```json
{
  "notifications_enabled": true,
  "notify_on_song_available": true,
  "notify_on_pending_request": false,
  "notify_on_lyrics_updated": false,
  "notify_on_popular_song": false,
  "language": "en"
}
```

**Response 200:** Updated device object

---

### GET /api/v1/devices/:id
**Get device details**

**Response 200:**
```json
{
  "id": "uuid",
  "token": "...",
  "platform": "ios",
  "notifications_enabled": true,
  "notify_on_song_available": true,
  "last_seen_at": "2026-02-27T15:30:00Z"
}
```

---

### DELETE /api/v1/devices/:id
**Unregister device**

**Response 204:** No content

---

### POST /api/v1/notifications/send (Admin Only)
**Manually send notification**

**Request:**
```json
{
  "device_ids": ["uuid1", "uuid2"],
  "title": "Test Notification",
  "body": "This is a test",
  "data": {
    "type": "custom",
    "url": "https://..."
  }
}
```

**Response 200:**
```json
{
  "sent": 2,
  "failed": 0
}
```

## 🔄 Implementation Guide

### Step 1: Setup Push Services

#### A. Firebase Cloud Messaging (Android + iOS)
```bash
# 1. Créer projet Firebase
# 2. Télécharger service account key JSON
# 3. Installer SDK
go get firebase.google.com/go/v4
```

```go
// pusher.go
import "firebase.google.com/go/v4/messaging"

type FCMPusher struct {
    client *messaging.Client
}

func (p *FCMPusher) Send(token string, notification Notification) error {
    message := &messaging.Message{
        Token: token,
        Notification: &messaging.Notification{
            Title: notification.Title,
            Body:  notification.Body,
        },
        Data: notification.Data,
    }
    
    _, err := p.client.Send(context.Background(), message)
    return err
}
```

#### B. Apple Push Notification (iOS/macOS)
```bash
# 1. Créer APNs certificate
# 2. Export .p12 ou .p8 key
# 3. Installer SDK
go get github.com/sideshow/apns2
```

```go
// pusher.go
import "github.com/sideshow/apns2"

type APNsPusher struct {
    client *apns2.Client
}

func (p *APNsPusher) Send(token string, notification Notification) error {
    n := &apns2.Notification{
        DeviceToken: token,
        Topic:       "com.lyrify.app",
        Payload: &payload.Payload{
            Alert: payload.Alert{
                Title: notification.Title,
                Body:  notification.Body,
            },
            CustomData: notification.Data,
        },
    }
    
    res, err := p.client.Push(n)
    return err
}
```

### Step 2: Create Unified Pusher Interface

```go
// pusher.go
type Pusher interface {
    Send(token string, notification Notification) error
}

type Notification struct {
    Title string
    Body  string
    Data  map[string]string
}

// Factory
func NewPusher(platform models.DevicePlatform) Pusher {
    switch platform {
    case models.PlatformIOS, models.PlatformMacOS:
        return newAPNsPusher()
    case models.PlatformAndroid:
        return newFCMPusher()
    default:
        return newNoOpPusher() // For web/windows/linux
    }
}
```

### Step 3: Create Notification Service

```go
// service.go
package notification

type Service struct {
    repo        *Repository
    songService *song.Service
    pushers     map[models.DevicePlatform]Pusher
}

func (s *Service) NotifySongAvailable(songID string) error {
    // 1. Get song details
    song, err := s.songService.GetSongByID(songID)
    if err != nil {
        return err
    }
    
    // 2. Find devices that requested this song
    devices, err := s.repo.GetDevicesForSong(song.Title, song.Artist)
    if err != nil {
        return err
    }
    
    // 3. Filter devices with notify_on_song_available=true
    devices = filterDevices(devices, func(d models.Device) bool {
        return d.IsActive() && d.NotifyOnSongAvailable
    })
    
    // 4. Send notifications
    notification := Notification{
        Title: "Lyrics disponibles! 🎵",
        Body:  fmt.Sprintf("Les lyrics pour '%s' de %s sont maintenant disponibles", song.Title, song.Artist),
        Data: map[string]string{
            "type":       "song_available",
            "song_id":    song.ID,
            "lyrics_id":  song.Lyrics[0].ID,
        },
    }
    
    for _, device := range devices {
        if err := s.sendToDevice(device, notification); err != nil {
            log.Printf("Failed to send notification to device %s: %v", device.ID, err)
        }
    }
    
    return nil
}

func (s *Service) sendToDevice(device models.Device, notif Notification) error {
    pusher := s.pushers[device.Platform]
    
    err := pusher.Send(device.Token, notif)
    
    // Log notification
    s.repo.CreateNotificationLog(&models.NotificationLog{
        DeviceID:         device.ID,
        NotificationType: notif.Data["type"],
        Title:            notif.Title,
        Body:             notif.Body,
        Status:           getStatus(err),
        ErrorMessage:     getErrorMessage(err),
    })
    
    return err
}
```

### Step 4: Create Background Worker

```go
// worker.go
package notification

type Worker struct {
    service *Service
    ticker  *time.Ticker
}

func (w *Worker) Start() {
    w.ticker = time.NewTicker(1 * time.Minute)
    
    go func() {
        for range w.ticker.C {
            w.processCompletedPendingRequests()
        }
    }()
}

func (w *Worker) processCompletedPendingRequests() {
    // 1. Find pending requests marked as completed
    requests, _ := w.service.repo.GetCompletedPendingRequests()
    
    for _, req := range requests {
        // 2. Find associated song
        song, err := w.service.songService.FindBySongMetadata(req.Title, req.Artist, req.Duration)
        if err != nil {
            continue
        }
        
        // 3. Send notifications
        w.service.NotifySongAvailable(song.ID)
        
        // 4. Delete or archive pending request
        w.service.repo.DeletePendingRequest(req.ID)
    }
}
```

### Step 5: Integration avec Feature Song

Dans `features/song/service.go`, après création de pending request:

```go
// service.go (song feature)
func (s *Service) Match(req MatchRequest, deviceID, ipAddress string) (*MatchResponse, error) {
    // ... existing matching logic ...
    
    // If not found
    if err := s.repo.CreatePendingRequest(req.Title, req.Artist, req.Duration); err == nil {
        // Notify device that request was created
        if s.notificationService != nil && deviceID != "" {
            s.notificationService.NotifyPendingRequestCreated(deviceID, req.Title, req.Artist)
        }
    }
    
    return &MatchResponse{Found: false}, errors.ErrSongNotFound
}
```

## 📊 Database Tables

Cette feature utilise 2 nouvelles tables:

### devices
Voir [internal/models/device.go](../../models/device.go)

Colonnes clés:
- `token` (UNIQUE) - FCM/APNs token
- `platform` - ios, android, macos, etc.
- `notifications_enabled` - User preference
- `notify_on_song_available`, `notify_on_pending_request`, etc.

### notification_logs
Log de toutes les notifications envoyées (debugging, analytics)

## 🧪 Testing Strategy

### Unit Tests
```go
func TestService_NotifySongAvailable(t *testing.T) {
    // Mock pusher
    mockPusher := &MockPusher{}
    service := NewService(repo, mockPusher)
    
    // Test
    err := service.NotifySongAvailable("song_id")
    
    // Assert pusher was called
    assert.NoError(t, err)
    assert.Equal(t, 1, mockPusher.CallCount)
}
```

### Integration Tests
- Test device registration
- Test notification delivery (sandbox mode)
- Test error handling (invalid token)

## 🚀 Deployment Checklist

- [ ] Create Firebase project (FCM)
- [ ] Create APNs certificates
- [ ] Add environment variables (FCM_KEY, APNS_KEY)
- [ ] Setup background worker as separate service
- [ ] Implement retry logic for failed notifications
- [ ] Monitor notification_logs for errors
- [ ] Setup alerts for high failure rate

## 🔐 Security

- Store device tokens encrypted
- Validate device ownership before update
- Rate limit notification sending
- Sanitize notification content (no PII)
- Implement opt-out mechanism

## 📝 TODO (Implementation)

1. [ ] Create pusher.go with FCM + APNs integration
2. [ ] Create repository.go (device CRUD, logs)
3. [ ] Create service.go (notification logic)
4. [ ] Create templates.go (message templates per language)
5. [ ] Create handlers.go (device registration API)
6. [ ] Create worker.go (background job processor)
7. [ ] Add notification service to database migrations
8. [ ] Write tests
9. [ ] Document API with Swagger
10. [ ] Setup monitoring and alerts

## 📚 Resources

- [Firebase Cloud Messaging](https://firebase.google.com/docs/cloud-messaging)
- [Apple Push Notification Service](https://developer.apple.com/documentation/usernotifications)
- [FCM Go SDK](https://firebase.google.com/docs/cloud-messaging/admin/send-messages)
- [APNs Go Library](https://github.com/sideshow/apns2)

---

**Status:** 🚧 À implémenter (Phase 2)  
**Estimation:** 2-3 semaines  
**Priority:** Medium (après MVP Song feature)

**Note:** Cette feature peut être développée en parallèle du MVP car elle est découplée.
