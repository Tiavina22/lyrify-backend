# Lyrify Backend

Backend API en Go pour Lyrify - Système de matching intelligent de paroles synchronisées (LRC) pour applications desktop/mobile.

## Objectif

Fournir une API performante pour :
- Matcher des chansons avec leurs paroles synchronisées (format LRC)
- Gérer des durées exactes (±2 sec tolérance) pour éviter les décalages
- Tracker les demandes et identifier les chansons populaires sans lyrics
- Supporter les collaborations d'artistes (features)
- Offrir une base pour contributions communautaires

## Architecture

### Stack Technique
- **Language:** Go 1.25+
- **Framework:** Gin (HTTP router)
- **ORM:** GORM v2
- **Database:** PostgreSQL 14+
- **Dépendances clés:**
  - `golang.org/x/text` - Normalisation de texte
  - `github.com/google/uuid` - Génération UUID
  - `github.com/joho/godotenv` - Configuration env

### Structure du Projet
```
lyrify-backend/
├── cmd/
│   └── server/
│       └── main.go              # Point d'entrée de l'application
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration (env vars)
│   ├── errors/
│   │   └── errors.go            # Erreurs custom + APIError
│   ├── features/
│   │   └── song/                # Feature songs (TODO: handlers/repo)
│   ├── models/
│   │   ├── artist.go            # Modèle Artist (GORM)
│   │   ├── song.go              # Modèle Song (GORM)
│   │   ├── lyrics.go            # Modèle Lyrics (GORM)
│   │   ├── request_history.go   # Modèle RequestHistory (GORM)
│   │   └── pending_request.go   # Modèle PendingRequest (GORM)
│   ├── services/
│   │   ├── database.go          # DB connection + migrations
│   │   ├── logger.go            # Middleware logging
│   │   └── matching.go          # Service de matching songs/lyrics
│   └── utils/
│       ├── normalize.go         # Normalisation strings/durées
│       ├── normalize_test.go    # Tests normalisation
│       ├── validation.go        # Validation LRC/metadata
│       └── validation_test.go   # Tests validation
├── go.mod
├── go.sum
├── .env.example                 # Template configuration
├── DATABASE.md                  # Documentation détaillée du schéma DB
└── README.md                    # Ce fichier
```

## Base de Données

Le système utilise 5 tables principales + 1 table de liaison:

### Tables
1. **artists** - Artistes musicaux
2. **songs** - Métadonnées des chansons (sans paroles)
3. **lyrics** - Paroles synchronisées (format LRC)
4. **request_history** - Historique des recherches (analytics)
5. **pending_requests** - Chansons demandées sans lyrics (PR list)
6. **song_artists** - Liaison many-to-many songs ↔ artists

Voir [DATABASE.md](DATABASE.md) pour la documentation complète du schéma.

### Relations Clés
- **Song ↔ Artist:** Many-to-Many (support features: "Kalash ft Satori")
- **Song → Lyrics:** One-to-Many (plusieurs versions possibles)
- **Lyrics.duration** doit matcher **Song.duration** (±2 sec)

## Installation

### Prérequis
- Go 1.25 ou supérieur
- PostgreSQL 14 ou supérieur
- Make (optionnel)

### Setup

1. **Cloner le repository**
```bash
git clone https://github.com/Tiavina22/lyrify-backend.git
cd lyrify-backend
```

2. **Installer les dépendances**
```bash
go mod download
```

3. **Créer la base de données PostgreSQL**
```bash
createdb lyrify
```

4. **Configuration**
```bash
cp .env.example .env
# Éditer .env avec vos paramètres
```

Exemple `.env`:
```env
PORT=8080
DATABASE_URL=postgres://postgres:password@localhost:5432/lyrify_dev?sslmode=disable
```

5. **Lancer l'application** (migrations automatiques au démarrage)
```bash
go run cmd/server/main.go
```

Output attendu:
```
✓ Database connected successfully
✓ Connection pool configured (MaxOpen: 25, MaxIdle: 5)
Running database migrations...
✓ Database migrations completed successfully
✓ Custom indexes added
Server is running on port 8080
```

## Tests

### Lancer tous les tests
```bash
go test ./...
```

### Tests avec couverture
```bash
go test ./... -cover
```

### Tests verbose (utilities)
```bash
go test ./internal/utils/... -v
```

### Build
```bash
go build -o bin/lyrify-backend cmd/server/main.go
```

## API Endpoints

### Actuellement Implémentés

#### Health Check
```http
GET /health
```
**Response:**
```json
{
  "status": "ok",
  "timestamp": "2026-02-27T10:30:00Z"
}
```

### À Implémenter (Prochaines Étapes)

#### Match Song/Lyrics
```http
POST /api/v1/match
Content-Type: application/json

{
  "title": "Moment gâché",
  "artist": "Kalash ft Satori",
  "duration": 345,
  "hash": "abc123..."  // SHA256 optionnel
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

## Configuration

### Variables d'Environnement

| Variable | Description | Défaut | Requis |
|----------|-------------|--------|--------|
| `PORT` | Port du serveur HTTP | `8080` | Non |
| `DATABASE_URL` | URL de connexion PostgreSQL | - | Oui |

### Connection Pool (Database)
Configuré dans `internal/services/database.go`:
- **MaxOpenConns:** 25 (connexions simultanées max)
- **MaxIdleConns:** 5 (connexions idle maintenues)
- **ConnMaxLifetime:** 5 minutes
- **ConnMaxIdleTime:** 10 minutes

## Modules Principaux

### 1. Normalisation (`internal/utils/normalize.go`)
Fonctions pour normaliser les strings et durées pour matching performant.

**Exemple:**
```go
import "github.com/Tiavina22/lyrify-backend/internal/utils"

normalized := utils.NormalizeString("Kalash Ft. Satori")
// Result: "kalash"

match := utils.CompareDurations(345, 347, 2)
// Result: true (dans tolérance ±2 sec)
```

### 2. Validation (`internal/utils/validation.go`)
Validation format LRC, métadonnées de chansons, hash SHA256.

**Exemple:**
```go
lrcContent := `[00:12.50]Premier vers
[00:18.30]Deuxième vers`

err := utils.ValidateLRCFormat(lrcContent)
// err == nil si valide

timestamps, _ := utils.ExtractLRCTimestamps(lrcContent)
// []Timestamp{{Time: 12500, Text: "Premier vers"}, ...}
```

### 3. Matching Service (`internal/services/matching.go`)
Service principal pour trouver la meilleure correspondance song/lyrics.

**Stratégies de matching (ordre):**
1. **Hash Match** - Exact hash SHA256 (si fourni)
2. **Normalized Match** - Title + Artist normalisés + Duration (±2s)
3. **Fuzzy Match** - À implémenter (Levenshtein distance)

## Exemples d'Usage

### Cas 1: Song avec Durée Exacte
```
User demande: "Kalash - Moment gâché" (5m45 = 345 sec)
Backend trouve: Song avec duration = 345 sec
→ Match avec lyrics.duration = 345 sec
Success
```

### Cas 2: Song avec Durée Légèrement Différente
```
User demande: "Damso - Macarena" (215 sec)
Backend trouve: Song avec duration = 217 sec
Tolérance: ±2 sec
→ Match accepté (217 - 215 = 2 sec ≤ tolérance)
Success
```

### Cas 3: Durée Trop Différente
```
User demande: "Moment gâché" (5m45 = 345 sec)
Backend a: Version 1 (345 sec) et Version 2 (5m58 = 358 sec)
→ Match avec Version 1 uniquement
→ Version 2 a des lyrics différents (13 sec de différence)
Lyrics corrects pour la bonne version
```

### Cas 4: Song Non Trouvée
```
User demande: "Artist - Unknown Song" (200 sec)
Backend ne trouve rien
→ Crée entrée dans pending_requests
→ Analytics: identifier chansons populaires manquantes
```

## Sécurité

### Implémenté
- Validation des inputs (title, artist, duration)
- Validation format LRC
- SQL injection prevention (GORM parameterized queries)
- Connection pooling avec limites

### À Implémenter ⏳
- Rate limiting par device_id/IP
- Authentication JWT
- CORS configuration
- Request size limits
- DMCA takedown process

## Roadmap

### Phase 1 (MVP) - En Cours
- [x] Modèles database (Artist, Song, Lyrics, etc.)
- [x] Utilities (normalisation, validation)
- [x] Service de matching
- [x] Migrations automatiques
- [ ] Handler POST /match
- [ ] Handler GET /lyrics/:id
- [ ] Tests d'intégration

### Phase 2 - Features Avancées
- [ ] Repository pattern (abstraction data access)
- [ ] Fuzzy matching (Levenshtein)
- [ ] Cache Redis pour matching fréquents
- [ ] Background jobs (cleanup, analytics)
- [ ] Rate limiting middleware
- [ ] API authentication

### Phase 3 - Communauté
- [ ] POST /lyrics (contribution communauté)
- [ ] Upvote system
- [ ] User accounts
- [ ] Ranking lyrics par quality
- [ ] Dashboard admin

### Phase 4 - Scale
- [ ] Full-text search (PostgreSQL tsvector)
- [ ] Read replicas
- [ ] Metrics & monitoring (Prometheus)
- [ ] API documentation (Swagger)
- [ ] Docker deployment

## Contribution

### Structure d'une PR
1. Fork le projet
2. Créer une branche feature (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push branche (`git push origin feature/amazing-feature`)
5. Ouvrir une Pull Request

### Standards
- Tests unitaires requis pour nouvelles fonctionnalités
- Code formaté avec `go fmt`
- Pas de warnings `go vet`
- Documentation inline pour fonctions publiques

## License

À définir (MIT/Apache 2.0 recommandé pour projet open-source)

## Contact

**Projet:** Lyrify Backend  
**Version:** 0.0.1 (MVP)  
**Maintainer:** [Tiavina22](https://github.com/Tiavina22)

---

**Note:** Ce backend est conçu pour être utilisé avec l'application Flutter desktop Lyrify. Voir [lyrify](https://github.com/Tiavina22/lyrify) pour le frontend.
