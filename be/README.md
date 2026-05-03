# Lot-Based Inventory Accounting Backend

## Overview

A backend service for lot-based inventory accounting — tracking inventory by discrete lots (batches) with full traceability of cost, quantity, age, and movement history. Built with Go, PostgreSQL, and designed for a Next.js frontend.

---

## Tech Stack

| Layer        | Technology               |
|--------------|--------------------------|
| Backend API  | Go (net/http + chi router) |
| Database     | PostgreSQL 16+           |
| Migrations   | golang-migrate           |
| Auth         | JWT (access + refresh tokens) with HMAC-SHA256 |
| Frontend     | Next.js (future)         |
| API Protocol | REST (JSON)              |
| Deployment   | Docker / Docker Compose  |

---

## Architecture

```
┌─────────────┐       HTTPS/TLS        ┌──────────────────────────────┐
│  Next.js FE  │ ◄────────────────────► │        Go Backend API        │
│  (future)    │   JWT in HttpOnly      │                              │
│              │   Secure Cookies       │  ┌────────┐  ┌───────────┐  │
└─────────────┘                         │  │Handler │─►│ Service   │  │
                                        │  │ Layer  │  │  Layer    │  │
                                        │  └────────┘  └─────┬─────┘  │
                                        │                    │        │
                                        │              ┌─────▼─────┐  │
                                        │              │Repository │  │
                                        │              │  Layer    │  │
                                        │              └─────┬─────┘  │
                                        └────────────────────┼────────┘
                                                             │
                                                      ┌──────▼──────┐
                                                      │ PostgreSQL  │
                                                      │   Database  │
                                                      └─────────────┘
```

### Layered Structure

```
cmd/
  server/
    main.go                  # Entrypoint — wires dependencies, starts server

internal/
  config/
    config.go                # Env-based configuration (no .env file reads in code)

  middleware/
    auth.go                  # JWT validation, RBAC enforcement
    cors.go                  # CORS policy for Next.js origin
    ratelimit.go             # Request rate limiting
    requestid.go             # Request ID injection for tracing
    logging.go               # Structured request logging

  auth/
    handler.go               # Login, refresh, logout endpoints
    service.go               # Token generation, password verification
    tokens.go                # JWT creation / validation helpers

  user/
    handler.go
    service.go
    repository.go
    model.go                 # User, Role models

  lot/
    handler.go               # CRUD for inventory lots
    service.go               # Lot business logic (costing, aging)
    repository.go            # SQL queries for lots
    model.go                 # Lot, LotStatus models

  inventory/
    handler.go               # Stock movements, adjustments
    service.go               # Movement validation, lot allocation
    repository.go
    model.go                 # Movement, Adjustment models

  product/
    handler.go
    service.go
    repository.go
    model.go                 # Product, Category models

  warehouse/
    handler.go
    service.go
    repository.go
    model.go                 # Warehouse, Location models

  report/
    handler.go               # Inventory valuation, lot aging reports
    service.go

  db/
    postgres.go              # Connection pool setup
    tx.go                    # Transaction helper

migrations/
  000001_init_schema.up.sql
  000001_init_schema.down.sql

docker-compose.yml
Dockerfile
go.mod
go.sum
```

---

## Database Schema (Core Tables)

```sql
-- Users & access control
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT UNIQUE NOT NULL,
    password    TEXT NOT NULL,  -- bcrypt hash
    role        TEXT NOT NULL DEFAULT 'viewer', -- admin, manager, viewer
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Products being tracked
CREATE TABLE products (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku         TEXT UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    unit        TEXT NOT NULL,  -- kg, pcs, liters, etc.
    category    TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Warehouses / storage locations
CREATE TABLE warehouses (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT UNIQUE NOT NULL,
    address     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Core: Inventory lots (batches)
CREATE TABLE lots (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lot_number      TEXT UNIQUE NOT NULL,
    product_id      UUID NOT NULL REFERENCES products(id),
    warehouse_id    UUID NOT NULL REFERENCES warehouses(id),
    quantity         NUMERIC(15,4) NOT NULL CHECK (quantity >= 0),
    unit_cost        NUMERIC(15,4) NOT NULL CHECK (unit_cost >= 0),
    total_cost       NUMERIC(15,4) GENERATED ALWAYS AS (quantity * unit_cost) STORED,
    received_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expiry_date      DATE,
    status           TEXT NOT NULL DEFAULT 'available', -- available, reserved, depleted, expired
    supplier         TEXT,
    reference_doc    TEXT,  -- PO number, GRN, etc.
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_lots_product ON lots(product_id);
CREATE INDEX idx_lots_warehouse ON lots(warehouse_id);
CREATE INDEX idx_lots_status ON lots(status);
CREATE INDEX idx_lots_received ON lots(received_at);

-- Stock movements against specific lots
CREATE TABLE lot_movements (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lot_id          UUID NOT NULL REFERENCES lots(id),
    movement_type   TEXT NOT NULL, -- receipt, issue, transfer, adjustment, return
    quantity         NUMERIC(15,4) NOT NULL,
    reference_doc    TEXT,
    notes            TEXT,
    performed_by     UUID REFERENCES users(id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_movements_lot ON lot_movements(lot_id);
CREATE INDEX idx_movements_type ON lot_movements(movement_type);
CREATE INDEX idx_movements_created ON lot_movements(created_at);

-- Audit log for all state changes
CREATE TABLE audit_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name  TEXT NOT NULL,
    record_id   UUID NOT NULL,
    action      TEXT NOT NULL, -- INSERT, UPDATE, DELETE
    old_data    JSONB,
    new_data    JSONB,
    performed_by UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_record ON audit_log(table_name, record_id);
```

---

## FE ↔ BE Security Model

### Authentication Flow

1. **Login** — `POST /api/v1/auth/login` with email/password
2. Backend verifies bcrypt hash, issues:
   - **Access token** (JWT, 15 min TTL) — sent in `HttpOnly`, `Secure`, `SameSite=Strict` cookie
   - **Refresh token** (opaque, 7 day TTL) — stored in DB, sent in separate `HttpOnly` cookie
3. **Every request** — middleware extracts JWT from cookie, validates signature + expiry
4. **Token refresh** — `POST /api/v1/auth/refresh` rotates both tokens (refresh token rotation prevents reuse)
5. **Logout** — invalidates refresh token in DB, clears cookies

### Security Measures

| Concern               | Approach                                                       |
|------------------------|----------------------------------------------------------------|
| XSS token theft        | Tokens in `HttpOnly` + `Secure` + `SameSite=Strict` cookies   |
| CSRF                   | `SameSite=Strict` cookies + CSRF token header for mutations    |
| CORS                   | Allowlist Next.js origin only                                  |
| SQL injection           | Parameterized queries via `pgx` — no string concatenation     |
| Brute force             | Rate limiting on auth endpoints                                |
| Password storage        | bcrypt with cost 12+                                           |
| Transport               | TLS termination at reverse proxy (nginx/caddy)                 |
| Authorization            | Role-based access control enforced in middleware               |
| Input validation        | Validate all payloads at handler layer before service calls    |
| Audit trail             | All mutations logged to `audit_log` table                     |

### CORS Configuration

```go
AllowedOrigins:   []string{"https://your-nextjs-domain.com"}
AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
AllowedHeaders:   []string{"Content-Type", "X-CSRF-Token"}
AllowCredentials: true
```

---

## API Routes

```
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout

GET    /api/v1/products
POST   /api/v1/products
GET    /api/v1/products/:id
PUT    /api/v1/products/:id
DELETE /api/v1/products/:id

GET    /api/v1/warehouses
POST   /api/v1/warehouses
GET    /api/v1/warehouses/:id
PUT    /api/v1/warehouses/:id

GET    /api/v1/lots
POST   /api/v1/lots
GET    /api/v1/lots/:id
PUT    /api/v1/lots/:id
GET    /api/v1/lots/:id/movements

POST   /api/v1/movements              # Record a stock movement
GET    /api/v1/movements               # List movements (filterable)

GET    /api/v1/reports/valuation        # Inventory valuation by lot
GET    /api/v1/reports/aging            # Lot aging analysis
GET    /api/v1/reports/movement-summary # Movement summary over period
```

---

## Lot-Based Accounting Concepts

### What is a "Lot"?

A **lot** is a batch of the same product purchased at a specific price on a specific date. Instead of averaging all your inventory into one number, you track each purchase separately so you always know exactly what you paid, when you got it, and how much is left.

### Why bother?

Without lots, you just know "I have 50 packs in stock." With lots, you know "I have 20 packs I bought at ₱150 on Monday and 30 packs I bought at ₱180 on Wednesday." This matters when prices fluctuate because it directly affects your profit calculation and inventory valuation.

---

### Real-Life Example: Selling Pokémon Card Packs

You run a small shop that buys and resells Pokémon Scarlet & Violet booster packs. Supplier prices change daily based on demand.

#### Step 1: You buy inventory (creating lots)

| Date   | Lot # | Packs Bought | Unit Cost | Total Cost |
|--------|-------|-------------|-----------|------------|
| May 1  | LOT-001 | 20       | ₱150      | ₱3,000     |
| May 3  | LOT-002 | 30       | ₱180      | ₱5,400     |
| May 5  | LOT-003 | 15       | ₱160      | ₱2,400     |

Your warehouse now has **65 packs** worth **₱10,800** total. Each lot remembers its own cost.

#### Step 2: A customer buys 25 packs at ₱220 each

Now you need to decide **which lots** to pull from. This is where the **costing method** matters:

---

**FIFO (First In, First Out)** — sell oldest lots first

| Source  | Packs Sold | Unit Cost | COGS     |
|---------|-----------|-----------|----------|
| LOT-001 | 20        | ₱150      | ₱3,000   |
| LOT-002 | 5         | ₱180      | ₱900     |
| **Total** | **25**  |           | **₱3,900** |

- Revenue: 25 × ₱220 = **₱5,500**
- Cost of Goods Sold (COGS): **₱3,900**
- Gross Profit: **₱1,600**

LOT-001 is now **depleted** (0 remaining). LOT-002 has **25 packs** left.

---

**LIFO (Last In, First Out)** — sell newest lots first

| Source  | Packs Sold | Unit Cost | COGS     |
|---------|-----------|-----------|----------|
| LOT-003 | 15        | ₱160      | ₱2,400   |
| LOT-002 | 10        | ₱180      | ₱1,800   |
| **Total** | **25**  |           | **₱4,200** |

- Revenue: 25 × ₱220 = **₱5,500**
- Cost of Goods Sold (COGS): **₱4,200**
- Gross Profit: **₱1,300**

Same sale, different profit — because LIFO used the more expensive recent lots first.

---

**Specific Identification** — you pick exactly which lots to pull from

Maybe you intentionally want to sell LOT-002 first because those packs are from a less popular print run:

| Source  | Packs Sold | Unit Cost | COGS     |
|---------|-----------|-----------|----------|
| LOT-002 | 25        | ₱180      | ₱4,500   |
| **Total** | **25**  |           | **₱4,500** |

- Gross Profit: **₱1,000**

You chose a specific lot — full control, full traceability.

---

#### Step 3: What's left in stock? (Inventory Valuation)

After selling 25 packs using **FIFO**:

| Lot     | Remaining | Unit Cost | Value    |
|---------|----------|-----------|----------|
| LOT-001 | 0        | ₱150      | ₱0       |
| LOT-002 | 25       | ₱180      | ₱4,500   |
| LOT-003 | 15       | ₱160      | ₱2,400   |
| **Total** | **40** |           | **₱6,900** |

Without lot tracking, you'd only know "40 packs in stock" and have to guess the value. With lots, you know the exact value is ₱6,900.

#### Step 4: Lot Aging (shelf life awareness)

| Lot     | Received | Days Old | Status    |
|---------|----------|----------|-----------|
| LOT-002 | May 3    | 2 days   | available |
| LOT-003 | May 5    | 0 days   | available |

For Pokémon cards this isn't critical, but if you were tracking perishable goods (or sets that rotate out of competitive play), aging tells you which lots to move first.

---

### Key Takeaways

| Concept                | What it answers                                              |
|------------------------|--------------------------------------------------------------|
| **Lot**                | "Which specific batch did this come from?"                   |
| **FIFO**               | "Sell oldest stock first" — higher profit when prices rise   |
| **LIFO**               | "Sell newest stock first" — lower profit, lower tax exposure |
| **Specific ID**        | "I choose exactly which batch to sell"                       |
| **Valuation**          | "What is my remaining inventory actually worth?"             |
| **COGS**               | "What did the items I just sold actually cost me?"           |
| **Aging**              | "How long has this batch been sitting in my warehouse?"      |
| **Movement tracking**  | "Full history of every receipt, sale, transfer, adjustment"  |

### How this maps to the system

- `POST /api/v1/lots` — you bought 20 packs at ₱150 → creates LOT-001
- `POST /api/v1/movements` with `movement_type: "issue"` — customer buys 25 packs → system deducts from lots based on costing method
- `GET /api/v1/reports/valuation` — "what's my stock worth right now?" → sums remaining qty × unit_cost per lot
- `GET /api/v1/reports/aging` — "which lots have been sitting the longest?" → sorted by received_at

---

## Running Locally

```bash
# Start PostgreSQL
docker compose up -d db

# Run migrations
go run cmd/migrate/main.go up

# Start server
go run cmd/server/main.go
```

### Environment Variables

| Variable            | Description                    |
|---------------------|--------------------------------|
| `DB_HOST`           | PostgreSQL host                |
| `DB_PORT`           | PostgreSQL port                |
| `DB_USER`           | Database user                  |
| `DB_PASSWORD`       | Database password              |
| `DB_NAME`           | Database name                  |
| `JWT_SECRET`        | HMAC signing key for JWTs      |
| `CORS_ORIGIN`       | Allowed frontend origin        |
| `SERVER_PORT`       | API server port (default 8080) |

---

## Development Roadmap

- [x] Architecture design
- [ ] Project scaffolding (Go modules, Docker Compose)
- [ ] Database migrations
- [ ] Auth module (login, JWT, refresh rotation)
- [ ] User management
- [ ] Product CRUD
- [ ] Warehouse CRUD
- [ ] Lot management (create, update status)
- [ ] Stock movements (receipt, issue, transfer, adjustment)
- [ ] Inventory valuation reports
- [ ] Lot aging reports
- [ ] Audit logging
- [ ] Next.js frontend integration
