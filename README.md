# WebSocket Chat Server - Go Backend

WebSocket server sederhana untuk chat real-time dengan best practice Go structure.

## Struktur Project (Clean & Simple)

```
go-websocket/
├── cmd/server/
│   └── main.go              # Entry point aplikasi
├── internal/
│   ├── hub/
│   │   └── hub.go           # Hub untuk manage clients & broadcasting
│   ├── models/
│   │   ├── client.go        # Client model & methods (ReadPump, WritePump)
│   │   └── message.go       # Message struct
│   └── handlers/
│       └── websocket.go     # WebSocket handler
├── go.mod
└── README.md
```

## Quick Start

### 1. Install Dependencies
```bash
go mod download
```

### 2. Run Server
```bash
go run cmd/server/main.go
```

Server akan running di: `http://localhost:8080`

### 3. Test dengan Postman

**WebSocket Request #1 (User Alice):**
```
URL: ws://localhost:8080/ws?username=Alice
Method: WebSocket
Click: Connect
```

**WebSocket Request #2 (User Bob) - Tab Baru:**
```
URL: ws://localhost:8080/ws?username=Bob
Method: WebSocket
Click: Connect
```

**Kirim Message dari Alice (Format Minimal):**
```json
{
  "username": "Alice",
  "content": "Hello Bob!"
}
```

Server otomatis menambahkan `timestamp` dan `type: "message"`.

Bob akan menerima message tersebut! ✅

## Arsitektur Backend

### 1. Hub Pattern ([internal/hub/hub.go](internal/hub/hub.go))

Hub adalah central manager yang mengelola semua WebSocket connections:

```go
type Hub struct {
    clients    map[*models.Client]bool  // Semua connected clients
    broadcast  chan []byte              // Channel untuk broadcast messages
    register   chan *models.Client      // Register client baru
    unregister chan *models.Client      // Unregister client
}
```

**Cara Kerja:**
- Hub running dalam goroutine terpisah
- Menggunakan channels untuk thread-safe communication
- Broadcast message ke semua connected clients

### 2. Client Goroutines Pattern ([internal/models/client.go](internal/models/client.go))

Setiap client memiliki **2 goroutines**:

**ReadPump** - Membaca dari WebSocket:
```go
func (c *Client) ReadPump(hub, unregister)
  - Baca message dari WebSocket
  - Kirim ke Hub untuk broadcast
  - Handle disconnect
```

**WritePump** - Menulis ke WebSocket:
```go
func (c *Client) WritePump()
  - Tunggu message dari Hub
  - Kirim ke WebSocket client
  - Kirim ping setiap 54 detik (keep-alive)
```

### 3. Message Flow

```
Client A → WebSocket → ReadPump → Hub.Broadcast() → WritePump → WebSocket → Client B
                                    ↓
                                WritePump → WebSocket → Client C
```

## API Endpoints

### WebSocket Endpoint

**URL:** `/ws`

**Query Parameters:**
- `username` (string) - Username client

**Connection:** Upgrade ke WebSocket

**Message Format (Send & Receive):**
```json
{
  "username": "string",
  "content": "string",
  "timestamp": "2025-11-14T10:00:00Z",
  "type": "message|join|leave"
}
```

**Message Types:**
- `join` - User bergabung (auto dari server)
- `leave` - User keluar (auto dari server)
- `message` - Chat message

## Configuration

### Timeouts & Limits ([internal/models/client.go](internal/models/client.go))

```go
writeWait      = 10 seconds   // Timeout untuk write
pongWait       = 60 seconds   // Timeout untuk pong
pingPeriod     = 54 seconds   // Interval ping
maxMessageSize = 8192 bytes   // Max message size
```

## Best Practices yang Diimplementasi

✅ **Separation of Concerns** - Models, Handlers, Business Logic terpisah
✅ **Hub Pattern** - Central manager untuk client connections
✅ **Goroutine per Client** - Scalable concurrent handling
✅ **Channel-based Communication** - Thread-safe messaging
✅ **Ping/Pong Keep-Alive** - Detect dead connections
✅ **Graceful Shutdown** - Proper cleanup saat disconnect
✅ **Structured Logging** - Debug-friendly logs

## Production Checklist

Untuk production deployment, tambahkan:

- [ ] **Authentication** - JWT/Session-based auth
- [ ] **Authorization** - Permission management
- [ ] **CORS Configuration** - Set proper allowed origins
- [ ] **Rate Limiting** - Prevent spam/abuse
- [ ] **Message Persistence** - Database (PostgreSQL/MongoDB)
- [ ] **Redis Pub/Sub** - Multi-server scaling
- [ ] **Monitoring** - Prometheus metrics
- [ ] **TLS/SSL** - HTTPS & WSS
- [ ] **Docker** - Containerization
- [ ] **Unit Tests** - Test coverage
- [ ] **Load Testing** - Performance benchmarks

## Scaling untuk Production

### Single Server → Multiple Servers

Gunakan **Redis Pub/Sub**:

```
Client A → Server 1 → Redis Pub/Sub → Server 2 → Client B
```

### Database Integration

Simpan chat history:

```go
// Setelah broadcast, simpan ke DB
db.SaveMessage(ctx, message)
```

### Add Rooms/Channels

```go
type Hub struct {
    rooms map[string]*Room
}

type Room struct {
    ID      string
    clients map[*Client]bool
}
```

## Stop Server

```bash
# Find & kill process
lsof -ti:8080 | xargs kill -9
```

## License

MIT
# go-websocket
