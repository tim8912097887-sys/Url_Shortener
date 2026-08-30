# URL Shortener

A production-oriented full-stack URL shortener built with **React** and **Go**, featuring authentication, OAuth login, Redis caching, URL expiration, and multi-layer rate limiting.

## 🚀 Tech Stack

### Frontend

- React
- TypeScript
- Zustand
- React Router
- Tailwind CSS

### Backend

- Golang
- REST API

### Database & Cache

- PostgreSQL
- Redis

### Deployment

| Service    | Platform |
| ---------- | -------- |
| Frontend   | Vercel   |
| Backend    | Render   |
| PostgreSQL | Neon     |
| Redis      | Upstash  |

---

# ✨ Features

## Authentication

The application provides a complete authentication flow:

- User signup
- User login
- User logout
- Access token refresh
- OAuth login
- Token-based authentication

Authentication allows the application to provide different URL policies for authenticated and unauthenticated users.

---

## URL Shortening

Users can create shortened URLs with different expiration policies.

### Authenticated Users

- Shortened URLs expire after **30 days**
- Redis cache TTL is extended by **24 hours** when the URL is accessed

### Unauthenticated Users

- Shortened URLs expire after **7 days**
- Redis cache TTL is extended by **15 minutes** when the URL is accessed

---

# ⚡ Cache Strategy

## Cache-Aside Pattern

The URL lookup flow uses the **cache-aside pattern** to reduce database load and improve redirect performance.

```text
Client
  │
  ▼
URL Request
  │
  ▼
Redis Cache
  │
  ├── Cache Hit ──────────► Return Long URL
  │
  └── Cache Miss
          │
          ▼
      PostgreSQL
          │
          ▼
     Store in Redis
          │
          ▼
     Return Long URL
```

# 🏗️ Architecture

```text
                    ┌─────────────────┐
                    │     Client      │
                    │ React + TS      │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │      Vercel     │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │   Go Backend    │
                    │     Render      │
                    └───────┬─────────┘
                            │
              ┌─────────────┴─────────────┐
              ▼                           ▼
     ┌─────────────────┐         ┌─────────────────┐
     │   PostgreSQL    │         │      Redis      │
     │      Neon       │         │     Upstash     │
     └─────────────────┘         └─────────────────┘
```

# 🔄 URL Redirect Flow

```text
User Request
│
▼
Rate Limiter
│
▼
Redis Lookup
│
├── Cache Hit
│ │
│ ▼
│ Extend Cache TTL
│ │
│ ▼
│ Redirect User
│
└── Cache Miss
│
▼
PostgreSQL
│
├── Not Found / Expired
│
└── URL Found
│
▼
Store in Redis
│
▼
Redirect User
```

# 🧠 Key Design Decisions

## Redis for Fast Redirects

URL redirects are performance-sensitive because every redirect may require a database lookup.

Redis reduces latency and database load by serving frequently accessed URLs directly from cache.

## 🔑 Token Versioning

To strengthen authentication and logout flows, the system implements token versioning:

- Each user account maintains a token version field.

- On single-device logout, only the refresh/access tokens tied to that device are invalidated.

- On global logout (all devices), the token version is incremented.

Any existing tokens with the old version become invalid immediately.

New login or refresh requests issue tokens bound to the updated version.

This design ensures fine-grained control over session invalidation while keeping token checks efficient.

## 📜 Cursor-Based Pagination

The application supports cursor-based pagination for listing shortened URLs:

- Instead of offset-based pagination, the URL expiration timestamp is used as the cursor.

- This approach ensures stable ordering even when URLs are deleted or expire.

- Example flow:

- Query returns URLs ordered by expiration time.

- The last item’s expiration timestamp becomes the next cursor.

- Subsequent queries fetch URLs with expiration times greater than the cursor.

This design leverages natural lifecycle data (expiration) to provide efficient, consistent pagination without gaps.

## Rolling TTL

Instead of keeping every cached URL for a fixed duration regardless of usage, active URLs receive a TTL extension.

Inactive URLs naturally expire from Redis, preventing cold URLs from consuming cache memory unnecessarily.

## Different User Policies

Authenticated users receive longer expiration periods:

- 30-day URL lifetime
- 24-hour rolling cache extension

Unauthenticated users use shorter durations:

- 7-day URL lifetime
- 15-minute rolling cache extension

This provides a balance between user experience and infrastructure resource usage.

## Layered Rate Limiting

Using both token bucket and fixed-window strategies provides stronger protection than relying on a single algorithm.

- Token bucket handles bursts.
- Fixed window controls sustained traffic.

# 🌐 Production Deployment

```text
Vercel
│
│ HTTPS API Requests
▼
Render
│
├──────────────► Neon PostgreSQL
│
└──────────────► Upstash Redis
```

# 🎯 Project Goals

This project demonstrates production-oriented full-stack development concepts:

- Authentication and authorization
- OAuth integration
- Access and refresh token handling
- Redis cache-aside pattern
- Rolling cache expiration
- URL lifecycle management
- Rate limiting
- Burst traffic handling
- Distributed cloud deployment
- Frontend state management
- Client-side routing
- Type-safe frontend development
