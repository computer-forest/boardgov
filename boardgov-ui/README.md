# BoardGov UI — Angular Frontend

Production-grade Angular 17 frontend for the Board Governance System.

## Tech Stack

| Component        | Choice                    |
|-----------------|---------------------------|
| Framework        | Angular 17 (Standalone)   |
| UI Library       | Angular Material v17       |
| Styling          | SCSS + CSS Variables       |
| HTTP             | HttpClient + interceptors  |
| Auth             | JWT with auto-refresh      |
| Charts           | Chart.js + ng2-charts      |
| Fonts            | Playfair Display + IBM Plex Mono |
| Theme            | Dark luxury / governance   |

---

## Project Structure

```
src/app/
├── core/
│   ├── auth/           auth.service, auth.guard, guestGuard
│   ├── interceptors/   JWT auto-attach + 401 refresh
│   ├── models/         All TypeScript interfaces
│   └── services/       MeetingService, DocumentService, DashboardService, AuditService
├── features/
│   ├── auth/           Login + Register pages
│   ├── dashboard/      Full analytics dashboard
│   ├── meetings/       List, Create, Detail
│   ├── documents/      List (queues), Create, Detail (full workflow UI)
│   └── audit/          Immutable audit log viewer + integrity check
├── layout/
│   └── shell/          Sidenav + router-outlet shell
└── shared/
    └── pipes/          ReplacePipe
```

---

## Prerequisites

- Node.js 18+
- Angular CLI 17+: `npm install -g @angular/cli@17`
- Go backend running (see `../boardgov/README.md`)

---

## Quick Start

```bash
# Install dependencies
npm install

# Start dev server (proxies /api → localhost:8080)
npm start

# App available at http://localhost:4200
```

---

## Build for Production

```bash
# Optimized production build
npm run build:prod

# Output: dist/boardgov-ui/
# Deploy to any static file server or NGINX
```

---

## NGINX Static Serving

```nginx
server {
    listen 443 ssl;
    server_name boardgov.internal;

    root /var/www/boardgov-ui;
    index index.html;

    # Angular routing — serve index.html for all routes
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Proxy API to backend
    location /api {
        proxy_pass http://boardgov-api:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## Features

### Auth Module
- Login with email/password
- Register new user
- JWT stored in localStorage
- Automatic token refresh on 401
- Route guards: `authGuard` (protected) + `guestGuard` (login/register)

### Dashboard
- KPI stats: pending approvals, reviews, overdue, avg approval time, rejection rate
- Pending approvals queue (per user)
- Pending reviews queue (per user)
- Meeting status summary with progress bars
- Document aging analysis
- Overdue documents alert panel
- Recent activity timeline
- Links to full audit log

### Meetings
- List all meetings (paginated)
- Create new meeting (Board, Committee, AGM, Special)
- Detail view with documents list

### Documents
- Queue view: Pending Approvals / Pending Reviews / Overdue tabs
- Create document with meeting selector + workflow mode picker
- Detail view with:
  - Visual workflow stepper (Draft → Submitted → Under Review → Approved → Archived)
  - File upload via drag-and-drop or click (PDF/DOCX, max 25 MB)
  - Version history with SHA-256 hashes
  - In-page approval form (Approve / Reject with comments)
  - Approval status tracker per approver

### Audit Log
- Full paginated audit log table
- One-click chain integrity validation
- Hash viewer with tooltip for full hash
- Action colour-coded badges
- State transition display (before → after)

---

## Design System

**Theme:** Dark luxury governance — deep navy (`#070B14`) with gold accents (`#C9A84C`)

**Fonts:**
- Headings: `Playfair Display` (serif, authoritative)
- Body: `DM Sans` (clean, modern)
- Monospace data: `IBM Plex Mono`

**Key CSS variables** (all in `src/styles/main.scss`):
```scss
--gold:       #C9A84C   // Primary accent
--bg-base:    #070B14   // Page background
--bg-card:    #111B2E   // Card background
--text-primary: #F1F5F9
```

---

## Environment Configuration

Edit `src/environments/environment.ts`:

```typescript
export const environment = {
  production: false,
  apiUrl: '/api/v1'   // proxied to backend in dev
};
```

For production, update `environment.prod.ts` with the absolute API URL.

---

## Proxy Configuration (dev)

`proxy.conf.json` proxies all `/api` requests to `http://localhost:8080`:

```json
{ "/api": { "target": "http://localhost:8080", "changeOrigin": true } }
```

---

## Adding the ReplacePipe

The `ReplacePipe` transforms `UNDER_REVIEW` → `UNDER REVIEW` for display.
Import it in any component that uses `| replace` in templates.
