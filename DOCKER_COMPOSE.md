# Docker Compose Quick Start Guide

This guide explains how to run WhatsPoints using Docker Compose.

## Prerequisites

- Docker and Docker Compose installed on your system
- Access to a PostgreSQL/Supabase database
- AWS S3 credentials (optional, for image storage)

## Setup Instructions

### 1. Configure Environment Variables

Create a `.env` file in the project root by copying the example:

```bash
cp .env.example .env
```

Edit `.env` and configure your settings:

```env
# Database Configuration (Required)
SUPABASE_HOST=your-project.pooler.supabase.com
SUPABASE_PORT=6543
SUPABASE_USER=postgres.your-project-id
SUPABASE_PASSWORD=your-password
SUPABASE_DB=postgres
SUPABASE_SSLMODE=require

# API Configuration (Required)
API_PORT=8080
API_USERNAME=admin
API_PASSWORD=your-secure-password

# AWS S3 Configuration (Optional)
AWS_REGION=ap-southeast-2
S3_BUCKET_NAME=your-bucket-name
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key

# WhatsApp Configuration
ALLOWED_PHONE_NUMBERS=6281234567890
```

### 2. Start the Service

Build and start the service in detached mode:

```bash
docker-compose up -d
```

Or to see logs in real-time:

```bash
docker-compose up
```

### 3. Add WhatsApp Sender

After the service is running, you need to add at least one WhatsApp sender.

**Option A: Using QR Code (Recommended)**

```bash
docker-compose exec whatspoints ./whatspoints -add-sender
```

Scan the QR code with WhatsApp on your phone (Settings → Linked Devices → Link a Device).

**Option B: Using Pairing Code**

```bash
docker-compose exec whatspoints ./whatspoints -add-sender-code=+6281234567890
```

Enter the pairing code you receive via SMS in WhatsApp.

### 4. Verify Service Status

Check if the service is running:

```bash
# View logs
docker-compose logs -f whatspoints

# Check container status
docker-compose ps

# Test API health endpoint
curl http://localhost:8080/health
```

## Common Operations

### View Available Senders

List all registered WhatsApp senders:

```bash
docker-compose exec whatspoints ./whatspoints -add-sender
# Cancel after seeing the list
```

### Clear All Sessions

Remove all WhatsApp sessions:

```bash
docker-compose exec whatspoints ./whatspoints -clear-sessions
```

### Restart Service

```bash
docker-compose restart whatspoints
```

### Stop Service

```bash
docker-compose stop
```

### Stop and Remove Containers

```bash
docker-compose down
```

### Stop and Remove Containers + Volumes (⚠️ This will delete WhatsApp sessions)

```bash
docker-compose down -v
```

### View Logs

```bash
# Follow logs
docker-compose logs -f

# View last 100 lines
docker-compose logs --tail=100

# View logs for specific service
docker-compose logs -f whatspoints
```

### Rebuild After Code Changes

```bash
docker-compose up -d --build
```

## API Usage

Once the service is running, you can use the API endpoints:

### Send Message

```bash
curl -X POST http://localhost:8080/api/send-message \
  -u admin:your-password \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "6281234567890",
    "message": "Hello from WhatsPoints!",
    "sender_id": "6281234567890@s.whatsapp.net"
  }'
```

### Get Status

```bash
curl http://localhost:8080/api/status \
  -u admin:your-password
```

### List Senders

```bash
curl http://localhost:8080/api/senders \
  -u admin:your-password
```

### Health Check

```bash
curl http://localhost:8080/health
```

## Troubleshooting

### Service Won't Start

1. Check if port 8080 is already in use
2. Verify all required environment variables are set in `.env`
3. Check logs: `docker-compose logs whatspoints`

### Database Connection Issues

1. Verify database credentials in `.env`
2. Ensure database is accessible from Docker container
3. Check SSL mode settings

### WhatsApp Not Connecting

1. Clear sessions: `docker-compose exec whatspoints ./whatspoints -clear-sessions`
2. Re-add sender with QR code or pairing code
3. Ensure phone number has WhatsApp installed and active

### Container Keeps Restarting

Check logs for errors:

```bash
docker-compose logs --tail=50 whatspoints
```

## Port Configuration

By default, the service runs on port 8080. To change it:

1. Update `API_PORT` in `.env`
2. Update port mapping in `docker-compose.yml`:
   ```yaml
   ports:
     - "3000:8080"  # External:Internal
   ```

## Data Persistence

WhatsApp session data is stored in a Docker volume named `whatsapp_data`. This persists between container restarts but will be lost if you run `docker-compose down -v`.

To backup session data:

```bash
docker run --rm -v whatspoints_whatsapp_data:/data -v $(pwd):/backup alpine tar czf /backup/whatsapp-backup.tar.gz -C /data .
```

To restore session data:

```bash
docker run --rm -v whatspoints_whatsapp_data:/data -v $(pwd):/backup alpine tar xzf /backup/whatsapp-backup.tar.gz -C /data
```

## Security Notes

- Never commit `.env` file to version control
- Use strong passwords for `API_PASSWORD`
- Restrict `ALLOWED_PHONE_NUMBERS` to trusted numbers only
- Consider using secrets management for production deployments
- Use HTTPS in production with a reverse proxy (nginx, Caddy, Traefik)

## Production Deployment

For production, consider:

1. Using Docker secrets instead of environment variables
2. Setting up a reverse proxy with SSL/TLS
3. Implementing rate limiting
4. Using a managed database service
5. Setting up monitoring and alerts
6. Regular backups of WhatsApp session data
