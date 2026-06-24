# RULES — MongoDB Communication Patterns

**Scope:** Every `infrastructure/mongodb/*.go` file. Governs how Go code reads, writes, queries, indexes, and manages collections in MongoDB.
**Package:** `project-abyssoftime-cms-v2/api/internal/infrastructure/mongodb`

---

## 1. Client Connection

### 1.1 Initialization (`client.go`)
```go
func NewClient(ctx context.Context, uri string) (*mongo.Client, error)
func Database(client *mongo.Client, name string) *mongo.Database
```
- URI from `MONGODB_URI` env var — **NEVER** hardcode
- `ConnectTimeout`: 30s, `ServerSelectionTimeout`: 30s
- TLS: `MinVersion: tls.VersionTLS12` — always enabled
- Ping with 15s timeout immediately after connect — fail fast
- Return `(*mongo.Client, error)` — caller handles retry/fatal

### 1.2 Database Reference
- `Database(client, name)` returns `*mongo.Database`
- Database name from `MONGODB_DB` env var
- Each repository receives `*mongo.Database` via constructor — **NOT** `*mongo.Client`

---

## 2. Repository Struct Pattern

### 2.1 Static Collections (User, Role, ContentType, MediaAsset, Invite, AccessToken, Locale)
```go
type <entity>Repository struct {
    col *mongo.Collection      // fixed collection, set once in constructor
}

func New<Entity>Repository(db *mongo.Database) repository.XRepository {
    return &<entity>Repository{col: db.Collection("<collection_name>")}
}
```
- Collection reference stored as struct field — created once in constructor
- Collection names (exact): `"users"`, `"roles"`, `"content_types"`, `"media_assets"`, `"invites"`, `"access_tokens"`, `"locales"`

### 2.2 Dynamic Collections (Document)
```go
type documentRepository struct {
    db *mongo.Database          // database ref, NOT a single collection
}

func (r *documentRepository) collection(contentTypeSlug string) *mongo.Collection {
    return r.db.Collection("documents_" + contentTypeSlug)
}
```
- Document repo stores `*mongo.Database` (not `*mongo.Collection`)
- Collection resolved at call time: `"documents_" + contentTypeSlug`
- Slug used as-is (no underscore conversion — MongoDB allows hyphens)

---

## 3. Interface Compliance Assertion

Every repository file starts with a compile-time check:
```go
var _ repository.XRepository = (*xRepository)(nil)
```
- **ALWAYS** include this line — catches interface mismatches at compile time
- Place immediately after the struct declaration

---

## 4. BSON Filter Patterns

### 4.1 Single-Field Filters
```go
bson.M{"email": email}             // exact match
bson.M{"_id": id}                  // by internal ID
bson.M{"documentId": documentID}   // by domain ID
bson.M{"slug": slug}               // by slug
bson.M{"code": code}               // by locale code
bson.M{"isDefault": true}          // boolean match
```

### 4.2 Compound Filters (Document-Specific)
```go
// Version filter — the primary document lookup pattern
bson.M{"documentId": documentID, "version": version, "locale": locale}

// Locale filter for deleting both draft + published
bson.M{"documentId": documentID, "locale": locale}

// List filter
bson.M{"version": entity.VersionDraft, "locale": locale}

// Batch lookup
bson.M{
    "version":    entity.VersionPublished,
    "locale":     locale,
    "documentId": bson.M{"$in": documentIDs},
}
```

### 4.3 Helper Functions (Document Repo)
```go
func versionFilter(documentID string, version entity.DocumentVersion, locale string) bson.M {
    return bson.M{"documentId": documentID, "version": version, "locale": locale}
}

func documentLocaleFilter(documentID, locale string) bson.M {
    return bson.M{"documentId": documentID, "locale": locale}
}
```
- Extract repeated filters into named helpers
- Helpers are package-private (`lowercase`)

---

## 5. CRUD Operation Patterns

### 5.1 Create (InsertOne)
```go
func (r *xRepository) Create(ctx context.Context, entity *entity.X) error {
    if entity.CreatedAt.IsZero() {
        entity.CreatedAt = time.Now().UTC()
    }
    _, err := r.col.InsertOne(ctx, entity)
    if mongo.IsDuplicateKeyError(err) {
        return pkgerrors.ErrConflict
    }
    return err
}
```
- Set `CreatedAt` if zero — use `time.Now().UTC()` (**always** UTC)
- Check `mongo.IsDuplicateKeyError` → return `pkgerrors.ErrConflict`
- Pass entity struct directly — BSON tags handle serialization

### 5.2 FindOne
```go
func (r *xRepository) FindByX(ctx context.Context, value string) (*entity.X, error) {
    var result entity.X
    err := r.col.FindOne(ctx, bson.M{"field": value}).Decode(&result)
    if err == mongo.ErrNoDocuments {
        return nil, pkgerrors.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    return &result, nil
}
```
- **ALWAYS** check `mongo.ErrNoDocuments` → map to `pkgerrors.ErrNotFound`
- Use `==` comparison (not `errors.Is`) for `mongo.ErrNoDocuments` — this is the established pattern
- Separate nil-err check after ErrNoDocuments check

### 5.3 FindMany (Cursor)
```go
func (r *xRepository) FindAll(ctx context.Context) ([]*entity.X, error) {
    opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
    cursor, err := r.col.Find(ctx, bson.M{}, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)           // ALWAYS defer Close
    var results []*entity.X
    if err := cursor.All(ctx, &results); err != nil {
        return nil, err
    }
    return results, nil
}
```
- **ALWAYS** `defer cursor.Close(ctx)` immediately after `Find`
- Use `cursor.All()` for full result set
- Sort with `bson.D` (ordered) — **NOT** `bson.M` (unordered)
- Sort direction: `1` = ascending, `-1` = descending

### 5.4 FindMany Paginated
```go
opts := options.Find().
    SetSort(bson.D{{Key: sortKey, Value: sortDir}}).
    SetSkip(int64(start)).
    SetLimit(int64(size))
```
- Count total first: `col.CountDocuments(ctx, filter)`
- Then fetch page with `Skip` + `Limit`
- Return `([]*entity.X, int64, error)` — items, total, error

### 5.5 Update (ReplaceOne)
```go
func (r *xRepository) Update(ctx context.Context, entity *entity.X) error {
    entity.UpdatedAt = time.Now().UTC()
    res, err := r.col.ReplaceOne(ctx, bson.M{"<lookupField>": entity.LookupValue}, entity)
    if err != nil {
        return err
    }
    if res.MatchedCount == 0 {
        return pkgerrors.ErrNotFound
    }
    return nil
}
```
- Use `ReplaceOne` — full document replacement
- **NEVER** use `UpdateOne` with `$set` for entity updates (too fragile)
- Check `MatchedCount == 0` → return `pkgerrors.ErrNotFound`
- Set `UpdatedAt` before replace
- Lookup field varies by entity:

| Entity | Lookup Field | Lookup Value |
|--------|-------------|--------------|
| User | `_id` | `user.ID` |
| Role | `documentId` | `role.DocumentID` |
| ContentType | `_id` | `ct.DocumentID` |
| Locale | `code` | `locale.Code` |

### 5.6 Upsert (ReplaceOne + Upsert Option) — Documents Only
```go
_, err := r.collection(slug).ReplaceOne(
    ctx,
    versionFilter(doc.DocumentID, version, doc.Locale),
    doc,
    options.Replace().SetUpsert(true),
)
```
- Used only for document upsert (draft/published)
- Filter by `(documentId, version, locale)` triple
- `SetUpsert(true)` — creates if not found, replaces if found
- Generate `GormID` for new drafts by finding max existing + 1

### 5.7 Delete (DeleteOne / DeleteMany)
```go
// Single delete
func (r *xRepository) Delete(ctx context.Context, id string) error {
    res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
    if err != nil {
        return err
    }
    if res.DeletedCount == 0 {
        return pkgerrors.ErrNotFound
    }
    return nil
}

// Bulk delete (documents by locale)
_, err := r.collection(slug).DeleteMany(ctx, filter)

// Delete all in collection
_, err := r.collection(slug).DeleteMany(ctx, bson.M{})
```
- `DeleteOne`: check `DeletedCount == 0` → `pkgerrors.ErrNotFound`
- `DeleteMany`: ignore `DeletedCount` (0 is valid)

### 5.8 Batch Update (UpdateMany) — Locale ClearDefault
```go
_, err := r.col.UpdateMany(ctx, bson.M{"isDefault": true}, bson.M{"$set": bson.M{"isDefault": false}})
```
- Only used for `ClearDefault` in locale repository
- `$set` operator for partial field update across multiple documents

---

## 6. Index Management

### 6.1 Static Indexes (`indexes.go`)
```go
func EnsureIndexes(ctx context.Context, db *mongo.Database) error
```
- Called once at startup in `main.go`
- Indexes:
  - `users.email` — unique
  - `content_types.slug` — unique
  - `media_assets.documentRef` — non-unique

### 6.2 Dynamic Indexes (Document Collections)
```go
func (r *documentRepository) EnsureCollection(ctx context.Context, slug string, _ []FieldDefinition) error {
    col := r.collection(slug)
    _, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys:    bson.D{{Key: "documentId", Value: 1}, {Key: "version", Value: 1}, {Key: "locale", Value: 1}},
        Options: options.Index().SetUnique(true),
    })
    return err
}
```
- Creates unique compound index: `(documentId, version, locale)`
- Called on every startup via sync — idempotent (MongoDB ignores duplicate index creation)
- `FieldDefinition` parameter ignored — MongoDB is schemaless
- **NEVER** drop collections in MongoDB `EnsureCollection` — it's already non-destructive

---

## 7. Sort Key Mapping

```go
var mongoBsonSortKey = map[string]string{
    "id":        "gormId",
    "createdAt": "createdAt",
    "updatedAt": "updatedAt",
}
```
- Maps API sort field names to BSON field names
- Default fallback: `"createdAt"` when unknown key
- Sort direction: `1` = ASC, `-1` = DESC (MongoDB convention)

---

## 8. Document-Specific — GormID Generation

```go
if version == entity.VersionDraft && doc.GormID == 0 {
    findOpts := options.FindOne().
        SetSort(bson.D{{Key: "gormId", Value: -1}}).
        SetProjection(bson.D{{Key: "gormId", Value: 1}})
    // ... find max gormId, set doc.GormID = max + 1
}
```
- `GormID` is manually managed in MongoDB (no auto-increment)
- Find the max `gormId` in draft documents, increment by 1
- Only for draft version — published copies the draft's `GormID`

---

## 9. Component Storage in MongoDB

- Components are **NOT** stored in separate collections
- Components stay **nested in the document's BSON `data` field**
- Non-repeatable: `{ "banner": { "title": "...", "background": "..." } }` — object
- Repeatable: `{ "skills": [ {...}, {...} ] }` — array
- **NEVER** create MongoDB component collections
- `ComponentRepository` is GORM-only — MongoDB has no component repository

---

## 10. Error Import Convention

```go
import pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
```
- Domain errors imported as `pkgerrors` to avoid collision with stdlib `errors`
- **ALWAYS** use this alias consistently across all MongoDB repo files

---

## 11. Collection Drop

```go
func (r *documentRepository) DropCollection(ctx context.Context, slug string) error {
    return r.collection(slug).Drop(ctx)
}
```
- Only called when a content-type JSON file is deleted (sync engine)
- Drops the entire `documents_<slug>` collection
- **NEVER** call `DropCollection` outside of the sync removal flow

---

## 12. Boundaries Summary

| Rule | Detail |
|---|---|
| **Always** | Use `time.Now().UTC()` for all timestamps |
| **Always** | Check `mongo.IsDuplicateKeyError` → `pkgerrors.ErrConflict` |
| **Always** | Check `mongo.ErrNoDocuments` → `pkgerrors.ErrNotFound` |
| **Always** | `defer cursor.Close(ctx)` after every `Find` |
| **Always** | Use `bson.D` (ordered) for sort keys, `bson.M` (unordered) for filters |
| **Always** | Include interface compliance assertion: `var _ repo.X = (*x)(nil)` |
| **Always** | Use `ReplaceOne` for entity updates (full replacement) |
| **Always** | Pass `*mongo.Database` to constructors (not `*mongo.Client`) |
| **Always** | Components nested in BSON `data` — no separate collections |
| **Never** | Use `UpdateOne` with `$set` for entity updates |
| **Never** | Use `errors.Is` for `mongo.ErrNoDocuments` — use `==` |
| **Never** | Hardcode connection URIs |
| **Never** | Drop document collections in `EnsureCollection` |
| **Never** | Create component collections in MongoDB |
| **Never** | Use `bson.M` for sort keys (unordered maps break sort) |
