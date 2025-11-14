# Panduan Test WebSocket dengan Postman - Step by Step

## Prerequisites
- Server running: `go run cmd/server/main.go`
- Postman installed (atau gunakan Postman Web)

---

## STEP 1: Buka Postman dan Create Connection Pertama (Alice)

### 1.1 Create WebSocket Request
1. Buka Postman
2. Klik **"New"** → **"WebSocket"** (atau **"WebSocket Request"**)
3. Jika tidak ada, klik dropdown yang tadinya "HTTP" → pilih **"WebSocket"**

### 1.2 Setup Connection Alice
```
URL: ws://localhost:8080/ws?username=Alice
```

**PENTING:** Pastikan:
- Protocol: `ws://` (BUKAN `http://`)
- Port: `8080`
- Path: `/ws`
- Query param: `?username=Alice`

### 1.3 Connect
1. Klik button **"Connect"**
2. Status akan berubah jadi **"Connected"** (hijau)
3. Anda akan melihat message masuk:

**Expected Response:**
```json
{
  "username": "Alice",
  "content": "Alice joined the chat",
  "timestamp": "2025-11-14T14:30:00+07:00",
  "type": "join"
}
```

✅ **Checkpoint 1:** Alice sudah connect dan terima join message

---

## STEP 2: Buka Tab Kedua untuk Bob

### 2.1 JANGAN TUTUP Tab Alice!
**PENTING:** Biarkan tab Alice tetap terbuka dan connected!

### 2.2 Buka Tab Baru di Postman
- Klik **"+"** (New Tab) di Postman
- ATAU: Klik **"New"** → **"WebSocket"** lagi

### 2.3 Setup Connection Bob
```
URL: ws://localhost:8080/ws?username=Bob
```

### 2.4 Connect Bob
1. Klik **"Connect"**
2. Status jadi **"Connected"**

### 2.5 Cek Messages di KEDUA Tab

**Tab Alice akan terima:**
```json
{
  "username": "Bob",
  "content": "Bob joined the chat",
  "timestamp": "2025-11-14T14:31:00+07:00",
  "type": "join"
}
```

**Tab Bob akan terima:**
```json
{
  "username": "Bob",
  "content": "Bob joined the chat",
  "timestamp": "2025-11-14T14:31:00+07:00",
  "type": "join"
}
```

✅ **Checkpoint 2:** Kedua user sudah connected dan melihat join notification

---

## STEP 3: Kirim Message dari Alice ke Bob

### 3.1 Pindah ke Tab Alice
Klik tab Postman yang Alice (pastikan masih connected)

### 3.2 Di Bagian "Message" (bawah)
Ketik JSON berikut:

```json
{
  "username": "Alice",
  "content": "Hello Bob, apa kabar?",
  "type": "message"
}
```

### 3.3 Klik "Send"

### 3.4 Cek Hasil di KEDUA Tab

**Tab Alice (sender) akan terima echo:**
```json
{
  "username": "Alice",
  "content": "Hello Bob, apa kabar?",
  "timestamp": "2025-11-14T14:32:00+07:00",
  "type": "message"
}
```

**Tab Bob (receiver) HARUS terima:**
```json
{
  "username": "Alice",
  "content": "Hello Bob, apa kabar?",
  "timestamp": "2025-11-14T14:32:00+07:00",
  "type": "message"
}
```

✅ **Checkpoint 3:** Bob menerima pesan dari Alice!

---

## STEP 4: Reply dari Bob ke Alice

### 4.1 Pindah ke Tab Bob

### 4.2 Kirim Message dari Bob
```json
{
  "username": "Bob",
  "content": "Halo Alice! Saya baik, terima kasih!",
  "type": "message"
}
```

### 4.3 Klik "Send"

### 4.4 Cek di Tab Alice
Alice HARUS menerima reply dari Bob!

✅ **Checkpoint 4:** 2-way communication berhasil!

---

## Troubleshooting

### Problem 1: "Connection Failed" saat Connect

**Penyebab:**
- Server tidak running
- Salah URL/port

**Solusi:**
```bash
# Cek apakah server running
lsof -ti:8080

# Jika tidak ada output, start server
go run cmd/server/main.go
```

---

### Problem 2: Bob Tidak Menerima Message dari Alice

**Debug Steps:**

#### A. Cek Server Logs
Di terminal tempat server running, Anda harus lihat:
```
2025/11/14 14:30:00 New WebSocket connection: Alice (uuid-xxx)
2025/11/14 14:30:00 Client registered: Alice (uuid-xxx)
2025/11/14 14:30:00 ReadPump: Client Alice (uuid-xxx) started reading
2025/11/14 14:31:00 New WebSocket connection: Bob (uuid-yyy)
2025/11/14 14:31:00 Client registered: Bob (uuid-yyy)
2025/11/14 14:31:00 ReadPump: Client Bob (uuid-yyy) started reading
```

Jika Bob TIDAK muncul di log, berarti Bob tidak connect!

#### B. Pastikan Kedua Tab Masih Connected
- Status di Postman harus **"Connected"** (hijau)
- Jika **"Disconnected"** (merah), klik **"Connect"** lagi

#### C. Cek Message Format
Pastikan JSON valid:
```json
{
  "username": "Alice",
  "content": "Test message",
  "type": "message"
}
```

**JANGAN:**
- ❌ Lupa tanda kutip di value
- ❌ Lupa koma antar field
- ❌ Typo di field name

---

### Problem 3: Join Message Tidak Muncul

**Solusi:**
- Disconnect kedua client
- Restart server
- Connect ulang dengan urutan: Alice dulu, baru Bob

---

## Screenshot Expected Flow

```
┌─────────────────────────────────────────┐
│ TAB 1: Alice (Connected)                │
├─────────────────────────────────────────┤
│ Messages:                               │
│ ← {"username":"Alice",                  │
│    "content":"Alice joined the chat"}   │
│                                         │
│ ← {"username":"Bob",                    │
│    "content":"Bob joined the chat"}     │
│                                         │
│ → {"username":"Alice",                  │
│    "content":"Hello Bob"}               │
│                                         │
│ ← {"username":"Alice",                  │
│    "content":"Hello Bob"}               │
│                                         │
│ ← {"username":"Bob",                    │
│    "content":"Hi Alice!"}               │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ TAB 2: Bob (Connected)                  │
├─────────────────────────────────────────┤
│ Messages:                               │
│ ← {"username":"Bob",                    │
│    "content":"Bob joined the chat"}     │
│                                         │
│ ← {"username":"Alice",                  │
│    "content":"Hello Bob"}               │
│                                         │
│ → {"username":"Bob",                    │
│    "content":"Hi Alice!"}               │
│                                         │
│ ← {"username":"Bob",                    │
│    "content":"Hi Alice!"}               │
└─────────────────────────────────────────┘
```

---

## Tips Testing

1. **Buka 2 tab Postman secara berdampingan** (side-by-side) agar bisa lihat kedua tab sekaligus
2. **Jangan close tab** saat testing - biarkan tetap connected
3. **Perhatikan server logs** di terminal untuk debugging
4. **Test disconnect/reconnect** untuk lihat leave message

---

## Next: Test dengan 3+ Users

Buat tab ke-3 dengan username "Charlie":
```
ws://localhost:8080/ws?username=Charlie
```

Semua user (Alice, Bob, Charlie) akan menerima semua messages!

---

## Jika Masih Gagal

Jalankan command ini di terminal dan paste hasilnya:

```bash
# Cek server logs
go run cmd/server/main.go

# Di terminal lain, cek apakah port listening
lsof -i:8080
```

Lalu screenshot Postman saat error terjadi.
