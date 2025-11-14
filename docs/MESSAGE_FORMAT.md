# Message Format - WebSocket Chat

## Request Format (Client → Server)

### Minimal Format (Recommended)
Cukup kirim `username` dan `content`:

```json
{
  "username": "Tim",
  "content": "Hello from Postman!"
}
```

Server akan otomatis menambahkan:
- `timestamp`: Waktu server saat ini
- `type`: "message"

---

### Full Format (Optional)
Jika Anda ingin set semua field:

```json
{
  "username": "Tim",
  "content": "Hello from Postman!",
  "type": "message",
  "timestamp": "2025-11-14T14:30:00+07:00"
}
```

Tapi **TIDAK PERLU** karena server akan handle otomatis!

---

## Response Format (Server → Client)

Server akan selalu kirim format lengkap:

```json
{
  "username": "Tim",
  "content": "Hello from Postman!",
  "timestamp": "2025-11-14T14:43:05.848892+07:00",
  "type": "message"
}
```

---

## Message Types

### 1. Chat Message (type: "message")
**Client kirim:**
```json
{
  "username": "Tim",
  "content": "Apa kabar?"
}
```

**Server broadcast:**
```json
{
  "username": "Tim",
  "content": "Apa kabar?",
  "timestamp": "2025-11-14T14:30:00+07:00",
  "type": "message"
}
```

---

### 2. Join Notification (type: "join")
**OTOMATIS dari server** saat user connect:

```json
{
  "username": "Tim",
  "content": "Tim joined the chat",
  "timestamp": "2025-11-14T14:30:00+07:00",
  "type": "join"
}
```

❌ **JANGAN kirim manual dari client!**

---

### 3. Leave Notification (type: "leave")
**OTOMATIS dari server** saat user disconnect:

```json
{
  "username": "Tim",
  "content": "Tim left the chat",
  "timestamp": "2025-11-14T14:35:00+07:00",
  "type": "leave"
}
```

❌ **JANGAN kirim manual dari client!**

---

## Field Descriptions

| Field | Required | Auto-generated | Description |
|-------|----------|----------------|-------------|
| `username` | ✅ Yes | ❌ No | Username pengirim |
| `content` | ✅ Yes | ❌ No | Isi pesan |
| `type` | ❌ No | ✅ Yes | Type: "message", "join", "leave" |
| `timestamp` | ❌ No | ✅ Yes | ISO 8601 timestamp |

---

## Examples

### Example 1: Send Simple Message
**Request:**
```json
{
  "username": "Alice",
  "content": "Halo semua!"
}
```

**Broadcast to all clients:**
```json
{
  "username": "Alice",
  "content": "Halo semua!",
  "timestamp": "2025-11-14T14:30:00+07:00",
  "type": "message"
}
```

---

### Example 2: User Connect
**Action:** User "Bob" membuka connection

**Server automatically broadcasts:**
```json
{
  "username": "Bob",
  "content": "Bob joined the chat",
  "timestamp": "2025-11-14T14:31:00+07:00",
  "type": "join"
}
```

---

### Example 3: User Disconnect
**Action:** User "Bob" menutup connection

**Server automatically broadcasts:**
```json
{
  "username": "Bob",
  "content": "Bob left the chat",
  "timestamp": "2025-11-14T14:35:00+07:00",
  "type": "leave"
}
```

---

## Validation Rules

Server akan:
- ✅ Set `timestamp` otomatis jika kosong
- ✅ Set `type` ke "message" jika kosong
- ✅ Broadcast ke semua connected clients
- ✅ Include sender (sender menerima echo)

---

## Testing dengan Postman

### Test 1: Minimal Message
```json
{
  "username": "TestUser",
  "content": "Test message"
}
```
Click **Send** → Semua clients akan terima message lengkap dengan timestamp & type

### Test 2: Message dengan Emoji
```json
{
  "username": "TestUser",
  "content": "Hello! 👋🎉"
}
```

### Test 3: Long Message
```json
{
  "username": "TestUser",
  "content": "This is a very long message that contains multiple sentences. WebSocket can handle messages up to 8192 bytes by default."
}
```

---

## Error Cases

### Invalid JSON
```json
{
  "username": "Test"
  "content": "Missing comma"  ← Error!
}
```
Server akan log error dan skip message.

### Missing Required Fields
```json
{
  "username": "Test"
  // missing "content"
}
```
Message akan dikirim tapi `content` akan kosong.

---

## Best Practices

✅ **DO:**
- Kirim minimal format: `username` + `content`
- Biarkan server handle `timestamp` dan `type`
- Gunakan JSON valid

❌ **DON'T:**
- Jangan kirim `type: "join"` atau `type: "leave"` manual
- Jangan hardcode timestamp dari client
- Jangan kirim message tanpa `username`

---

## Code Reference

- Message struct: [internal/models/message.go](internal/models/message.go)
- Auto-fill logic: [internal/hub/hub.go:82-90](internal/hub/hub.go#L82-L90)
- Broadcast logic: [internal/hub/hub.go:99-107](internal/hub/hub.go#L99-L107)
