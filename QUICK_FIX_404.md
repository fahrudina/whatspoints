# Quick Fix: 404 Error on Tencent Cloud

## Problem
Getting `404` error when accessing the root path `/` on your Tencent Cloud server.

## Root Cause
The `web/` directory is not in the same location as the binary when running on the server.

## Solution

### Fix Applied
The application now automatically searches for the `web/` directory in multiple locations:
1. `./web` (relative to current directory)
2. `/app/web` (common deployment path)
3. One level up from current directory

### What to Check on Your Server

1. **SSH into your server:**
   ```bash
   ssh your-user@your-server-ip
   ```

2. **Check where your application is running from:**
   ```bash
   pwd
   # Should show something like: /app or /home/user/whatspoints
   ```

3. **Verify the web directory exists:**
   ```bash
   ls -la web/
   # You should see: index.html and register.html
   ```

4. **If web directory is missing, upload it:**
   ```bash
   # From your local machine
   scp -r web/ your-user@your-server-ip:/app/
   ```

5. **Rebuild and redeploy:**
   ```bash
   # On your local machine
   go build -o whatspoints

   # Upload the new binary
   scp whatspoints your-user@your-server-ip:/app/
   ```

6. **Restart your application:**
   ```bash
   # If using systemd
   sudo systemctl restart whatspoints

   # If running directly, stop the old process and start new one
   pkill whatspoints
   ./whatspoints
   ```

7. **Check the logs to verify web directory is found:**
   ```bash
   # You should see output like:
   # Current working directory: /app
   # Found web directory at: /app/web
   # Using web directory: /app/web
   ```

### Quick Deploy Using Script

Use the provided deployment script:
```bash
# From your local machine
./deploy.sh user@server-ip /app
```

This will automatically:
- Build the application
- Upload binary and web files
- Set correct permissions

### Directory Structure Should Look Like This

```
/app/
├── whatspoints          # The binary
├── web/                 # This must exist!
│   ├── index.html      # Main page
│   └── register.html   # Registration page
└── .env                # Environment variables (optional)
```

### Test After Deployment

```bash
# On the server
curl http://localhost:8080/health
# Should return: {"status":"ok"}

curl http://localhost:8080/
# Should return HTML content, not 404

# Check from external access
curl http://your-server-ip:8080/
```

### Still Getting 404?

If you're still getting 404 after following the above steps:

1. **Check file permissions:**
   ```bash
   chmod +x whatspoints
   chmod -R 644 web/*
   ```

2. **Verify the working directory when starting:**
   ```bash
   # Start from the directory containing web/
   cd /app
   ./whatspoints
   ```

3. **Check application logs for web directory detection:**
   ```bash
   # Look for lines like:
   # "Found web directory at: ..."
   # "Using web directory: ..."
   ```

4. **Manually specify the web directory** (if needed):
   You can modify the code to use an absolute path or set an environment variable.

## Prevention

Always deploy these together:
- ✅ `whatspoints` binary
- ✅ `web/` directory
- ✅ `.env` file (or environment variables)

## Need More Help?

See the full deployment guide in [DEPLOYMENT.md](./DEPLOYMENT.md)
