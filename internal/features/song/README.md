# Feature: Song

Cette feature gère tout ce qui concerne les **songs** et **lyrics** dans Lyrify.

## 📋 Responsabilités

- Matching intelligent songs ↔ lyrics
- CRUD operations sur songs et lyrics
- Gestion versions lyrics
- Tracking requêtes (analytics)
- Gestion pending requests (PR list)

## 🏗️ Architecture

```
song/
├── handlers.go      # HTTP handlers (routes API)
├── service.go       # Business logic (matching, validation)
├── repository.go    # Data access (queries GORM)
├── types.go         # DTOs (Request/Response)
└── README.md        # Ce fichier
```

## 📡 API Endpoints

### POST /api/v1/match
**Match song avec lyrics**

**Request:**
```json
{
  "title": "Moment gâché",
  "artist": "Kalash ft Satori",
  "duration": 345,
  "hash": "abc123..."  // Optional SHA256
}
```

**Response 200 (Found):**
```json
{
  "found": true,
  "song_id": "uuid",
  "lyrics_id": "uuid",
  "version": 1,
  "offset": 0,
  "lrc": "[00:12.50]Premier vers...",
  "song": { ... },
  "lyrics": { ... }
}
```

**Response 404 (Not Found):**
```json
{
  "found": false
}
```

**Side Effects:**
- Log dans `request_history` (analytics)
- Créer/update `pending_request` si not found

---

### GET /api/v1/songs/:id
**Récupérer un song par ID**

**Response 200:**
```json
{
  "id": "uuid",
  "title": "Moment gâché",
  "duration": 345,
  "artists": [...],
  "lyrics": [...]
}
```

---

### GET /api/v1/lyrics/:id
**Récupérer lyrics par ID**

**Response 200:**
```json
{
  "id": "uuid",
  "song_id": "uuid",
  "lrc_content": "[00:12.50]...",
  "duration": 345,
  "version": 1,
  "song": { ... }
}
```

---

### GET /api/v1/songs?title=...&artist=...&limit=20
**Rechercher songs**

**Query Params:**
- `title` (optional) - Titre partiel
- `artist` (optional) - Artiste partiel
- `limit` (optional, default 20, max 100) - Nombre résultats

**Response 200:**
```json
{
  "songs": [...],
  "total": 42
}
```

---

### GET /api/v1/pending-requests?limit=50
**Récupérer pending requests (PR list)**

**Usage:** Admin/Contributors pour voir quelles chansons créer

**Response 200:**
```json
{
  "requests": [
    {
      "id": "uuid",
      "title": "Unknown Song",
      "artist": "Artist",
      "duration": 200,
      "request_count": 15,
      "priority": 15,
      "status": "pending",
      "first_requested_at": "2026-02-27T10:00:00Z",
      "last_requested_at": "2026-02-27T15:30:00Z"
    }
  ],
  "total": 10
}
```

## 🔄 Flow du Matching

### Stratégie 1: Hash Match (le plus rapide)
```
1. User envoie hash SHA256 du fichier
2. Query: SELECT * FROM songs WHERE file_hash = ?
3. Si trouvé → Return avec lyrics
4. Sinon → Stratégie 2
```

### Stratégie 2: Normalized Match (standard)
```
1. Normaliser title et artist
   "Kalash Ft. Satori" → "kalash"
2. Query avec tolérance duration (±2 sec)
3. Si trouvé → Return avec lyrics
4. Sinon → Stratégie 3
```

### Stratégie 3: Not Found
```
1. Log dans request_history (status 404)
2. Créer/update pending_request
   - Si existe: increment request_count
   - Si nouveau: créer entrée
3. Return { found: false }
```

## 🧩 Service Layer

### NewService(repo *Repository)
Crée une instance du service.

### Match(req MatchRequest, deviceID, ipAddress) (*MatchResponse, error)
Match intelligent avec 3 stratégies.

**Side Effects:**
- Log `request_history`
- Create/update `pending_request` si not found

### GetSongByID(id string) (*models.Song, error)
Récupère song avec relations (artists, lyrics).

### GetLyricsByID(id string) (*models.Lyrics, error)
Récupère lyrics avec song associé.

### Search(req SearchRequest) (*SearchResponse, error)
Search songs par title/artist partial.

### GetTopPendingRequests(limit int) (*PendingRequestsResponse, error)
Récupère top pending requests (triés par priority DESC).

## 🗄️ Repository Layer

### FindByHash(hash string) (*models.Song, error)
Trouve song par SHA256 hash.

### FindByNormalizedMatch(title, artist string, duration int) (*models.Song, error)
Trouve song par normalized fields + duration (±2 sec).

### FindByID(id string) (*models.Song, error)
Trouve song par ID avec preload relations.

### FindLyricsByID(id string) (*models.Lyrics, error)
Trouve lyrics par ID avec preload song.

### Search(title, artist string, limit int) ([]models.Song, error)
Search songs avec LIKE.

### CreateRequestHistory(history *RequestHistory) error
Log une requête dans history.

### CreatePendingRequest(title, artist string, duration int) error
Crée ou update pending request.

## 📊 Analytics

Toutes les requêtes match sont loggées dans `request_history`:

**Queries utiles:**
```sql
-- Top songs demandées (30 derniers jours)
SELECT search_title, search_artist, COUNT(*) as requests
FROM request_history
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY search_title, search_artist
ORDER BY requests DESC
LIMIT 20;

-- Taux de succès matching
SELECT 
  COUNT(CASE WHEN found = true THEN 1 END)::float / COUNT(*) * 100 as success_rate
FROM request_history
WHERE created_at > NOW() - INTERVAL '7 days';

-- Devices les plus actifs
SELECT device_id, COUNT(*) as requests
FROM request_history
WHERE device_id IS NOT NULL
GROUP BY device_id
ORDER BY requests DESC
LIMIT 10;
```

## 🔔 Integration avec Feature Notification

Quand la feature notification sera implémentée:

```go
// service.go - après création pending request
if err := s.repo.CreatePendingRequest(...); err == nil {
    // Notifier que la demande a été enregistrée
    s.notificationService.NotifyPendingRequestCreated(...)
}

// Quand lyrics sont créés pour un pending request
if pending.Status == models.StatusCompleted {
    // Notifier tous les devices qui ont demandé cette song
    s.notificationService.NotifySongAvailable(song.ID)
}
```

## 🧪 Tests

TODO: Créer tests unitaires et d'intégration

### Tests à implémenter:
- [x] Repository: FindByHash, FindByNormalizedMatch
- [x] Service: Match (success, not found)
- [x] Service: Pending request creation
- [x] Handlers: POST /match validation
- [x] Integration: Full flow match → DB → response

## 🚀 Utilisation

### Dans main.go:
```go
import "github.com/Tiavina22/lyrify-backend/internal/features/song"

// Setup
db := services.NewDatabase(config.DatabaseURL)
songHandler := song.NewHandler(db)

// Register routes
api := router.Group("/api/v1")
songHandler.RegisterRoutes(api)
```

### Headers recommandés:
```
X-Device-ID: unique-device-identifier
Content-Type: application/json
```

## 📝 TODO

- [ ] Tests unitaires service
- [ ] Tests intégration handlers
- [ ] Rate limiting par device_id
- [ ] Validation input plus stricte
- [ ] Cache Redis pour matching fréquents
- [ ] Fuzzy matching (Levenshtein)
- [ ] Endpoints CRUD songs/lyrics (admin)
- [ ] Upvote system lyrics

---

**Maintainer:** Lyrify Team  
**Last Updated:** 27 février 2026
