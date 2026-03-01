# Board Governance System

Production-grade Board & Committee Documentation Governance System built with Go 1.22, Gin, PostgreSQL, Clean Architecture, JWT, dynamic RBAC, immutable audit logs with hash chaining, and a full workflow state machine.

---

## Architecture

```
boardgov/
├── cmd/server/          # Entrypoint
├── config/              # Env-based config loader
├── internal/
│   ├── domain/          # Entities, interfaces, state machine
│   ├── usecase/         # Business logic (Auth, Document, Meeting)
│   ├── repository/
│   │   └── postgres/    # pgx implementations
│   ├── delivery/
│   │   └── http/        # Gin router + handlers
│   └── middleware/      # JWT, RBAC, rate limiter, secure headers
├── pkg/
│   ├── crypto/          # bcrypt, JWT, SHA256, audit chain hash
│   ├── logger/          # zerolog setup
│   └── validator/       # struct + file validation
├── migrations/          # SQL migration files
├── deployments/
│   ├── k8s/             # Kubernetes manifests
│   └── nginx/           # NGINX reverse proxy config
└── tests/unit/          # Unit tests
```

---

## Tech Stack

| Component       | Technology                    |
|----------------|-------------------------------|
| Language        | Go 1.22                       |
| Framework       | Gin                           |
| Database        | PostgreSQL 16 (pgx/v5 driver) |
| Auth            | JWT (access + refresh tokens) |
| RBAC            | Dynamic, DB-driven            |
| Logging         | zerolog (structured JSON)     |
| File Storage    | Local encrypted volume        |
| Container       | Docker / Kubernetes           |
| Reverse Proxy   | NGINX                         |

---

## Quick Start (Docker Compose)

```bash
# 1. Clone and enter repo
git clone <repo-url> && cd boardgov

# 2. Copy env config
cp .env.example .env
# Edit .env — change JWT secrets!

# 3. Start PostgreSQL + API
docker compose up -d

# 4. API is live at http://localhost:8080
```

---

## Local Development (without Docker)

### Prerequisites
- Go 1.22+
- PostgreSQL 16+
- `psql` CLI

### Steps

```bash
# 1. Install dependencies
go mod download

# 2. Create database
createdb boardgov

# 3. Run migrations
psql "postgres://postgres:postgres@localhost:5432/boardgov?sslmode=disable" -f migrations/001_initial.sql

# 4. Set environment variables
export DATABASE_DSN="postgres://postgres:postgres@localhost:5432/boardgov?sslmode=disable"
export JWT_ACCESS_SECRET="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
export JWT_REFRESH_SECRET="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
export STORAGE_BASE_PATH="/tmp/boardgov-storage"
export APP_ENV="development"
export CSRF_SECRET="1f3a5b7c9d1e3f5g7h9i1j3k5l7m9n1o3p5q7r9s1t3u5v7w9x1y3z5a7b9c1d3e"
export ENCRYPTION_KEY="development"

# 5. Run server
make run
# or: go run ./cmd/server
```

---

## Building

```bash
# Binary
make build
# Output: bin/boardgov

# Docker image
make docker
# Output: boardgov:latest
```

---

## Running Tests

```bash
# All tests with race detector
make test

# Unit tests only
make test-unit

# Coverage report (opens browser)
make coverage
```

---

## API Reference

### Authentication

| Method | Endpoint                | Description         |
|--------|------------------------|---------------------|
| POST   | `/api/v1/auth/register`| Register user       |
| POST   | `/api/v1/auth/login`   | Login, get tokens   |
| POST   | `/api/v1/auth/refresh` | Refresh access token|

**Login request:**
```json
POST /api/v1/auth/login
{
  "email": "admin@example.com",
  "password": "SecurePass123!"
}
```

**Login response:**
```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "expires_at": 1700000000,
  "user": { "id": "...", "email": "...", "full_name": "..." }
}
```

All protected endpoints require: `Authorization: Bearer <access_token>`

---

### Meetings

| Method | Endpoint             | Permission          |
|--------|---------------------|---------------------|
| POST   | `/api/v1/meetings`  | meetings:create     |
| GET    | `/api/v1/meetings`  | meetings:read       |
| GET    | `/api/v1/meetings/:id` | meetings:read    |

**Create meeting:**
```json
POST /api/v1/meetings
{
  "title": "Q4 Board Meeting",
  "meeting_type": "BOARD",
  "scheduled_at": "2025-01-15T09:00:00Z",
  "location": "Boardroom A",
  "agenda": "Q4 review, budget approval"
}
```

Meeting types: `BOARD`, `COMMITTEE`, `AGM`, `SPECIAL`

---

### Documents — Full Workflow

```
DRAFT → SUBMITTED → UNDER_REVIEW → APPROVED → ARCHIVED
                                 ↘ REJECTED → DRAFT
```

| Method | Endpoint                            | Permission           |
|--------|-------------------------------------|----------------------|
| POST   | `/api/v1/documents`                 | documents:create     |
| POST   | `/api/v1/documents/:id/upload`      | documents:upload     |
| PUT    | `/api/v1/documents/:id/reviewers`   | documents:assign     |
| PUT    | `/api/v1/documents/:id/approvers`   | documents:assign     |
| POST   | `/api/v1/documents/:id/submit`      | documents:submit     |
| POST   | `/api/v1/documents/:id/review/start`| documents:review     |
| POST   | `/api/v1/documents/:id/approve`     | documents:approve    |
| GET    | `/api/v1/documents/:id/download`    | documents:download   |
| POST   | `/api/v1/documents/:id/archive`     | documents:archive    |

**Step 1 — Create:**
```json
POST /api/v1/documents
{
  "meeting_id": "uuid",
  "title": "Annual Report 2024",
  "description": "Full year financial summary",
  "workflow_mode": "SEQUENTIAL"
}
```

**Step 2 — Upload file (multipart):**
```bash
curl -X POST /api/v1/documents/{id}/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@report.pdf"
```
- Max size: 25 MB
- Allowed types: `.pdf`, `.docx`
- File renamed to UUID, stored at `/secure-storage/{meeting_id}/{doc_id}/v{N}/`
- SHA-256 hash stored in DB, verified on every download

**Step 3 — Assign reviewers (dynamic):**
```json
PUT /api/v1/documents/:id/reviewers
{
  "user_ids": ["uuid1", "uuid2"]
}
```
In `SEQUENTIAL` mode, order_index determines sequence. In `PARALLEL`, all act simultaneously.

**Step 4 — Assign approvers:**
```json
PUT /api/v1/documents/:id/approvers
{ "user_ids": ["uuid3", "uuid4"] }
```

**Step 5 — Submit:**
```
POST /api/v1/documents/:id/submit
```

**Step 6 — Start review:**
```
POST /api/v1/documents/:id/review/start
```

**Step 7 — Approve or Reject:**
```json
POST /api/v1/documents/:id/approve
{
  "status": "APPROVED",
  "comments": "Looks good, approved."
}
```
When all approvers approve → document status becomes `APPROVED` and is **locked** (immutable).

**Step 8 — Download:**
```
GET /api/v1/documents/:id/download?version=1
```
Returns file with `Content-Disposition: attachment`. Hash integrity verified before serving.

---

### Dashboard

| Endpoint                              | Description                     |
|---------------------------------------|---------------------------------|
| GET `/api/v1/dashboard/pending-approvals` | Approvals pending for me     |
| GET `/api/v1/dashboard/pending-reviews`   | Reviews pending for me       |
| GET `/api/v1/dashboard/overdue`           | Documents overdue >7 days    |
| GET `/api/v1/dashboard/meeting-summary`   | Per-meeting status counts    |
| GET `/api/v1/dashboard/aging`             | Document aging buckets       |
| GET `/api/v1/dashboard/turnaround`        | Avg/P50/P90 approval hours   |
| GET `/api/v1/dashboard/rejection-ratio`   | Rejection rate analytics     |
| GET `/api/v1/dashboard/timeline`          | Activity feed (paginated)    |
| GET `/api/v1/dashboard/audit-integrity`   | Hash chain validation        |

All endpoints support `?offset=0&limit=20` pagination.

---

### Health Endpoints (unauthenticated)

```
GET /health  → {"status":"ok"}
GET /ready   → {"status":"ready"}
GET /live    → {"status":"alive"}
```

---

## RBAC — Role & Permission System

Roles and permissions are **fully DB-driven** — no hardcoded roles in code.

**Default roles seeded by migration:**

| Role              | Description                     |
|------------------|---------------------------------|
| `super_admin`    | Full system access              |
| `board_secretary`| Create/manage meetings & docs   |
| `board_member`   | Read, approve, download         |
| `reviewer`       | Read, review, download          |
| `viewer`         | Read-only                       |

**Assign a role to a user (via psql):**
```sql
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM users u, roles r
WHERE u.email='admin@example.com' AND r.name='super_admin';
```

**Add custom role with permissions:**
```sql
INSERT INTO roles (id, name, description) 
VALUES (uuid_generate_v4(), 'compliance_officer', 'Compliance audit access');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name='compliance_officer' AND p.name IN ('audit:read','documents:read','meetings:read');
```

---

## Audit Log — Immutable Hash Chain

Every state transition, upload, download, and login is recorded in `audit_logs`.

- **No UPDATE** — enforced via Postgres rule
- **No DELETE** — enforced via Postgres rule
- Each log entry contains `hash` = SHA-256 of `(user_id | action | doc_id | timestamp | previous_hash)`
- Chain validated via `GET /api/v1/dashboard/audit-integrity`

**Validate integrity:**
```
GET /api/v1/dashboard/audit-integrity
→ { "valid": true, "errors": [] }
```

---

## Workflow Engine — State Machine

Transitions enforced in `internal/domain/workflow.go`:

```
DRAFT       → SUBMITTED
SUBMITTED   → UNDER_REVIEW | REJECTED
UNDER_REVIEW → APPROVED | REJECTED
APPROVED    → ARCHIVED
REJECTED    → DRAFT | ARCHIVED
ARCHIVED    → (terminal)
```

- **SEQUENTIAL mode**: reviewers/approvers act in `order_index` sequence
- **PARALLEL mode**: all reviewers/approvers act simultaneously; all must approve
- Document is **locked** (immutable) upon reaching `APPROVED`
- State skip prevention is enforced — any invalid transition returns HTTP 422

---

## Security

| Feature                | Implementation                          |
|-----------------------|-----------------------------------------|
| JWT Auth              | HS256, 15-min access / 7-day refresh   |
| RBAC                  | Dynamic, DB-driven per-request check   |
| Rate Limiting         | Token bucket per IP (golang.org/x/time)|
| IP Whitelist          | Configurable via env                   |
| Secure Headers        | CSP, HSTS, X-Frame-Options, etc.       |
| File Validation       | Extension + MIME magic byte check      |
| File Integrity        | SHA-256 hash verified on every download|
| Audit Trail           | Immutable hash-chained log             |
| DB Connection Pool    | pgxpool with configurable limits       |
| Panic Recovery        | Middleware with structured logging     |
| No Direct File URLs   | Files served only via API download     |
| Read-only Filesystem  | Dockerfile runs as UID 10001, scratch  |

---

## Kubernetes Deployment

```bash
# Apply all manifests
kubectl apply -f deployments/k8s/deployment.yaml

# Check rollout
kubectl rollout status deployment/boardgov-api -n boardgov

# View logs
kubectl logs -n boardgov -l app=boardgov-api -f

# Port-forward for testing
kubectl port-forward -n boardgov svc/boardgov-svc 8080:80
```

**Production checklist:**
- Replace `Secret` values with Sealed Secrets or Vault
- Use `storageClassName: standard-encrypted` for PVC (encrypted at rest)
- Configure NetworkPolicy ingress for your NGINX namespace label
- Enable PodSecurityAdmission for `restricted` policy

---

## NGINX

```bash
# Install config
cp deployments/nginx/boardgov.conf /etc/nginx/sites-available/boardgov
ln -s /etc/nginx/sites-available/boardgov /etc/nginx/sites-enabled/
nginx -t && nginx -s reload
```

---

## Environment Variables Reference

| Variable              | Default              | Description                    |
|-----------------------|----------------------|--------------------------------|
| `APP_ENV`             | `production`         | `development` for console logs |
| `SERVER_PORT`         | `8080`               | HTTP listen port               |
| `DATABASE_DSN`        | *(required)*         | pgx connection string          |
| `JWT_ACCESS_SECRET`   | *(required)*         | Min 32 chars                   |
| `JWT_REFRESH_SECRET`  | *(required)*         | Min 32 chars, different        |
| `JWT_ACCESS_EXPIRY`   | `15m`                | Access token TTL               |
| `JWT_REFRESH_EXPIRY`  | `168h`               | Refresh token TTL              |
| `STORAGE_BASE_PATH`   | `/secure-storage`    | File storage root              |
| `STORAGE_MAX_SIZE_MB` | `25`                 | Max upload size                |
| `STORAGE_ALLOWED_EXT` | `.pdf,.docx`         | Comma-separated                |
| `IP_WHITELIST`        | *(empty = allow all)*| Comma-separated CIDRs          |
| `RATE_LIMIT_RPS`      | `20`                 | Requests/sec per IP            |
| `RATE_LIMIT_BURST`    | `40`                 | Burst allowance                |
| `DB_MAX_OPEN_CONNS`   | `25`                 | pgxpool max connections        |

---

## Migration

```bash
# Apply
psql "$DATABASE_DSN" -f migrations/001_initial.sql

# Rollback (manual — drop and recreate)
psql "$DATABASE_DSN" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
psql "$DATABASE_DSN" -f migrations/001_initial.sql
```

---

## Make Targets

```
make build       Build binary to bin/boardgov
make run         Run server locally
make test        All tests with race detector
make test-unit   Unit tests only
make coverage    HTML coverage report
make migrate     Apply SQL migrations
make lint        golangci-lint
make docker      Build Docker image
make tidy        go mod tidy
```
