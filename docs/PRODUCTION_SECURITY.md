# Production Security - Username Injection

## 🔒 Security Improvement: Server-Side Username Injection

### Problem Sebelumnya (Insecure)

**Client bisa spoofing username!**

```json
// Alice connect dengan username "Alice"
ws://localhost:8080/ws?username=Alice

// Tapi Alice bisa kirim message dengan username "Bob"!
{
  "username": "Bob",      ← SPOOFING!
  "content": "I'm Bob!"   ← Alice pretending to be Bob
}
```

❌ **Masalah:**
- Client tidak terpercaya
- Bisa impersonate user lain
- Tidak ada validasi username
- Security hole besar!

---

## ✅ Solusi: Server Override Username

### Flow Baru (Secure)

1. **Client Connect**
```
ws://localhost:8080/ws?username=Alice
```
Server save: `client.Username = "Alice"`

2. **Client Kirim Message (TANPA username)**
```json
{
  "content": "Hello!"
}
```

3. **Server Otomatis Inject Username**
```go
// Di hub.BroadcastFromClient()
msg.Username = client.Username  // ALWAYS dari server session!
```

4. **Broadcast dengan Username yang Benar**
```json
{
  "username": "Alice",  ← Dari server, bukan dari client!
  "content": "Hello!",
  "timestamp": "2025-11-14T...",
  "type": "message"
}
```

---

## 📝 Format Request Baru

### ✅ Format BARU (Production-Ready)

**Kirim message TANPA username:**
```json
{
  "content": "Hello everyone!"
}
```

Server akan otomatis inject:
- `username` → dari `client.Username` (session)
- `timestamp` → waktu server
- `type` → "message" (default)

---

### ⚠️ Format LAMA (Masih Supported, Tapi Username Di-Override)

**Jika Anda tetap kirim username:**
```json
{
  "username": "Bob",      ← IGNORED!
  "content": "Hello!"
}
```

**Server akan override:**
```go
msg.Username = client.Username  // "Alice", bukan "Bob"!
```

**Result:**
```json
{
  "username": "Alice",    ← Server override dengan session username
  "content": "Hello!",
  "timestamp": "2025-11-14T...",
  "type": "message"
}
```

---

## 🛡️ Security Benefits

### 1. **Prevent Username Spoofing**
```
❌ BEFORE: Client bisa claim sebagai user lain
✅ AFTER:  Server enforce username dari session
```

### 2. **Single Source of Truth**
```
❌ BEFORE: Username di 2 tempat (connection + message body)
✅ AFTER:  Username HANYA di connection (server-side)
```

### 3. **Consistency**
```
❌ BEFORE: Username bisa berbeda per message
✅ AFTER:  Username always sama selama session
```

### 4. **Audit Trail**
```
✅ Server log PASTI akurat karena username di-enforce server-side
```

---

## 🔧 Implementation Details

### Code Changes

**1. Client.ReadPump** ([internal/models/client.go:50](internal/models/client.go#L50))
```go
// OLD: Broadcast raw message
hub.Broadcast(message)

// NEW: Broadcast with client context
hub.BroadcastFromClient(message, c)
```

**2. Hub.BroadcastFromClient** ([internal/hub/hub.go:117-150](internal/hub/hub.go#L117-L150))
```go
func (h *Hub) BroadcastFromClient(message []byte, client *models.Client) {
    var msg models.Message
    json.Unmarshal(message, &msg)

    // SECURITY: Override username from session
    msg.Username = client.Username  ← KEY LINE!

    // Auto-fill other fields
    if msg.Timestamp.IsZero() {
        msg.Timestamp = time.Now()
    }
    if msg.Type == "" {
        msg.Type = "message"
    }

    // Broadcast to all
    h.broadcast <- msgJSON
}
```

---

## 📊 Comparison

### Before (Insecure)

```
Client                         Server
  |                              |
  | Connect (username=Alice)     |
  |----------------------------->|
  |                              | Store: client.Username = "Alice"
  |                              |
  | Send: {username:"Bob",       |
  |        content:"Hello"}      |
  |----------------------------->|
  |                              | Broadcast as "Bob" ← WRONG!
```

### After (Secure)

```
Client                         Server
  |                              |
  | Connect (username=Alice)     |
  |----------------------------->|
  |                              | Store: client.Username = "Alice"
  |                              |
  | Send: {content:"Hello"}      |
  |----------------------------->|
  |                              | Inject: msg.Username = "Alice"
  |                              | Broadcast as "Alice" ← CORRECT!
```

---

## 🚀 Testing

### Test 1: Normal Message
```
Connect: ws://localhost:8080/ws?username=Alice

Send:
{
  "content": "Hello!"
}

Receive:
{
  "username": "Alice",
  "content": "Hello!",
  "timestamp": "2025-11-14T...",
  "type": "message"
}
```

### Test 2: Spoofing Attempt (Should Fail)
```
Connect: ws://localhost:8080/ws?username=Alice

Send:
{
  "username": "Bob",    ← Trying to spoof
  "content": "I'm Bob!"
}

Receive:
{
  "username": "Alice",  ← Server enforces real username!
  "content": "I'm Bob!",
  "timestamp": "2025-11-14T...",
  "type": "message"
}
```

✅ **Spoofing failed! Username is enforced by server.**

---

## 📝 Migration Guide

### Untuk Existing Clients

**Option 1: Remove Username (Recommended)**
```diff
- {
-   "username": "Alice",
-   "content": "Hello"
- }

+ {
+   "content": "Hello"
+ }
```

**Option 2: Keep Username (Masih Works, Tapi Di-Override)**
```json
{
  "username": "Alice",  // This will be ignored
  "content": "Hello"
}
```
Server akan override `username` dengan session username.

---

## 🎯 Best Practices

### ✅ DO:
- Trust server untuk username
- Kirim minimal payload: `{content: "..."}`
- Validate username di server-side
- Use session/token untuk authentication

### ❌ DON'T:
- Jangan trust username dari client
- Jangan validasi username di client-side saja
- Jangan assume client jujur

---

## 🔐 Next Steps for Production

Untuk production-ready, tambahkan:

1. **Authentication**
   ```
   ws://localhost:8080/ws?token=JWT_TOKEN
   Server decode JWT → get real username
   ```

2. **Authorization**
   ```go
   // Check if user has permission to send message
   if !client.HasPermission("send_message") {
       return errors.New("unauthorized")
   }
   ```

3. **Rate Limiting**
   ```go
   // Prevent spam
   if client.MessageCount > 10 per minute {
       return errors.New("rate limit exceeded")
   }
   ```

4. **Input Validation**
   ```go
   // Validate content
   if len(msg.Content) > 1000 {
       return errors.New("message too long")
   }
   ```

---

## 📚 Related Files

- [internal/hub/hub.go](internal/hub/hub.go) - BroadcastFromClient implementation
- [internal/models/client.go](internal/models/client.go) - ReadPump changes
- [MESSAGE_FORMAT.md](MESSAGE_FORMAT.md) - Updated message format
