# Tech Stack

## Backend (core)
- **Language**: Go 1.22+
- **Framework**: GoFrame (gogf/gf v2) — provides routing, ORM, config, logging, context management
- **Database**: PostgreSQL 17 (via `gogf/gf/contrib/drivers/pgsql/v2`)
- **Cache/Queue**: Redis 7.4 (via `gogf/gf/contrib/nosql/redis/v2`)
- **Architecture**: Controller → Service → DAO layered pattern (GoFrame convention)
- **Auth**: JWT (`golang-jwt/jwt/v5`)
- **Docker SDK**: `docker/docker` client for container management
- **ACME/SSL**: `go-acme/lego/v4` for Let's Encrypt certificates

## Frontend (core/frontend)
- **Language**: TypeScript
- **Framework**: Vue 3 (Composition API)
- **Build Tool**: Rsbuild (Rspack-based)
- **UI Library**: Naive UI
- **State**: Pinia (with persisted state plugin)
- **Styling**: UnoCSS (Attributify + Icons presets), SASS
- **Routing**: Vue Router 4
- **i18n**: vue-i18n
- **Code Editor**: Monaco Editor
- **Charts**: ECharts
- **HTTP Client**: Axios
- **Package Manager**: pnpm

## Infrastructure (Docker Compose)
- **Postfix**: SMTP server (custom Docker image `billionmail/postfix`)
- **Dovecot**: IMAP/POP3 server (custom Docker image `billionmail/dovecot`)
- **Rspamd**: Spam filtering and DKIM signing (custom Docker image `billionmail/rspamd`)
- **Roundcube**: Webmail client (official `roundcube/roundcubemail` image)
- **Core**: Management API + admin panel (custom Docker image `billionmail/core`)
- **PostgreSQL**: Primary data store (official `postgres:17.4-alpine`)
- **Redis**: Caching and Rspamd statistics (official `redis:7.4.2-alpine`)

## Common Commands

### Docker (production)
```bash
# Start all services
docker compose up -d

# Stop all services
docker compose down

# View logs for a service
docker compose logs -f <service-name>

# Management CLI
bash bm.sh help          # Show available commands
bash bm.sh default       # Show login info
bash bm.sh update        # Update BillionMail
bash bm.sh show-record   # Show DNS records for domain
```

### Backend (Go)
```bash
cd core
go mod tidy              # Sync dependencies
go run main.go           # Run in dev mode
sh go-build.sh all       # Build for both amd64 and arm64
sh go-build.sh x86       # Build for amd64 only
sh go-build.sh arm       # Build for arm64 only
```

### Frontend
```bash
cd core/frontend
pnpm install             # Install dependencies
pnpm dev                 # Start dev server
pnpm build              # Production build
pnpm lint               # Run ESLint
pnpm lint:fix           # Auto-fix lint issues
pnpm format             # Run Prettier
```
