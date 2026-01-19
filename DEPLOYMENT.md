# Deployment Guide for Tencent Cloud

## Prerequisites
- Go 1.21 or higher installed on the server
- PostgreSQL database (Supabase or other)
- Environment variables configured

## Deployment Steps

### 1. Build the Application

```bash
# On your local machine or build server
go build -o whatspoints
```

### 2. Prepare Deployment Files

Make sure to include these files/directories when deploying:
- `whatspoints` (the compiled binary)
- `web/` directory (contains index.html and register.html)
- `.env` file (with your environment variables)

### 3. Directory Structure on Server

Your deployment should look like this:
```
/app/
├── whatspoints          # The binary
├── web/                 # Web UI files
│   ├── index.html
│   └── register.html
└── .env                 # Environment variables
```

### 4. Transfer Files to Tencent Cloud

```bash
# Example using scp
scp whatspoints user@your-server:/app/
scp -r web/ user@your-server:/app/
scp .env user@your-server:/app/
```

Or using rsync:
```bash
rsync -avz --exclude 'node_modules' --exclude '.git' \
  whatspoints web/ .env user@your-server:/app/
```

### 5. Set Correct Permissions

```bash
ssh user@your-server
cd /app
chmod +x whatspoints
chmod 644 web/*.html
```

### 6. Environment Variables

Ensure your `.env` file contains:
```env
# Database
SUPABASE_URL=your-supabase-url
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key
DATABASE_URL=postgresql://user:password@host:5432/database

# API Configuration
API_PORT=8080
API_USERNAME=admin
API_PASSWORD=your-secure-password

# WhatsApp Configuration
WHATSAPP_LOG_LEVEL=INFO
```

### 7. Run the Application

#### Option A: Direct Run
```bash
cd /app
./whatspoints
```

#### Option B: Using systemd (Recommended for production)

Create a systemd service file:
```bash
sudo nano /etc/systemd/system/whatspoints.service
```

Add this content:
```ini
[Unit]
Description=WhatsPoints WhatsApp Service
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/app
ExecStart=/app/whatspoints
Restart=always
RestartSec=10
Environment="PATH=/usr/local/bin:/usr/bin:/bin"

[Install]
WantedBy=multi-user.target
```

Enable and start the service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable whatspoints
sudo systemctl start whatspoints
sudo systemctl status whatspoints
```

View logs:
```bash
sudo journalctl -u whatspoints -f
```

### 8. Set Up Reverse Proxy (Optional but Recommended)

Using Nginx:
```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 9. Verify Deployment

Check that the application is running:
```bash
# Check process
ps aux | grep whatspoints

# Check port
netstat -tulpn | grep 8080

# Test endpoints
curl http://localhost:8080/health
curl http://localhost:8080/
```

## Troubleshooting

### Issue: 404 on root path "/"

**Cause**: Web directory not found

**Solution**: The application now automatically searches for the web directory. Check the logs to see where it's looking:
```bash
# You should see output like:
Current working directory: /app
Found web directory at: /app/web
Using web directory: /app/web
```

If the web directory is not found, ensure:
1. The `web/` directory exists in the same location as the binary
2. Files have correct permissions: `ls -la /app/web/`
3. The working directory is correct when starting the app

### Issue: Cannot find .env file

**Solution**: Either:
1. Place `.env` in the same directory as the binary
2. Or export environment variables directly:
```bash
export DATABASE_URL="postgresql://..."
export API_PASSWORD="your-password"
./whatspoints
```

### Issue: Database connection failed

**Solution**:
1. Check database URL is correct
2. Ensure firewall allows connections to database
3. Verify database credentials
4. Test connection: `psql "$DATABASE_URL"`

## Adding a Sender (Important!)

After deployment, you need to add at least one WhatsApp sender:

```bash
# Using QR code
./whatspoints -add-sender

# Using pairing code
./whatspoints -add-sender-code +1234567890
```

Note: With the latest update, the app will now wait for the connection to complete before exiting, ensuring successful registration.

## Security Checklist

- [ ] Change default API_PASSWORD
- [ ] Use HTTPS in production (configure Nginx with SSL)
- [ ] Set up firewall rules
- [ ] Secure database credentials
- [ ] Use strong authentication for server access
- [ ] Regular backups of database

## Monitoring

Monitor the application:
```bash
# Using systemd
sudo journalctl -u whatspoints -f

# Check application logs
tail -f /var/log/whatspoints.log  # if logging to file

# Monitor resource usage
htop
```
