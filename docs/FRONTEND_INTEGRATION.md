# Frontend Integration Guide - WebSocket Chat

## 🎯 Problem: Bagaimana FE Tahu Pesan Berhasil Terkirim?

Saat ini server meng-echo message kembali ke sender, tapi FE perlu cara untuk:
1. ✅ Tahu message mana yang sudah terkirim
2. ✅ Show status: Sending → Sent → Delivered → Read
3. ✅ Handle failures & retries

---

## ✅ Solution: Message ID + Status Tracking

### Pattern yang Digunakan: **Client-Generated Message ID**

```
FE generate UUID → Kirim dengan messageId → Server echo dengan messageId yang sama → FE match & update status
```

---

## 📝 Implementation Guide

### Backend (Already Implemented)

Server sekarang support field `messageId`:

```go
type Message struct {
    MessageID string    `json:"messageId,omitempty"`  // ← NEW!
    Username  string    `json:"username"`
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
    Type      string    `json:"type"`
}
```

Server akan preserve `messageId` yang dikirim client dan echo kembali.

---

## 🚀 Frontend Implementation

### 1. Generate Unique Message ID

```javascript
// Helper function untuk generate UUID
function generateMessageId() {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
    // Atau gunakan library: uuid.v4()
}
```

---

### 2. Send Message dengan Message ID

```javascript
class ChatClient {
    constructor(url, username) {
        this.ws = new WebSocket(url);
        this.username = username;
        this.pendingMessages = new Map(); // Track pending messages

        this.setupHandlers();
    }

    sendMessage(content) {
        // 1. Generate unique ID
        const messageId = generateMessageId();

        // 2. Create message object
        const message = {
            messageId: messageId,
            content: content
            // username akan di-inject oleh server
        };

        // 3. Track as pending
        this.pendingMessages.set(messageId, {
            content: content,
            status: 'sending',
            sentAt: new Date(),
            timeoutId: setTimeout(() => {
                this.handleTimeout(messageId);
            }, 5000) // 5 second timeout
        });

        // 4. Update UI: show as "sending"
        this.displayMessage({
            messageId: messageId,
            username: this.username,
            content: content,
            status: 'sending',
            timestamp: new Date(),
            isMine: true
        });

        // 5. Send to server
        this.ws.send(JSON.stringify(message));

        return messageId;
    }

    setupHandlers() {
        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleIncomingMessage(message);
        };
    }

    handleIncomingMessage(message) {
        // Check if this is an echo of our sent message
        if (this.pendingMessages.has(message.messageId)) {
            // SUCCESS! Our message was received by server
            const pending = this.pendingMessages.get(message.messageId);

            // Clear timeout
            clearTimeout(pending.timeoutId);

            // Update status: sending → sent
            this.updateMessageStatus(message.messageId, 'sent');

            // Remove from pending
            this.pendingMessages.delete(message.messageId);

            console.log(`Message ${message.messageId} confirmed`);
        } else {
            // This is a message from another user
            this.displayMessage({
                ...message,
                isMine: message.username === this.username
            });
        }
    }

    handleTimeout(messageId) {
        const pending = this.pendingMessages.get(messageId);
        if (pending) {
            console.error(`Message ${messageId} timeout`);

            // Update status: sending → failed
            this.updateMessageStatus(messageId, 'failed');

            // Show retry button
            this.showRetryOption(messageId);

            this.pendingMessages.delete(messageId);
        }
    }

    retryMessage(messageId) {
        // Re-send failed message
        const messageElement = document.getElementById(`msg-${messageId}`);
        const content = messageElement.dataset.content;
        this.sendMessage(content);
    }

    updateMessageStatus(messageId, status) {
        const element = document.getElementById(`msg-${messageId}`);
        if (element) {
            element.dataset.status = status;
            element.querySelector('.status').textContent = this.getStatusIcon(status);
        }
    }

    getStatusIcon(status) {
        const icons = {
            'sending': '🕐',   // Clock
            'sent': '✓',       // Single check
            'delivered': '✓✓', // Double check
            'read': '✓✓',      // Blue double check
            'failed': '❌'     // Failed
        };
        return icons[status] || '';
    }

    displayMessage(message) {
        const messageDiv = document.createElement('div');
        messageDiv.id = `msg-${message.messageId}`;
        messageDiv.className = `message ${message.isMine ? 'own' : ''}`;
        messageDiv.dataset.status = message.status || 'sent';
        messageDiv.dataset.content = message.content;

        messageDiv.innerHTML = `
            <div class="username">${message.username}</div>
            <div class="content">${message.content}</div>
            <div class="meta">
                <span class="timestamp">${formatTime(message.timestamp)}</span>
                <span class="status">${this.getStatusIcon(message.status || 'sent')}</span>
            </div>
        `;

        document.getElementById('chat').appendChild(messageDiv);
    }
}
```

---

### 3. Usage Example

```javascript
// Initialize
const chat = new ChatClient('ws://localhost:8080/ws?username=Alice', 'Alice');

// Send message
document.getElementById('sendBtn').onclick = () => {
    const content = document.getElementById('messageInput').value;
    const messageId = chat.sendMessage(content);
    console.log('Sent message:', messageId);
};
```

---

## 📊 Message Status Flow

```
User clicks Send
    ↓
[1] Generate messageId
    ↓
[2] Display with status: "sending" 🕐
    ↓
[3] Send to WebSocket
    ↓
[4] Wait for echo...
    ↓
┌───────────────────────────────┐
│  Timeout (5s)?                │
│  ├─ YES → Status: "failed" ❌ │
│  └─ NO → Continue...          │
└───────────────────────────────┘
    ↓
[5] Receive echo with same messageId
    ↓
[6] Update status: "sent" ✓
    ↓
[7] Remove from pending
```

---

## 🎨 UI Examples

### Message dengan Status (WhatsApp-style)

```html
<div class="message own" data-status="sending">
    <div class="content">Hello!</div>
    <div class="meta">
        <span class="timestamp">14:30</span>
        <span class="status">🕐</span> <!-- Sending -->
    </div>
</div>

<div class="message own" data-status="sent">
    <div class="content">Hello!</div>
    <div class="meta">
        <span class="timestamp">14:30</span>
        <span class="status">✓</span> <!-- Sent -->
    </div>
</div>

<div class="message own" data-status="failed">
    <div class="content">Hello!</div>
    <div class="meta">
        <span class="timestamp">14:30</span>
        <span class="status">❌</span> <!-- Failed -->
        <button class="retry">Retry</button>
    </div>
</div>
```

---

## 🔧 Advanced Features

### 1. Optimistic UI Updates

```javascript
sendMessage(content) {
    const messageId = generateMessageId();

    // Immediately show in UI (optimistic)
    this.displayMessage({
        messageId,
        username: this.username,
        content,
        status: 'sending',
        timestamp: new Date(),
        isMine: true
    });

    // Then send to server
    this.ws.send(JSON.stringify({ messageId, content }));
}
```

### 2. Retry Failed Messages

```javascript
retryMessage(messageId) {
    const element = document.getElementById(`msg-${messageId}`);
    const content = element.dataset.content;

    // Update status back to sending
    this.updateMessageStatus(messageId, 'sending');

    // Re-send with NEW messageId
    this.sendMessage(content);

    // Remove old failed message
    element.remove();
}
```

### 3. Batch Status Updates

```javascript
// Jika banyak pending messages, cek semuanya
checkPendingMessages() {
    const now = Date.now();
    for (const [messageId, pending] of this.pendingMessages) {
        if (now - pending.sentAt > 5000) {
            this.handleTimeout(messageId);
        }
    }
}

// Run every second
setInterval(() => this.checkPendingMessages(), 1000);
```

---

## 📱 React Example

```jsx
import { useState, useRef } from 'react';

function ChatApp() {
    const [messages, setMessages] = useState([]);
    const ws = useRef(null);
    const pendingMessages = useRef(new Map());

    const sendMessage = (content) => {
        const messageId = generateMessageId();
        const tempMessage = {
            messageId,
            username: 'Alice',
            content,
            status: 'sending',
            timestamp: new Date(),
            isMine: true
        };

        // Optimistic update
        setMessages(prev => [...prev, tempMessage]);

        // Track pending
        pendingMessages.current.set(messageId, {
            content,
            timeoutId: setTimeout(() => handleTimeout(messageId), 5000)
        });

        // Send
        ws.current.send(JSON.stringify({ messageId, content }));
    };

    const handleIncomingMessage = (message) => {
        if (pendingMessages.current.has(message.messageId)) {
            // Update existing message status
            setMessages(prev =>
                prev.map(msg =>
                    msg.messageId === message.messageId
                        ? { ...msg, status: 'sent' }
                        : msg
                )
            );

            // Cleanup
            const pending = pendingMessages.current.get(message.messageId);
            clearTimeout(pending.timeoutId);
            pendingMessages.current.delete(message.messageId);
        } else {
            // New message from others
            setMessages(prev => [...prev, { ...message, isMine: false }]);
        }
    };

    return (
        <div className="chat">
            {messages.map(msg => (
                <Message key={msg.messageId} {...msg} />
            ))}
        </div>
    );
}

function Message({ content, status, isMine, timestamp }) {
    const statusIcon = {
        sending: '🕐',
        sent: '✓',
        failed: '❌'
    }[status];

    return (
        <div className={`message ${isMine ? 'own' : ''}`}>
            <div className="content">{content}</div>
            <div className="meta">
                <span>{formatTime(timestamp)}</span>
                <span>{statusIcon}</span>
            </div>
        </div>
    );
}
```

---

## 🧪 Testing dengan Postman

### Test 1: Send dengan messageId

**Send:**
```json
{
  "messageId": "abc123",
  "content": "Hello!"
}
```

**Receive (Echo):**
```json
{
  "messageId": "abc123",
  "username": "Alice",
  "content": "Hello!",
  "timestamp": "2025-11-14T...",
  "type": "message"
}
```

✅ `messageId` preserved! FE dapat match echo dengan message yang dikirim.

---

## 🎯 Best Practices

### ✅ DO:
- Generate `messageId` di client-side
- Track pending messages dengan timeout
- Show immediate feedback (optimistic UI)
- Handle network failures gracefully
- Implement retry mechanism

### ❌ DON'T:
- Jangan rely hanya pada echo untuk confirmation
- Jangan trust server untuk generate messageId
- Jangan ignore timeouts
- Jangan re-send dengan messageId yang sama (gunakan ID baru untuk retry)

---

## 📊 Alternative Patterns

### Pattern 2: Server ACK (More Complex)

Server kirim explicit acknowledgment:

```json
// Client send
{ "messageId": "abc123", "content": "Hello" }

// Server ACK (immediate)
{ "type": "ack", "messageId": "abc123", "status": "received" }

// Server broadcast
{ "messageId": "abc123", "username": "Alice", "content": "Hello", ... }
```

**Pro:** Explicit confirmation
**Con:** More complexity, extra messages

### Pattern 3: Sequence Number

```javascript
let sequenceNumber = 0;

function sendMessage(content) {
    const seq = ++sequenceNumber;
    ws.send(JSON.stringify({ seq, content }));
    // Wait for echo with same seq
}
```

**Pro:** Simple counter
**Con:** Perlu sync sequence number

---

## 📝 Summary

✅ **Dengan messageId pattern:**
- FE dapat track message status (sending/sent/failed)
- User experience lebih baik (WhatsApp-like)
- Dapat implement retry mechanism
- Production-ready approach

**Server sudah support `messageId`** - tinggal implement di FE!

