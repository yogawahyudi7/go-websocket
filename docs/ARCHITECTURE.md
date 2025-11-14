# Architecture Documentation

## System Overview

WebSocket Chat Server menggunakan **Hub Pattern** untuk central message management dengan **Goroutine-per-Client** pattern untuk concurrency.

---

## 📐 High-Level Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                     Load Balancer (Future)                   │
└────────────────────────┬─────────────────────────────────────┘
                         │
┌────────────────────────┼─────────────────────────────────────┐
│                   Go Server                                   │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐  │
│  │                  HTTP Server                           │  │
│  │                  (port 8080)                           │  │
│  └───────────────────────┬────────────────────────────────┘  │
│                          │                                    │
│                          ▼                                    │
│  ┌────────────────────────────────────────────────────────┐  │
│  │            WebSocket Handler                           │  │
│  │         (internal/handlers/websocket.go)               │  │
│  │                                                        │  │
│  │  • Upgrade HTTP → WebSocket                           │  │
│  │  • Create Client instance                             │  │
│  │  • Register to Hub                                    │  │
│  │  • Spawn goroutines                                   │  │
│  └───────────────────────┬────────────────────────────────┘  │
│                          │                                    │
│                          ▼                                    │
│  ┌────────────────────────────────────────────────────────┐  │
│  │                   Hub                                  │  │
│  │            (internal/hub/hub.go)                       │  │
│  │                                                        │  │
│  │  Channels:                                            │  │
│  │  • register   chan *Client                            │  │
│  │  • unregister chan *Client                            │  │
│  │  • broadcast  chan []byte                             │  │
│  │                                                        │  │
│  │  Clients Map:                                         │  │
│  │  • map[*Client]bool                                   │  │
│  │                                                        │  │
│  │  Methods:                                             │  │
│  │  • Run() - Main event loop                            │  │
│  │  • BroadcastFromClient() - Send to all               │  │
│  └───────────────────────┬────────────────────────────────┘  │
│                          │                                    │
│                          │                                    │
│         ┌────────────────┼────────────────┐                  │
│         │                │                │                  │
│         ▼                ▼                ▼                  │
│  ┌──────────┐     ┌──────────┐     ┌──────────┐            │
│  │ Client 1 │     │ Client 2 │     │ Client 3 │            │
│  │ (Alice)  │     │  (Bob)   │     │ (Carol)  │            │
│  │          │     │          │     │          │            │
│  │ ReadPump │     │ ReadPump │     │ ReadPump │            │
│  │ WritePump│     │ WritePump│     │ WritePump│            │
│  └────┬─────┘     └────┬─────┘     └────┬─────┘            │
│       │                │                │                    │
└───────┼────────────────┼────────────────┼────────────────────┘
        │                │                │
        ▼                ▼                ▼
   WebSocket        WebSocket        WebSocket
        │                │                │
        ▼                ▼                ▼
    Browser          Browser          Browser
```

---

## 🔄 Component Interaction Flow

### 1. Connection Flow

```
[Browser]
    │
    │ HTTP GET /ws?username=Alice
    │ Upgrade: websocket
    ▼
[HTTP Server]
    │
    │ Route to WebSocketHandler
    ▼
[WebSocketHandler]
    │
    ├─ Upgrade HTTP → WebSocket
    ├─ Create Client(conn, id, username)
    ├─ hub.Register(client)
    ├─ go client.WritePump()
    └─ go client.ReadPump()
         │
         ▼
[Hub.Run()] receives on register channel
    │
    ├─ Add client to clients map
    ├─ Log: "Client registered: Alice"
    └─ Broadcast join message
```

### 2. Message Flow

```
[Alice Browser]
    │
    │ Send: {"content": "Hello"}
    ▼
[WebSocket Connection]
    │
    ▼
[Client.ReadPump()] - Alice's goroutine
    │
    ├─ Read from WebSocket
    ├─ Parse JSON
    │
    ▼
[Hub.BroadcastFromClient()]
    │
    ├─ Inject username (security)
    │   msg.Username = "Alice"
    ├─ Set timestamp & type
    ├─ Marshal to JSON
    │
    └─ Send to broadcast channel
         │
         ▼
[Hub.Run()] receives on broadcast channel
    │
    └─ Loop through all clients:
         │
         ├─ client1.Send <- msgJSON  (Alice)
         ├─ client2.Send <- msgJSON  (Bob)
         └─ client3.Send <- msgJSON  (Carol)
              │
              ▼
[Client.WritePump()] - Each client's goroutine
    │
    ├─ Receive from Send channel
    ├─ Write to WebSocket
    │
    ▼
[Browser] - All connected users receive message
```

### 3. Disconnect Flow

```
[Browser] Close connection
    │
    ▼
[Client.ReadPump()] detects error
    │
    ├─ defer func() runs
    ├─ Send to unregister channel
    └─ Close WebSocket
         │
         ▼
[Hub.Run()] receives on unregister channel
    │
    ├─ Delete from clients map
    ├─ Close client.Send channel
    ├─ Log: "Client unregistered: Alice"
    └─ Broadcast leave message
```

---

## 🧩 Component Details

### 1. Hub (Central Broker)

**File:** `internal/hub/hub.go`

**Responsibilities:**
- Maintain list of connected clients
- Route messages between clients
- Handle client registration/unregistration
- Broadcast messages to all clients

**Key Data Structures:**
```go
type Hub struct {
    clients    map[*models.Client]bool  // Active clients
    broadcast  chan []byte              // Message queue
    register   chan *models.Client      // Registration queue
    unregister chan *models.Client      // Unregistration queue
}
```

**Main Loop:**
```go
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            // Handle new client
        case client := <-h.unregister:
            // Handle disconnect
        case message := <-h.broadcast:
            // Broadcast to all clients
        }
    }
}
```

---

### 2. Client (Connection Wrapper)

**File:** `internal/models/client.go`

**Responsibilities:**
- Wrap WebSocket connection
- Read messages (ReadPump)
- Write messages (WritePump)
- Handle ping/pong

**Key Data Structures:**
```go
type Client struct {
    Conn     *websocket.Conn  // WebSocket connection
    Send     chan []byte      // Outbound message queue
    ID       string           // Unique client ID
    Username string           // Session username
}
```

**Two Goroutines:**

**ReadPump:**
```go
func (c *Client) ReadPump(hub, unregister) {
    for {
        _, message, err := c.Conn.ReadMessage()
        // Parse & forward to Hub
        hub.BroadcastFromClient(message, c)
    }
}
```

**WritePump:**
```go
func (c *Client) WritePump() {
    ticker := time.NewTicker(pingPeriod)
    for {
        select {
        case message := <-c.Send:
            // Write to WebSocket
        case <-ticker.C:
            // Send ping
        }
    }
}
```

---

### 3. Message Model

**File:** `internal/models/message.go`

**Structure:**
```go
type Message struct {
    MessageID string    `json:"messageId,omitempty"`
    Username  string    `json:"username"`
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
    Type      string    `json:"type"` // "message"|"join"|"leave"
}
```

---

### 4. WebSocket Handler

**File:** `internal/handlers/websocket.go`

**Responsibilities:**
- Upgrade HTTP to WebSocket
- Extract username from query params
- Create Client instance
- Register to Hub
- Start goroutines

**Flow:**
```go
func WebSocketHandler(h *hub.Hub) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Get username
        username := r.URL.Query().Get("username")

        // 2. Upgrade connection
        conn, err := upgrader.Upgrade(w, r, nil)

        // 3. Create client
        client := models.NewClient(conn, id, username)

        // 4. Register
        h.Register(client)

        // 5. Start goroutines
        go client.WritePump()
        go client.ReadPump(h, h.GetUnregisterChan())
    }
}
```

---

## 🔐 Security Architecture

### Username Injection

```
Client Request                    Server Processing
    │                                   │
    │ {                                 │
    │   "content": "Hello"              │
    │ }                                 │
    │                                   │
    ├──────────────────────────────────>│
                                        │
                                        ▼
                            [Hub.BroadcastFromClient]
                                        │
                                        │ msg.Username = client.Username
                                        │ ← From session, NOT from request!
                                        │
                                        ▼
                            {
                              "username": "Alice",  ← Injected
                              "content": "Hello",
                              "timestamp": "...",
                              "type": "message"
                            }
```

---

## 📊 Concurrency Model

### Goroutine Hierarchy

```
Main Goroutine
    │
    ├─ Hub.Run() goroutine (1x)
    │    └─ Event loop (infinite)
    │
    └─ For each Client (N clients):
         ├─ ReadPump goroutine
         └─ WritePump goroutine

Total goroutines: 1 + (2 * N clients)
```

### Thread Safety

**Channels (Thread-Safe):**
- `Hub.register` - Client registration
- `Hub.unregister` - Client removal
- `Hub.broadcast` - Message distribution
- `Client.Send` - Per-client message queue

**Maps (Single-Writer):**
- `Hub.clients` - Only modified in Hub.Run() goroutine

**No Locks Needed!** All coordination via channels.

---

## ⚡ Performance Characteristics

### Memory Usage

Per Client:
```
Client struct:     ~100 bytes
Send channel:      256 * 8 = 2KB (buffer)
Goroutine stack:   2KB (initial)
─────────────────────────────
Total per client:  ~4KB
```

For 10,000 clients: ~40MB

### Message Latency

```
Client A                      Hub                        Client B
   │                           │                            │
   │ Write to WebSocket        │                            │
   ├──────────> 1ms ──────────>│                            │
   │                           │ Process & Route            │
   │                           ├────────> 0.1ms ──────────>│
   │                           │                            │
   │                           │                            │ Read from channel
   │                           │                            ├──> 0.5ms
   │                           │                            │
   │                           │                            │ Write to WebSocket
   │                           │                            ├──> 1ms
                                                            │
Total: ~2.6ms (local network)
```

---

## 🔄 Scalability Patterns

### Single Server (Current)

```
┌─────────────┐
│   Server    │
│             │
│  10K users  │
└─────────────┘
```

Limits: ~10,000 concurrent connections

### Multi-Server with Redis (Future)

```
┌──────────┐    ┌──────────┐    ┌──────────┐
│ Server 1 │    │ Server 2 │    │ Server 3 │
│  3K users│    │  4K users│    │  3K users│
└────┬─────┘    └────┬─────┘    └────┬─────┘
     │               │               │
     └───────────────┼───────────────┘
                     │
              ┌──────▼──────┐
              │ Redis Pub/Sub│
              └─────────────┘
```

Unlimited horizontal scaling!

---

## 📈 Future Enhancements

### 1. Rooms/Channels
```go
type Hub struct {
    rooms map[string]*Room
}

type Room struct {
    ID      string
    clients map[*Client]bool
}
```

### 2. Message Persistence
```go
func (h *Hub) BroadcastFromClient(msg, client) {
    // Current: Just broadcast
    h.broadcast <- msgJSON

    // Future: Save to DB
    go h.saveToDatabase(msg)
}
```

### 3. Authentication
```go
func WebSocketHandler(h *hub.Hub) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Validate JWT token
        token := r.URL.Query().Get("token")
        claims, err := validateJWT(token)
        // ...
    }
}
```

---

## 🧪 Testing Architecture

### Unit Tests
- `hub_test.go` - Hub logic
- `client_test.go` - Client behavior
- `handlers_test.go` - HTTP handlers

### Integration Tests
- End-to-end message flow
- Multi-client scenarios
- Disconnect handling

### Load Tests
- Concurrent connections
- Message throughput
- Memory usage under load

---

## 📚 References

- [Gorilla WebSocket](https://github.com/gorilla/websocket)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [WebSocket RFC 6455](https://tools.ietf.org/html/rfc6455)
