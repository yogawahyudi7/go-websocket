# Message Status Pattern - Quick Reference

## 🎯 Masalah

**Pertanyaan:** Bagaimana FE tahu pesan berhasil terkirim jika server hanya echo message?

**Jawaban:** Gunakan **Message ID Tracking Pattern**

---

## 📊 Flow Diagram

```
┌─────────────┐
│   Frontend  │
└──────┬──────┘
       │
       │ 1. Generate messageId = "abc123"
       │    Content = "Hello"
       │
       │ 2. Display dengan status "sending" 🕐
       ▼
   [Message]
   Hello         🕐
   14:30
       │
       │ 3. Send ke server
       │    { messageId: "abc123", content: "Hello" }
       │
       ▼
┌──────────────┐
│   Server     │
└──────┬───────┘
       │
       │ 4. Process & Broadcast
       │    (preserve messageId)
       │
       │ 5. Broadcast ke semua clients
       ▼
┌─────────────┐         ┌─────────────┐
│  Frontend   │         │  Others     │
│  (Sender)   │         │             │
└──────┬──────┘         └──────┬──────┘
       │                       │
       │ 6. Receive:           │ 6. Receive:
       │    { messageId:       │    { messageId:
       │      "abc123",        │      "abc123",
       │      username: "Alice"│      username: "Alice"
       │      content: "Hello" │      content: "Hello"
       │    }                  │    }
       │                       │
       │ 7. Check:             │ 7. Check:
       │    pendingMessages    │    NOT in pending
       │    .has("abc123")?    │    → New message!
       │    → YES! ✓           │
       │                       │
       │ 8. Update status:     │ 8. Display:
       │    sending → sent ✓   │    [Message]
       ▼                       │    Alice: Hello
   [Message]                  ▼
   Hello         ✓           [Message]
   14:30                     Alice: Hello
                             14:30
```

---

## 💻 Implementation (Minimal)

### Backend (✅ Already Done)

```go
type Message struct {
    MessageID string `json:"messageId,omitempty"`  // ← Added!
    Username  string `json:"username"`
    Content   string `json:"content"`
    // ...
}
```

Server akan preserve dan echo `messageId`.

---

### Frontend (Simple JavaScript)

```javascript
const pendingMessages = new Map();

// === SEND MESSAGE ===
function sendMessage(content) {
    // 1. Generate ID
    const msgId = Date.now() + '-' + Math.random();

    // 2. Track as pending
    pendingMessages.set(msgId, {
        content: content,
        status: 'sending'
    });

    // 3. Show in UI with "sending" status
    addMessageToUI(msgId, content, 'sending');

    // 4. Send to server
    ws.send(JSON.stringify({
        messageId: msgId,
        content: content
    }));
}

// === RECEIVE MESSAGE ===
ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);

    // Check if this is OUR message (echo)
    if (pendingMessages.has(msg.messageId)) {
        // YES! Update status
        updateMessageStatus(msg.messageId, 'sent');
        pendingMessages.delete(msg.messageId);
    } else {
        // NO! Message from other user
        addMessageToUI(msg.messageId, msg.content, 'received');
    }
};

// === UI HELPERS ===
function addMessageToUI(msgId, content, status) {
    const icon = status === 'sending' ? '🕐' : '✓';
    const html = `
        <div id="msg-${msgId}" data-status="${status}">
            ${content} <span class="status">${icon}</span>
        </div>
    `;
    document.getElementById('chat').innerHTML += html;
}

function updateMessageStatus(msgId, status) {
    const element = document.getElementById(`msg-${msgId}`);
    element.dataset.status = status;
    element.querySelector('.status').textContent = '✓';
}
```

---

## 🧪 Test dengan Postman

### Request (dari FE):
```json
{
  "messageId": "1699999999-abc123",
  "content": "Hello from frontend!"
}
```

### Response (echo dari server):
```json
{
  "messageId": "1699999999-abc123",  ← Same ID!
  "username": "Alice",
  "content": "Hello from frontend!",
  "timestamp": "2025-11-14T14:30:00Z",
  "type": "message"
}
```

✅ FE match `messageId` → Update status dari "sending" → "sent"

---

## 🎨 UI Status Icons

| Status | Icon | Meaning |
|--------|------|---------|
| `sending` | 🕐 | Mengirim ke server |
| `sent` | ✓ | Server sudah terima |
| `delivered` | ✓✓ | User lain sudah terima |
| `read` | 💙✓✓ | User lain sudah baca |
| `failed` | ❌ | Gagal kirim |

---

## ⚡ Optimistic UI (WhatsApp Pattern)

```javascript
function sendMessage(content) {
    const msgId = generateId();

    // 1. LANGSUNG tampilkan (optimistic)
    addMessageToUI(msgId, content, 'sending');

    // 2. THEN kirim ke server
    ws.send(JSON.stringify({ messageId: msgId, content }));

    // 3. Wait for confirmation...
    setTimeout(() => {
        if (pendingMessages.has(msgId)) {
            // Timeout! Mark as failed
            updateMessageStatus(msgId, 'failed');
        }
    }, 5000);
}
```

**Result:** User langsung lihat message-nya (instant feedback)!

---

## 🔄 Retry Failed Messages

```javascript
function retryMessage(msgId) {
    const oldMsg = document.getElementById(`msg-${msgId}`);
    const content = oldMsg.textContent;

    // Remove failed message
    oldMsg.remove();

    // Send again with NEW messageId
    sendMessage(content);
}
```

---

## 📱 React Hook Example

```jsx
function useChat(wsUrl, username) {
    const [messages, setMessages] = useState([]);
    const pendingRef = useRef(new Map());
    const ws = useRef(null);

    const sendMessage = (content) => {
        const msgId = generateId();

        // Optimistic update
        setMessages(prev => [...prev, {
            messageId: msgId,
            content,
            status: 'sending',
            isMine: true
        }]);

        // Track
        pendingRef.current.set(msgId, true);

        // Send
        ws.current.send(JSON.stringify({
            messageId: msgId,
            content
        }));
    };

    const handleMessage = (msg) => {
        if (pendingRef.current.has(msg.messageId)) {
            // Update our message: sending → sent
            setMessages(prev =>
                prev.map(m =>
                    m.messageId === msg.messageId
                        ? { ...m, status: 'sent' }
                        : m
                )
            );
            pendingRef.current.delete(msg.messageId);
        } else {
            // Add new message from others
            setMessages(prev => [...prev, {
                ...msg,
                isMine: false
            }]);
        }
    };

    return { messages, sendMessage };
}
```

---

## ✅ Benefits

1. ✅ **User Experience** - Instant feedback
2. ✅ **Reliability** - Dapat detect failures
3. ✅ **Retry** - Dapat re-send failed messages
4. ✅ **Status Tracking** - WhatsApp-like status
5. ✅ **Production Ready** - Industry standard pattern

---

## 📚 Full Implementation

Lihat dokumentasi lengkap di:
- **[FRONTEND_INTEGRATION.md](FRONTEND_INTEGRATION.md)** - Complete guide dengan React examples

---

## 🎯 Summary

**Q:** Bagaimana FE tahu pesan terkirim?
**A:** Match `messageId` pada echo response!

**Flow:**
```
Send (messageId: abc123) → Pending → Echo (messageId: abc123) → Match! → Sent ✓
```

Server sudah support - tinggal implement di FE! 🚀
