# WhatsApp Sender QR Code Registration Guide

## Overview

The WhatsPoints application now supports registering new WhatsApp senders via QR code scanning through a web interface. This guide explains how the implementation works and how to use it.

## How It Works

### Backend Implementation

1. **Session Management**: Each registration attempt creates a unique session with a session ID
2. **QR Code Generation**: The backend connects to WhatsApp servers and receives QR codes
3. **QR Code Refresh**: WhatsApp QR codes expire every ~20-30 seconds and are automatically refreshed
4. **Status Tracking**: The backend tracks the registration status (pending, connected, failed)
5. **Auto-Update**: The frontend polls the status endpoint every 2 seconds to get the latest QR code

### Frontend Implementation

1. **Method Selection**: Users can choose between QR code or pairing code registration
2. **QR Display**: The QR code is displayed as a base64-encoded PNG image
3. **Auto-Refresh**: The QR code automatically updates when it expires
4. **Status Monitoring**: Real-time status updates show when the device is successfully linked

## Usage Instructions

### For End Users

1. **Access the Registration Page**
   - Navigate to `http://your-server:8080/register`
   - Enter your API password when prompted

2. **Choose QR Code Method**
   - Click on the "QR Code" option

3. **Generate QR Code**
   - Click "Generate QR Code" button
   - Wait for the QR code to appear

4. **Scan with WhatsApp**
   - Open WhatsApp on your phone
   - Go to **Settings** → **Linked Devices**
   - Tap **Link a Device**
   - Point your camera at the QR code on screen

5. **Wait for Confirmation**
   - The page will automatically detect when pairing is successful
   - You'll see a success message with your Sender ID
   - Click "Go to Dashboard" to return to the main page

### For Administrators (Docker)

If you need to add senders via command line while using Docker:

```bash
# Using QR Code (interactive terminal required)
docker-compose exec whatspoints ./whatspoints -add-sender

# Using Pairing Code
docker-compose exec whatspoints ./whatspoints -add-sender-code=+6281234567890
```

## Technical Details

### API Endpoints

#### Start QR Registration
```http
POST /api/register-sender-qr
Authorization: Basic <base64(username:password)>
Content-Type: application/json
```

**Response:**
```json
{
  "success": true,
  "session_id": "uuid-here",
  "qr_code": "base64-encoded-png-image",
  "message": "QR code generated. Please scan with WhatsApp."
}
```

#### Check Registration Status
```http
GET /api/register-sender-status/{sessionId}
Authorization: Basic <base64(username:password)>
```

**Response (Pending):**
```json
{
  "success": true,
  "status": "pending",
  "qr_code": "updated-base64-encoded-png-image",
  "message": "Waiting for WhatsApp pairing..."
}
```

**Response (Connected):**
```json
{
  "success": true,
  "status": "connected",
  "sender_id": "6281234567890@s.whatsapp.net",
  "message": "Successfully registered! Sender ID: 6281234567890@s.whatsapp.net"
}
```

**Response (Failed):**
```json
{
  "success": true,
  "status": "failed",
  "message": "Registration failed. Please try again."
}
```

### QR Code Refresh Mechanism

1. **Backend**: 
   - Listens to WhatsApp QR channel
   - Generates new PNG image for each QR code received
   - Stores the latest QR code in the session
   - Returns updated QR code when status is checked

2. **Frontend**:
   - Polls status endpoint every 2 seconds
   - Compares current QR code with new one
   - Updates the image if QR code has changed
   - Provides console logging for debugging

### Session Management

- **Session Duration**: 10 minutes
- **Cleanup**: Old sessions are automatically cleaned up
- **Concurrent Sessions**: Multiple registration sessions can run simultaneously
- **Session ID**: Used to track and retrieve registration status

## Troubleshooting

### QR Code Not Displaying

1. **Check Browser Console**: Open browser developer tools (F12) and check for errors
2. **Verify API Connection**: Ensure the backend is running and accessible
3. **Check Authentication**: Verify API credentials are correct
4. **Network Issues**: Check if there are any network connectivity problems

### QR Code Scan Fails

1. **Try Refreshing**: The QR code auto-refreshes, but you can reload the page
2. **Check Phone Internet**: Ensure your phone has an active internet connection
3. **WhatsApp Version**: Make sure WhatsApp is updated to the latest version
4. **Clear WhatsApp Cache**: In WhatsApp settings, try clearing cache

### "Invalid QR Code" Error

This usually happens when:
- QR code has expired (should auto-refresh)
- Network latency is too high
- WebSocket connection was interrupted

**Solution**: The page should automatically refresh the QR code. If not, click "Try Again"

### Session Expired

Sessions expire after 10 minutes. If you see this error:
1. Click "Try Again" button
2. Generate a new QR code
3. Scan immediately to avoid expiration

## Code Changes Summary

### Backend Changes

**File: `internal/application/sender_registration_service.go`**
- Added debug logging for QR code generation
- Enhanced QR code update mechanism
- Continuous QR code refresh handling

### Frontend Changes

**File: `web/register.html`**
- Enabled QR code registration option (was commented out)
- Added automatic QR code refresh logic
- Enhanced status checking with QR code updates
- Added session expiration handling

## Security Considerations

1. **Authentication**: All registration endpoints require Basic Auth
2. **Session IDs**: Use UUIDs to prevent session hijacking
3. **Timeout**: Sessions automatically expire after 10 minutes
4. **HTTPS**: Use HTTPS in production to protect QR code data
5. **Rate Limiting**: Consider adding rate limiting to prevent abuse

## Best Practices

### For Production

1. **Use HTTPS**: Always use HTTPS to encrypt QR code transmission
2. **Reverse Proxy**: Use nginx or Caddy with proper SSL configuration
3. **Monitor Sessions**: Track registration sessions for debugging
4. **Limit Concurrent**: Consider limiting concurrent registration sessions
5. **Logging**: Enable detailed logging for troubleshooting

### For Development

1. **Test Both Methods**: Test both QR code and pairing code registration
2. **Check Logs**: Monitor backend logs during registration
3. **Browser Console**: Keep browser console open to see QR updates
4. **Multiple Devices**: Test with different phones and browsers

## Example Usage Flow

```bash
# 1. Start the application
docker-compose up -d

# 2. Open browser and navigate to registration page
# http://localhost:8080/register

# 3. Enter API password (from .env file)

# 4. Select QR Code method

# 5. Click "Generate QR Code"

# 6. Scan with WhatsApp on phone

# 7. Wait for success message

# 8. New sender is ready to use!
```

## Debugging Tips

### Enable Verbose Logging

Check backend logs:
```bash
docker-compose logs -f whatspoints
```

Look for these log messages:
- `QR Code received, generating image...`
- `QR Code updated (length: X bytes)`
- `QR Code scan successful!`
- `QR Channel closed`

### Frontend Debugging

Open browser console (F12) and look for:
- `QR code updated` - indicates QR refresh is working
- Any error messages from API calls
- Network tab to inspect API requests/responses

### Common Issues

**Issue**: QR code not refreshing
- **Cause**: Status polling not working
- **Solution**: Check browser console for errors, verify API endpoint is accessible

**Issue**: "Failed to connect" error
- **Cause**: WhatsApp servers unreachable or WebSocket issues
- **Solution**: Check internet connection, firewall settings, try again

**Issue**: Registration succeeds but sender not showing
- **Cause**: Database write failed or client manager not updated
- **Solution**: Check backend logs, verify database connection

## Future Enhancements

Potential improvements:
- WebSocket for real-time QR updates (instead of polling)
- Push notifications when registration succeeds
- Better error messages and recovery
- QR code size customization
- Mobile-responsive QR display
- Multi-language support

## Support

For issues or questions:
1. Check the logs first
2. Review this guide
3. Check the API response messages
4. Verify environment configuration
5. Test with simple curl commands to isolate issues
