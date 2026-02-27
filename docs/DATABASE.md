# Lyrify Backend - Database Schema Documentation

## Vue d'ensemble

Backend Go pour Lyrify avec architecture de matching intelligent de lyrics synchronisées (LRC).

## Structure de la base de données

### Tables principales

#### 1. `artists`
Stocke les informations des artistes musicaux.

| Colonne | Type | Contraintes | Description |
|---------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Identifiant unique |
| name | VARCHAR(255) | NOT NULL | Nom de l'artiste |
| normalized_name | VARCHAR(255) | NOT NULL, INDEX | Nom normalisé pour recherche |
| country | VARCHAR(100) | | Pays d'origine (optionnel) |
| biography | TEXT | | Biographie (optionnel) |
| created_at | TIMESTAMP | | Date de création |
| updated_at | TIMESTAMP | | Date de mise à jour |

**Relations:**
- Many-to-Many avec `songs` via table `song_artists`

**Indexes:**
- `idx_artist_normalized_name` sur `normalized_name`

---

#### 2. `songs`
Stocke les métadonnées des chansons (sans les paroles).

| Colonne | Type | Contraintes | Description |
|---------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Identifiant unique |
| title | VARCHAR(500) | NOT NULL | Titre de la chanson |
| normalized_title | VARCHAR(500) | NOT NULL, INDEX | Titre normalisé pour matching |
| normalized_artist | VARCHAR(500) | NOT NULL, INDEX | Artiste normalisé pour matching |
| duration | INTEGER | NOT NULL, INDEX | Durée en secondes |
| album | VARCHAR(500) | | Nom de l'album (optionnel) |
| year | INTEGER | | Année de sortie (optionnel) |
| file_hash | VARCHAR(64) | INDEX | Hash SHA256 du fichier (optionnel) |
| created_at | TIMESTAMP | | Date de création |
| updated_at | TIMESTAMP | | Date de mise à jour |

**Relations:**
- Many-to-Many avec `artists` via table `song_artists`
- One-to-Many avec `lyrics` (une song peut avoir plusieurs versions de lyrics)

**Indexes:**
- `idx_song_search` (composite) sur `(normalized_title, normalized_artist, duration)`
- `idx_song_hash` sur `file_hash`

**Logique métier:**
- `normalized_title` et `normalized_artist` sont auto-générés via hooks GORM
- Permettent le matching performant même avec variations d'écriture

---

#### 3. `lyrics`
Stocke les paroles synchronisées au format LRC.

| Colonne | Type | Contraintes | Description |
|---------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Identifiant unique |
| song_id | UUID | NOT NULL, FK, INDEX | Référence vers `songs` |
| lrc_content | TEXT | NOT NULL | Contenu LRC complet |
| lrc_timestamps | JSONB | | Timestamps extraits en JSON |
| duration | INTEGER | NOT NULL | Durée en secondes (doit matcher song) |
| version | INTEGER | NOT NULL, DEFAULT 1 | Version des lyrics |
| upvotes | INTEGER | DEFAULT 0 | Votes de la communauté |
| language | VARCHAR(10) | | Code langue (ex: "en", "fr") |
| offset | INTEGER | DEFAULT 0 | Décalage en millisecondes |
| source | VARCHAR(100) | | Source (ex: "community", "official") |
| created_at | TIMESTAMP | | Date de création |
| updated_at | TIMESTAMP | | Date de mise à jour |

**Relations:**
- Many-to-One avec `songs` (CASCADE DELETE)

**Indexes:**
- `idx_lyrics_song_version` (UNIQUE composite) sur `(song_id, version)`

**Contraintes importantes:**
- **Duration Match:** `lyrics.duration` doit être égal à `song.duration` (±2 sec tolérance)
- **Version unique:** Une song ne peut avoir qu'un seul lyrics par numéro de version
- **Format LRC:** Validé via `utils.ValidateLRCFormat()` avant insertion

**Exemple LRC:**
```lrc
[ar:Kalash]
[ti:Moment gâché]
[al:Kaos]
[00:12.50]Premier vers de la chanson
[00:18.30]Deuxième vers
```

---

#### 4. `request_history`
Historique de toutes les requêtes de recherche de lyrics.

| Colonne | Type | Contraintes | Description |
|---------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Identifiant unique |
| device_id | VARCHAR(255) | INDEX | Identifiant du device (optionnel) |
| user_id | VARCHAR(255) | INDEX | Identifiant user (optionnel, futur) |
| search_title | VARCHAR(500) | NOT NULL | Titre recherché |
| search_artist | VARCHAR(500) | NOT NULL | Artiste recherché |
| search_duration | INTEGER | NOT NULL | Durée recherchée (secondes) |
| status_code | INTEGER | NOT NULL, INDEX | Code HTTP (200, 404, etc.) |
| found | BOOLEAN | NOT NULL, DEFAULT false | Lyrics trouvés ou non |
| song_id | UUID | FK, INDEX, NULLABLE | ID de la song si trouvée |
| ip_address | VARCHAR(45) | | Adresse IP (IPv4/IPv6) |
| user_agent | TEXT | | User agent du client |
| created_at | TIMESTAMP | INDEX | Date de la requête |

**Relations:**
- Many-to-One avec `songs` (optionnel)

**Indexes:**
- `idx_request_device` sur `device_id`
- `idx_request_user` sur `user_id`
- `idx_request_status` sur `status_code`
- `idx_request_song` sur `song_id`
- `idx_request_created` sur `created_at`

**Cas d'usage:**
- Analytics: chansons les plus demandées
- Identifier les songs qui nécessitent des lyrics
- Tracking usage par device/user
- Taux de succès du matching

---

#### 5. `pending_requests`
Songs demandées mais sans lyrics disponibles (PR list).

| Colonne | Type | Contraintes | Description |
|---------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Identifiant unique |
| title | VARCHAR(500) | NOT NULL, UNIQUE* | Titre de la chanson |
| artist | VARCHAR(500) | NOT NULL, UNIQUE* | Artiste |
| album | VARCHAR(500) | | Album (optionnel) |
| duration | INTEGER | NOT NULL, UNIQUE* | Durée en secondes |
| file_hash | VARCHAR(64) | | Hash SHA256 (optionnel) |
| request_count | INTEGER | DEFAULT 1, INDEX | Nombre de demandes |
| priority | INTEGER | DEFAULT 0, INDEX | Priorité calculée |
| status | VARCHAR(20) | DEFAULT 'pending', INDEX | État: pending/in_progress/completed/rejected |
| metadata | JSONB | | Métadonnées supplémentaires |
| first_requested_at | TIMESTAMP | | Première demande |
| last_requested_at | TIMESTAMP | | Dernière demande |
| created_at | TIMESTAMP | | Date de création |
| updated_at | TIMESTAMP | | Date de mise à jour |

**Contraintes:**
- `idx_pending_unique` (UNIQUE composite) sur `(title, artist, duration)`
  - Empêche les doublons: même chanson avec même durée = un seul pending_request

**Indexes:**
- `idx_pending_count` sur `request_count`
- `idx_pending_priority` sur `priority`
- `idx_pending_status` sur `status`

**Workflow:**
1. User demande lyrics pour "Kalash - Moment gâché (5:45)"
2. Backend ne trouve pas → crée `pending_request` (ou incrémente `request_count` si existe)
3. `priority` calculé automatiquement (`request_count` = base)
4. Admin/Community voit les pending_requests triés par priority
5. Quand lyrics créés → status = "completed", optionnel: delete pending_request

**Cas d'usage:**
- Prioriser création de lyrics (plus demandé = plus urgent)
- Éviter doublons de demandes
- Tableau de bord pour contributors

---

#### 6. `song_artists` (table de liaison)
Table many-to-many entre `songs` et `artists`.

| Colonne | Type | Contraintes |
|---------|------|-------------|
| song_id | UUID | FK vers songs |
| artist_id | UUID | FK vers artists |

**Exemple:**
```
Song: "Moment gâché"
Artists: ["Kalash", "Satori"]
→ 2 entrées dans song_artists
```

---

## Normalisation des données

### Fonction `NormalizeString()`
Appliquée automatiquement via GORM hooks sur:
- `artists.normalized_name`
- `songs.normalized_title`
- `songs.normalized_artist`

**Transformations:**
1. Lowercase
2. Suppression accents/diacritics ("é" → "e")
3. Suppression feat/ft/featuring
4. Suppression caractères spéciaux
5. Multiple espaces → simple espace
6. Trim

**Exemples:**
```
"Kalash Ft. Satori" → "kalash satori"
"Damso (feat. Siboy)" → "damso"
"Beyoncé" → "beyonce"
```

---

## Matching Algorithm

### Stratégies (ordre d'exécution):

#### 1. Exact Hash Match (le plus rapide)
```sql
SELECT * FROM songs WHERE file_hash = ?
```

#### 2. Normalized Match (principal)
```sql
SELECT * FROM songs 
WHERE normalized_title = ? 
  AND normalized_artist = ? 
  AND duration BETWEEN (duration - 2) AND (duration + 2)
```

#### 3. Fuzzy Match (futur)
Utilise Levenshtein distance pour tolérer fautes de frappe.

---

## Tolérance de durée

**Règle:** ±2 secondes

**Pourquoi?**
- Versions radio/album peuvent différer légèrement
- Metadata parfois imprécises
- Trailing silence variable

**Exemple:**
```
Song: "Damso - Macarena" - 215 sec
Match accepté si lyrics.duration entre 213 et 217 sec
```

**Important:**
- Song 1: "Moment gâché" - 5m45 (345 sec)
- Song 2: "Moment gâché" - 5m58 (358 sec)
- Différence = 13 sec → **LYRICS DIFFÉRENTS**

---

## Contraintes d'intégrité

### 1. Duration mismatch protection
```go
// Dans Lyrics.BeforeSave()
if !utils.CompareDurations(lyrics.Duration, song.Duration, 2) {
    return ErrDurationMismatch
}
```

### 2. LRC Format validation
```go
// Format requis: [mm:ss.xx] text
if err := utils.ValidateLRCFormat(content); err != nil {
    return ErrInvalidLRCFormat
}
```

### 3. Unique version per song
```sql
UNIQUE INDEX idx_lyrics_song_version ON lyrics(song_id, version)
```

### 4. No duplicate pending requests
```sql
UNIQUE INDEX idx_pending_unique ON pending_requests(title, artist, duration)
```

---

## Migrations

Les migrations sont gérées par GORM AutoMigrate au démarrage de l'application.

**Commande:**
```bash
go run cmd/server/main.go
```

**Output attendu:**
```
✓ Database connected successfully
✓ Connection pool configured (MaxOpen: 25, MaxIdle: 5)
Running database migrations...
✓ Database migrations completed successfully
Adding custom indexes...
✓ Added unique index: idx_lyrics_song_version
✓ Custom indexes added
```

---

## Configuration

### Connection Pool
```go
MaxOpenConns:    25      // Connexions simultanées max
MaxIdleConns:    5       // Connexions idle maintenues
ConnMaxLifetime: 5min    // Durée de vie d'une connexion
ConnMaxIdleTime: 10min   // Temps avant fermeture idle conn
```

### Retry Logic
- Max retries: 5
- Backoff: exponentiel (2s, 4s, 6s, 8s, 10s)

---

## Queries exemples

### Rechercher une song
```go
matchService.FindBestMatch(&MatchRequest{
    Title: "Moment gâché",
    Artist: "Kalash ft Satori",
    Duration: 345,
    Hash: "abc123...",
})
```

### Top pending requests
```sql
SELECT * FROM pending_requests 
WHERE status = 'pending' 
ORDER BY priority DESC, request_count DESC 
LIMIT 50;
```

### Analytics: songs les plus demandées
```sql
SELECT search_title, search_artist, COUNT(*) as total_requests
FROM request_history
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY search_title, search_artist
ORDER BY total_requests DESC
LIMIT 20;
```

### Taux de succès matching
```sql
SELECT 
    COUNT(CASE WHEN found = true THEN 1 END)::float / COUNT(*) * 100 as success_rate
FROM request_history
WHERE created_at > NOW() - INTERVAL '7 days';
```

---

## Améliorations futures

1. **Full-text search** avec PostgreSQL `tsvector`
2. **Cache layer** avec Redis pour matching fréquent
3. **Fuzzy matching** avec Levenshtein distance
4. **Background jobs** pour cleanup automatique
5. **Partitioning** sur `request_history` par date
6. **Réplication read replicas** pour scalabilité
7. **Compression** LRC content avec gzip
8. **External storage** (S3) pour LRC volumineux

---

## Sécurité

### Implemented
- Input validation (title, artist, duration)
- LRC format validation
- File hash validation (SHA256)
- SQL injection protection (GORM parameterized queries)
- Connection pooling limits

### TODO
- Rate limiting par device_id/IP
- API authentication (JWT)
- DMCA takedown process
- Content sanitization
- Row-level security (RLS)

---

## Maintenance

### Cleanup jobs (à implémenter)
1. **request_history**: supprimer entrées > 90 jours
2. **pending_requests**: supprimer status=completed > 30 jours
3. **pending_requests**: supprimer status=pending + request_count=1 + age>180 jours

### Monitoring
- Query performance (EXPLAIN ANALYZE sur matching)
- Index usage stats
- Connection pool metrics
- Request success rate

---

## Contact
Pour questions ou contributions: tiaramilison@gmail.com
