# WhatsApp Sender Registration API

This document explains how to register new WhatsApp senders using the API instead of command-line tools.

## Overview

Your system supports sender registration through API endpoints with **FIXED QR CODE ISSUES**! 

**✅ Recent Fixes**:
- QR codes are now guaranteed to be valid before API responds
- Added 5-second timeout with proper error handling
- Status endpoint returns refreshed QR codes when they expire
- Comprehensive test page included

You have two methods:

1. **QR Code Method** - Scan a QR code with WhatsApp
2. **Pairing Code Method** - Enter a code in WhatsApp (recommended for web apps)

## API Endpoints

### 1. Register Sender with QR Code

**Endpoint:** `POST /api/register-sender-qr`

**Authentication:** Basic Auth (API_USERNAME / API_PASSWORD)

**Request:**
```bash
curl -X POST http://localhost:8080/api/register-sender-qr \
  -u admin:your_password \
  -H "Content-Type: application/json"
```

**Response:**
```json
{
  "success": true,
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "qr_code": "iVBORw0KGgoAAAANSUhEUgAA...",  // Base64 encoded PNG image
  "message": "QR code generated. Please scan with WhatsApp."
}
```

**Usage:**
1. Call the endpoint
2. Get the base64-encoded QR code from `qr_code` field
3. Display it in your web app: `<img src="data:image/png;base64,{qr_code}" />`
4. User scans with WhatsApp
5. Poll the status endpoint to check completion

---

### 2. Register Sender with Pairing Code

**Endpoint:** `POST /api/register-sender-code`

**Authentication:** Basic Auth (API_USERNAME / API_PASSWORD)

**Request:**
```bash
curl -X POST http://localhost:8080/api/register-sender-code \
  -u admin:your_password \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+1234567890"
  }'
```

**Response:**
```json
{
  "success": true,
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "pairing_code": "ABC-DEF-123",
  "phone_number": "+1234567890",
  "message": "Pairing code generated. Please enter it in WhatsApp."
}
```

**Usage:**
1. Call the endpoint with the phone number
2. Get the pairing code from `pairing_code` field
3. Display it to the user
4. User enters the code in WhatsApp:
   - Settings → Linked Devices → Link a Device
   - Tap "Link with phone number instead"
   - Enter the pairing code
5. Poll the status endpoint to check completion

---

### 3. Check Registration Status

**Endpoint:** `GET /api/register-sender-status/:sessionId`

**Authentication:** Basic Auth (API_USERNAME / API_PASSWORD)

**Request:**
```bash
curl -X GET http://localhost:8080/api/register-sender-status/550e8400-e29b-41d4-a716-446655440000 \
  -u admin:your_password
```

**Response (Pending):**
```json
{
  "success": true,
  "status": "pending",
  "message": "Waiting for user to complete pairing..."
}
```

**Response (Connected):**
```json
{
  "success": true,
  "status": "connected",
  "sender_id": "1234567890",
  "message": "Successfully registered!"
}
```

**Response (Failed):**
```json
{
  "success": false,
  "status": "failed",
  "message": "Registration failed or timed out"
}
```

---

## Complete Workflow Examples

### Example 1: QR Code Registration Flow

```javascript
// Step 1: Start QR registration
const startResponse = await fetch('http://localhost:8080/api/register-sender-qr', {
  method: 'POST',
  headers: {
    'Authorization': 'Basic ' + btoa('admin:your_password'),
    'Content-Type': 'application/json'
  }
});

const { session_id, qr_code } = await startResponse.json();

// Step 2: Display QR code
document.getElementById('qrImage').src = `data:image/png;base64,${qr_code}`;

// Step 3: Poll for status
const pollStatus = setInterval(async () => {
  const statusResponse = await fetch(
    `http://localhost:8080/api/register-sender-status/${session_id}`,
    {
      headers: {
        'Authorization': 'Basic ' + btoa('admin:your_password')
      }
    }
  );
  
  const status = await statusResponse.json();
  
  if (status.status === 'connected') {
    clearInterval(pollStatus);
    alert(`Success! Sender ID: ${status.sender_id}`);
  } else if (status.status === 'failed') {
    clearInterval(pollStatus);
    alert('Registration failed');
  }
}, 2000); // Poll every 2 seconds
```

### Example 2: Pairing Code Registration Flow

```javascript
// Step 1: Start code registration
const startResponse = await fetch('http://localhost:8080/api/register-sender-code', {
  method: 'POST',
  headers: {
    'Authorization': 'Basic ' + btoa('admin:your_password'),
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    phone_number: '+1234567890'
  })
});

const { session_id, pairing_code } = await startResponse.json();

// Step 2: Display pairing code to user
alert(`Enter this code in WhatsApp: ${pairing_code}`);

// Step 3: Poll for status (same as QR code example)
const pollStatus = setInterval(async () => {
  const statusResponse = await fetch(
    `http://localhost:8080/api/register-sender-status/${session_id}`,
    {
      headers: {
        'Authorization': 'Basic ' + btoa('admin:your_password')
      }
    }
  );
  
  const status = await statusResponse.json();
  
  if (status.status === 'connected') {
    clearInterval(pollStatus);
    alert(`Success! Sender ID: ${status.sender_id}`);
  } else if (status.status === 'failed') {
    clearInterval(pollStatus);
    alert('Registration failed');
  }
}, 2000);
```

---

## Python Example

```python
import requests
import base64
import time
from io import BytesIO
from PIL import Image

# Configuration
BASE_URL = "http://localhost:8080"
USERNAME = "admin"
PASSWORD = "your_password"

def register_with_qr():
    # Start QR registration
    response = requests.post(
        f"{BASE_URL}/api/register-sender-qr",
        auth=(USERNAME, PASSWORD)
    )
    
    data = response.json()
    session_id = data['session_id']
    qr_code_base64 = data['qr_code']
    
    # Decode and display QR code
    qr_bytes = base64.b64decode(qr_code_base64)
    image = Image.open(BytesIO(qr_bytes))
    image.show()
    
    # Poll for status
    while True:
        status_response = requests.get(
            f"{BASE_URL}/api/register-sender-status/{session_id}",
            auth=(USERNAME, PASSWORD)
        )
        status_data = status_response.json()
        
        if status_data['status'] == 'connected':
            print(f"✓ Success! Sender ID: {status_data['sender_id']}")
            break
        elif status_data['status'] == 'failed':
            print("✗ Registration failed")
            break
        else:
            print("⏳ Waiting for QR scan...")
            time.sleep(2)

def register_with_code(phone_number):
    # Start code registration
    response = requests.post(
        f"{BASE_URL}/api/register-sender-code",
        auth=(USERNAME, PASSWORD),
        json={'phone_number': phone_number}
    )
    
    data = response.json()
    session_id = data['session_id']
    pairing_code = data['pairing_code']
    
    print(f"Enter this code in WhatsApp: {pairing_code}")
    
    # Poll for status
    while True:
        status_response = requests.get(
            f"{BASE_URL}/api/register-sender-status/{session_id}",
            auth=(USERNAME, PASSWORD)
        )
        status_data = status_response.json()
        
        if status_data['status'] == 'connected':
            print(f"✓ Success! Sender ID: {status_data['sender_id']}")
            break
        elif status_data['status'] == 'failed':
            print("✗ Registration failed")
            break
        else:
            print("⏳ Waiting for code entry...")
            time.sleep(2)

# Use QR code method
register_with_qr()

# Or use pairing code method
# register_with_code("+1234567890")
```

---

## cURL Examples

### QR Code Registration
```bash
# Step 1: Start registration
SESSION_ID=$(curl -X POST http://localhost:8080/api/register-sender-qr \
  -u admin:your_password \
  -H "Content-Type: application/json" \
  -s | jq -r '.session_id')

echo "Session ID: $SESSION_ID"

# Step 2: Check status (run multiple times)
curl -X GET http://localhost:8080/api/register-sender-status/$SESSION_ID \
  -u admin:your_password \
  -s | jq '.'
```

### Pairing Code Registration
```bash
# Step 1: Start registration
curl -X POST http://localhost:8080/api/register-sender-code \
  -u admin:your_password \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+1234567890"}' \
  -s | jq '.'

# Copy the session_id from response

# Step 2: Check status
curl -X GET http://localhost:8080/api/register-sender-status/YOUR_SESSION_ID \
  -u admin:your_password \
  -s | jq '.'
```

---

## Important Notes

1. **Sessions expire after 10 minutes** - Complete registration within this time
2. **QR codes refresh** - The QR code may update if not scanned quickly
3. **One session at a time** - Each registration creates a new session
4. **Poll every 2-3 seconds** - Don't poll too frequently
5. **Clean phone numbers** - Use format: `+1234567890` (country code required)

---

## Comparison: CLI vs API

| Feature | CLI (`./whatspoints -add-sender`) | API (`POST /api/register-sender-qr`) |
|---------|-----------------------------------|--------------------------------------|
| **Interface** | Terminal | HTTP REST API |
| **QR Display** | Terminal (ASCII art) | Base64 PNG image |
| **Automation** | Limited | Full automation possible |
| **Web Integration** | Not possible | Easy integration |
| **Remote Access** | Requires SSH | Works over HTTP |
| **Use Case** | Manual setup | Production web apps |

---

## Testing the API

Start your server:
```bash
./whatspoints
```

Test with curl:
```bash
# Test QR registration
curl -X POST http://localhost:8080/api/register-sender-qr \
  -u admin:your_password \
  -H "Content-Type: application/json"
```

The response will contain a base64-encoded QR code that you can display in any web application!

---

## Next Steps

1. **Build a web UI** for sender registration
2. **Add to existing web pages** in the `web/` directory
3. **Create a management dashboard** for multiple senders
4. **Implement session cleanup** if needed
5. **Add webhooks** for registration completion notifications
