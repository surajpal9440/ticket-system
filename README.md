# Ticket System

A REST API backend for a ticket management system built with **Go**, **SQLite**, and **JWT** authentication.

---

## Tech Stack

- **Language:** Go 1.22
- **Database:** SQLite (via `modernc.org/sqlite` — pure Go, no CGO)
- **Auth:** JWT (HS256, 24h expiry)
- **Password Hashing:** bcrypt

---

## Local Run (without Docker)

### 1. Set environment variables

```bash
cp .env.example .env
# Edit .env with your JWT secret (DB_PATH is optional, defaults to ./tickets.db)
```

### 2. Run the service

```powershell
$env:JWT_SECRET = "your-secret-key"
go run .
```

Server starts on `http://localhost:8080`. The SQLite database file (`tickets.db`) is created automatically on first run.

---

## Local Run (with Docker)

```bash
docker build -t ticket-system .

docker run -p 8080:8080 \
  -e JWT_SECRET="your-secret-key" \
  ticket-system
```

> **Note:** The SQLite DB file is created inside the container. Data will reset when the container is restarted. For persistence, mount a volume:
> ```bash
> docker run -p 8080:8080 \
>   -e JWT_SECRET="your-secret-key" \
>   -v $(pwd)/data:/app/data \
>   -e DB_PATH=/app/data/tickets.db \
>   ticket-system
> ```

---

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/health` | No | Health check |
| POST | `/auth/register` | No | Register a new user |
| POST | `/auth/login` | No | Login and receive JWT |
| POST | `/tickets` | Yes | Create a ticket |
| GET | `/tickets` | Yes | List own tickets |
| GET | `/tickets/{id}` | Yes | Get own ticket by ID |
| PATCH | `/tickets/{id}/status` | Yes | Update ticket status |

### Authentication

Include the JWT token in the `Authorization` header:
```
Authorization: Bearer <token>
```

### Ticket Status Flow

```
open → in_progress → closed
```
- A `closed` ticket **cannot** be moved back to any previous status.

---

## Example Requests

### Health Check
```bash
curl http://localhost:8080/health
```

### Register
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "secret123"}'
```

### Login
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "secret123"}'
```

### Create Ticket
```bash
curl -X POST http://localhost:8080/tickets \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title": "Bug in login", "description": "Login page crashes on mobile"}'
```

### List Tickets
```bash
curl http://localhost:8080/tickets \
  -H "Authorization: Bearer <token>"
```

### Get Ticket by ID
```bash
curl http://localhost:8080/tickets/<id> \
  -H "Authorization: Bearer <token>"
```

### Update Status
```bash
# open → in_progress
curl -X PATCH http://localhost:8080/tickets/<id>/status \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"status": "in_progress"}'

# in_progress → closed
curl -X PATCH http://localhost:8080/tickets/<id>/status \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"status": "closed"}'
```

---

## Deployment URL

> **Live URL:** https://ticket-system-yc79.onrender.com
> **Health Check:** https://ticket-system-yc79.onrender.com/health

---

## Assumptions

- SQLite database file is auto-created on startup — no manual migration needed.
- Users can only view and update their own tickets (enforced at the DB query level with `user_id`).
- `closed` is a terminal status; no re-opening is allowed.
- `JWT_SECRET` must be set as an environment variable.
- `DB_PATH` is optional and defaults to `./tickets.db`.
