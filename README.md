# WhatsPoints: WhatsApp API Service with Clean Architecture

A modern, production-ready WhatsApp messaging service built with Go, featuring clean architecture principles, REST API endpoints, and comprehensive testing capabilities. This service integrates with WhatsApp using the latest [Whatsmeow](https://github.com/tulir/whatsmeow) package and provides both direct WhatsApp integration and HTTP API access.

## 🚀 Features

### Core Functionality
- **WhatsApp Integration**: Latest Whatsmeow package with PostgreSQL session storage
- **Multiple Sender Support**: Register and manage multiple WhatsApp sender accounts
- **REST API**: HTTP endpoints for sending messages and checking status
- **Sender Selection**: Choose specific sender per API call for flexible messaging
- **Clean Architecture**: Domain-driven design with proper separation of concerns
- **Database Integration**: Supabase PostgreSQL with transaction pooler support
- **Authentication**: HTTP Basic Auth for API security

### Architecture & Quality
- **Clean Architecture**: Domain, Application, Infrastructure, and Presentation layers
- **Dependency Injection**: Proper IoC container and dependency management
- **Unit Testing**: Comprehensive test suite with mocks and 100% coverage
- **Error Handling**: Structured error responses and proper HTTP status codes
- **Environment Configuration**: Secure configuration management with .env support

### API Endpoints
- `POST /api/send-message` - Send WhatsApp messages via REST API
- `GET /api/status` - Check WhatsApp connection and service status
- `GET /api/senders` - List all available WhatsApp sender accounts
- `POST /api/ai/reply` - Generate a suggested AI reply (optional; see [AI Reply Suggestion](#-ai-reply-suggestion-optional))
- `GET /health` - Health check endpoint for monitoring

## 📋 Prerequisites

- **Go 1.21+**: Latest Go version for optimal performance
- **Supabase Account**: PostgreSQL database with transaction pooler
- **WhatsApp Account**: For WhatsApp Business API integration

## 🛠 Installation

### 1. Clone the Repository

```bash
git clone https://github.com/fahrudina/whatspoints.git
cd whatspoints
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Environment Setup

Create a `.env` file in the root directory:

```bash
# Database Configuration (Supabase)
SUPABASE_HOST=your-project.supabase.co
SUPABASE_PORT=6543
SUPABASE_USER=postgres
SUPABASE_PASSWORD=your_password
SUPABASE_DB=postgres
SUPABASE_SSLMODE=require

# API Configuration
API_HOST=localhost
API_PORT=8080
API_USERNAME=admin
API_PASSWORD=your_secure_password

# WhatsApp Configuration (Optional)
WHATSAPP_LOG_LEVEL=INFO

# AWS Configuration (for future features)
AWS_REGION=us-east-1
S3_BUCKET_NAME=your_bucket_name
```

## 🚦 Usage

### Running the Application

#### Development Mode
```bash
# Load environment variables and start the service
go run main.go
```

#### Production Build

**Using Makefile (Recommended)**
```bash
# Build for current platform
make build

# Build for Ubuntu Linux (amd64)
make linux

# Build for macOS (Apple Silicon/arm64)
make macos-arm64

# Build for all platforms
make build-all

# View all available commands
make help
```

**Manual Build**
```bash
# Build the application
go build -o whatspoints

# Run the built binary
./whatspoints
```

The Makefile supports:
- **Linux builds**: `make linux` (amd64), `make linux-arm64` (ARM64)
- **macOS builds**: `make macos` (Intel), `make macos-arm64` (Apple Silicon)
- **Testing**: `make test`, `make test-coverage`
- **Utilities**: `make clean`, `make deps`, `make install`

All cross-platform builds are output to the `build/` directory.

#### CLI Commands

```bash
# Start the API server (default)
./whatspoints

# Add new sender using QR code
./whatspoints -add-sender

# Add new sender using SMS pairing code
./whatspoints -add-sender-code=+1234567890

# Clear all WhatsApp sessions
./whatspoints -clear-sessions

# Show help
./whatspoints -h
```

### 📱 WhatsApp Setup

#### Single Sender (Default)
1. **First Run**: When you start the application, it will generate a QR code
2. **Scan QR Code**: Use WhatsApp on your phone to scan the QR code
3. **Connection**: Once connected, the service will maintain the session in PostgreSQL

#### Multiple Senders (Advanced)
To register multiple sender phone numbers, you can use **two different pairing methods**:

##### Method 1: QR Code Pairing (Visual)

```bash
# Add a new WhatsApp phone number using QR code
./whatspoints -add-sender

# Or during development
go run main.go -add-sender
```

**Steps:**
1. Run the command above
2. A QR code will appear in your terminal
3. Open WhatsApp on the phone you want to add
4. Scan the QR code
5. The sender is automatically registered

##### Method 2: Phone Number Pairing (SMS Code)

```bash
# Add a new WhatsApp phone number using SMS pairing code
./whatspoints -add-sender-code=+1234567890

# Or during development
go run main.go -add-sender-code=+1234567890
```

**Steps:**
1. Run the command with your phone number (include country code)
2. A pairing code will be sent via SMS to that number
3. The code will also be displayed in the terminal
4. Open WhatsApp on your phone
5. Go to: Settings > Linked Devices > Link a Device > Link with phone number instead
6. Enter the pairing code
7. The sender is automatically registered

**Which method to use?**
- **QR Code**: Best when you have physical access to scan with your phone
- **Pairing Code**: Best for remote setups or when QR scanning is difficult
   
**After pairing:**
- The sender is automatically registered in the database
- The sender becomes available for sending messages via API
- Session is maintained in PostgreSQL for automatic reconnection

3. **Sender Management**: The application automatically tracks registered senders in the `senders` table
4. **Default Sender**: The first connected account becomes the default sender
5. **List Senders**: After adding a sender, the command shows all available sender IDs

**Database Schema for Senders:**
```sql
CREATE TABLE senders (
    sender_id VARCHAR(50) PRIMARY KEY,
    phone_number VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Note:** Each WhatsApp session is stored independently. You can manage multiple sessions by connecting different WhatsApp accounts through separate QR code scans or by programmatically managing the WhatsApp device store.

### 🌐 API Usage

#### Send Message via REST API

```bash
# Send a WhatsApp message (using default sender)
curl -X POST http://localhost:8080/api/send-message \
  -u admin:your_secure_password \
  -H "Content-Type: application/json" \
  -d '{
    "to": "+1234567890",
    "message": "Hello from WhatsPoints API!"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Message sent successfully",
  "id": "message_id_here"
}
```

#### Send Message from Specific Sender

When multiple sender phone numbers are registered, you can specify which sender to use:

```bash
# Send a WhatsApp message from a specific sender
curl -X POST http://localhost:8080/api/send-message \
  -u admin:your_secure_password \
  -H "Content-Type: application/json" \
  -d '{
    "to": "+1234567890",
    "message": "Hello from specific sender!",
    "from": "sender_id_123"
  }'
```

**Note:** The `from` parameter is optional. If not provided, the default sender will be used.

**Response:**
```json
{
  "success": true,
  "message": "Message sent successfully",
  "id": "message_id_here"
}
```

#### List All Available Senders

Get a list of all registered WhatsApp sender phone numbers:

```bash
# Get all available senders
curl -X GET http://localhost:8080/api/senders \
  -u admin:your_secure_password
```

**Response:**
```json
{
  "count": 2,
  "senders": [
    {
      "id": "1234567890",
      "phone_number": "1234567890",
      "name": "Sender 1234567890",
      "is_default": true,
      "is_active": true
    },
    {
      "id": "9876543210",
      "phone_number": "9876543210",
      "name": "Sender 9876543210",
      "is_default": false,
      "is_active": true
    }
  ]
}
```

**Use Case:** Call this endpoint to get the list of sender IDs before sending a message with a specific sender.

**Response:**
```json
{
  "success": true,
  "message": "Message sent successfully",
  "id": "message_id_here"
}
```

#### Check Service Status

```bash
# Check WhatsApp connection status
curl -X GET http://localhost:8080/api/status \
  -u admin:your_secure_password
```

**Response:**
```json
{
  "whatsapp": {
    "connected": true,
    "logged_in": true,
    "jid": "your_number@s.whatsapp.net"
  }
}
```

#### Health Check

```bash
# Health check (no authentication required)
curl -X GET http://localhost:8080/health
```

**Response:**
```json
{
  "status": "ok",
  "service": "whatspoints-api"
}
```

### 🐳 Docker Deployment

#### Using Docker Compose (Recommended)

The easiest way to run WhatsPoints is using Docker Compose:

**1. Ensure your `.env` file is configured:**

Make sure your `.env` file contains all required environment variables (see Environment Setup section above).

**2. Build and start the service:**

```bash
# Build the Docker image
docker-compose build

# Start the service in detached mode
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the service
docker-compose down
```

**3. Access the service:**

The API will be available at `http://localhost:8080` (or your configured `API_PORT`).

**4. QR Code Setup (First Time):**

To set up WhatsApp on first run:

```bash
# View the QR code in logs
docker-compose logs -f whatspoints

# Or connect to the container to see the QR code
docker exec -it whatspoints /bin/sh
```

Scan the QR code with WhatsApp on your phone to link the device.

**Notes:**
- WhatsApp session data persists in the `whatsapp_data` Docker volume
- The service automatically restarts on failure
- Health checks ensure the service is running correctly
- If AWS credentials are not set, you'll see warnings (they're optional)

**Running with the AI sidecar (optional):**

`docker-compose.yml` also defines the `ai-agent` service (the AI reply-suggestion
sidecar, exposed on port `8090`). `docker-compose up -d` starts both services; the
sidecar reads these from your `.env`:

```bash
DATABASE_URL=postgresql://user:password@host:5432/dbname
OPENROUTER_API_KEY=your_openrouter_api_key   # chat LLM
GOOGLE_API_KEY=your_google_ai_studio_api_key # embeddings (Gemini)

# Enable the Go -> sidecar integration:
ENABLE_AI_RESPONSE=true
AI_SERVICE_URL=http://ai-agent:8090          # service name on the compose network
```

Apply the pgvector schema and index your knowledge once before using it (see the
[AI Reply Suggestion](#-ai-reply-suggestion-optional) section). To run **only** the
API without the sidecar, start a single service:

```bash
docker-compose up -d whatspoints
```

#### Manual Docker Commands

**Build Docker Image:**

```bash
docker build -t whatspoints .
```

**Run with Docker:**

```bash
docker run -d --name whatspoints \
  -p 8080:8080 \
  -v whatsapp_data:/app/data \
  -e SUPABASE_HOST=your-project.supabase.co \
  -e SUPABASE_PORT=6543 \
  -e SUPABASE_USER=postgres \
  -e SUPABASE_PASSWORD=your_password \
  -e SUPABASE_DB=postgres \
  -e SUPABASE_SSLMODE=require \
  -e API_PORT=8080 \
  -e API_USERNAME=admin \
  -e API_PASSWORD=your_secure_password \
  whatspoints
```

## ⚙️ Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| **Database Configuration** |
| `SUPABASE_HOST` | ✅ | - | Supabase project host |
| `SUPABASE_PORT` | ✅ | `6543` | Transaction pooler port |
| `SUPABASE_USER` | ✅ | - | Database username |
| `SUPABASE_PASSWORD` | ✅ | - | Database password |
| `SUPABASE_DB` | ✅ | - | Database name |
| `SUPABASE_SSLMODE` | ❌ | `require` | SSL mode |
| **API Configuration** |
| `API_HOST` | ❌ | `localhost` | API server host |
| `API_PORT` | ❌ | `8080` | API server port |
| `API_USERNAME` | ✅ | - | Basic auth username |
| `API_PASSWORD` | ✅ | - | Basic auth password |
| **WhatsApp Configuration** |
| `WHATSAPP_LOG_LEVEL` | ❌ | `INFO` | WhatsApp client log level (DEBUG, INFO, WARN, ERROR) |
| **AWS Configuration (Future)** |
| `AWS_REGION` | ❌ | - | AWS region for S3 |
| `S3_BUCKET_NAME` | ❌ | - | S3 bucket for media storage |

### Database Setup

The application uses Supabase PostgreSQL with transaction pooler for optimal performance:

1. **Create Supabase Project**: Sign up at [supabase.com](https://supabase.com)
2. **Get Connection Details**: Navigate to Settings > Database
3. **Use Transaction Pooler**: Use port `6543` for connection pooling
4. **Session Storage**: WhatsApp sessions are automatically stored in PostgreSQL

## 🏗️ Architecture

This project follows **Clean Architecture** principles with clear separation of concerns:

```
├── internal/
│   ├── domain/          # Business entities and interfaces
│   │   ├── entities.go  # Core business models
│   │   └── interfaces.go # Repository and service contracts
│   ├── application/     # Business logic layer
│   │   ├── message_service.go # Message business logic
│   │   └── auth_service.go    # Authentication logic
│   ├── infrastructure/ # External services implementation
│   │   └── whatsapp_repository.go # WhatsApp client wrapper
│   └── presentation/   # HTTP handlers and routing
│       ├── handlers.go     # HTTP request handlers
│       ├── middleware.go   # Authentication middleware
│       └── router.go       # Route definitions
├── api/                # API server setup
│   └── server.go       # HTTP server configuration
├── whatsapp/          # WhatsApp client initialization
│   └── whatsapp.go    # WhatsApp client setup
└── main.go           # Application entry point
```

### Layer Responsibilities

- **Domain**: Core business rules and entities
- **Application**: Use cases and business logic orchestration
- **Infrastructure**: External service integrations (WhatsApp, Database)
- **Presentation**: HTTP API and request/response handling

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/application/...
go test ./internal/presentation/...
```

### Test Structure

- **Unit Tests**: Comprehensive test coverage for all layers
- **Mock Testing**: Uses `testify/mock` for dependency isolation
- **HTTP Testing**: Tests API endpoints with `httptest`
- **Integration Points**: Tests layer interactions

```bash
# Example test output
✅ internal/application/message_service_test.go
✅ internal/application/auth_service_test.go  
✅ internal/presentation/handlers_test.go
✅ internal/presentation/middleware_test.go
```

## 🔮 Roadmap

### Current Status ✅
- [x] WhatsApp integration with latest Whatsmeow
- [x] REST API for message sending
- [x] Multiple sender phone number support
- [x] Sender selection per API call
- [x] Clean architecture implementation
- [x] PostgreSQL session storage
- [x] Basic authentication
- [x] Comprehensive unit testing
- [x] Docker support

### Upcoming Features 🚧
- [ ] **Media Handling**: Support for image, document, and audio messages
- [ ] **Receipt Processing**: OCR integration for receipt data extraction
- [ ] **Point System**: Implement point calculation and management
- [ ] **LLM Integration**: AI-powered receipt analysis
- [ ] **User Management**: Multi-user support and user profiles
- [ ] **Notifications**: Real-time notifications and status updates
- [ ] **Analytics**: Usage statistics and reporting
- [ ] **Webhooks**: External system integration via webhooks

### Future Enhancements 🎯
- [ ] **GraphQL API**: Advanced query capabilities
- [ ] **Rate Limiting**: API rate limiting and throttling
- [ ] **Caching**: Redis integration for improved performance
- [ ] **Monitoring**: Prometheus metrics and health monitoring
- [ ] **Message Queue**: Async processing with message queues
- [ ] **Sender Management UI**: Web interface for managing multiple sender accounts

## 🤝 Contributing

We welcome contributions! Please follow these steps:

1. **Fork the Repository**
2. **Create Feature Branch**: `git checkout -b feature/amazing-feature`
3. **Make Changes**: Implement your feature with tests
4. **Run Tests**: `go test ./...`
5. **Commit Changes**: `git commit -m 'Add amazing feature'`
6. **Push Branch**: `git push origin feature/amazing-feature`
7. **Open Pull Request**

### Development Guidelines

- Follow Go best practices and conventions
- Maintain clean architecture principles
- Write comprehensive unit tests
- Update documentation for new features
- Use meaningful commit messages

## 🤖 AI Reply Suggestion (Optional)

An optional RAG assistant (Python sidecar under [`ai-agent/`](ai-agent/README.md))
generates **suggested** WhatsApp replies in Bahasa Indonesia from a pgvector
knowledge base. It is **disabled by default** and **never auto-sends** anything —
this phase only produces suggestions. Manual `/api/send-message` is completely
unaffected by the AI toggle.

Architecture: Go API → HTTP → Python sidecar (FastAPI + LangGraph) → Gemini
embeddings + Postgres/pgvector. Chat LLM runs through OpenRouter; embeddings run
through Google AI Studio (`gemini-embedding-001`).

### 1. Apply the pgvector schema

```bash
psql "$DATABASE_URL" -f database/vector_schema.sql
# On Supabase you can paste database/vector_schema.sql into the SQL editor.
```

### 2. Insert knowledge

```sql
INSERT INTO knowledge_base (title, content, category)
VALUES ('Promo Laundry', 'Promo Cuci 8KG Rp10.000 berlaku Senin sampai Rabu.', 'promo');
```

### 3. Generate embeddings (indexer)

```bash
cd ai-agent
pip install -r requirements.txt
cp .env.example .env   # set DATABASE_URL, OPENROUTER_API_KEY, GOOGLE_API_KEY
python index_knowledge.py
```

Only rows with `embedding IS NULL` are processed, so it is safe to rerun.
**No service restart is needed** after inserting + indexing new data — retrieval
queries pgvector on every request.

### 4. Start the AI sidecar

```bash
cd ai-agent
uvicorn app:api --host 0.0.0.0 --port 8090
# health: curl http://localhost:8090/health
```

### 5. Enable AI on the Go service

```env
ENABLE_AI_RESPONSE=true
ENABLE_AI_AUTO_SEND=false          # reserved; auto-send is NOT implemented yet
AI_SERVICE_URL=http://localhost:8090
```

`ENABLE_AI_RESPONSE` accepts `true`, `1`, `yes`, or `on`. Anything else (or a
missing value) keeps the feature **disabled**. `AI_SERVICE_URL` is only read when
enabled and defaults to `http://localhost:8090`.

### 6. Disable AI

```env
ENABLE_AI_RESPONSE=false
```

When disabled, `POST /api/ai/reply` stays registered but returns **HTTP 503**:

```json
{ "success": false, "message": "AI response feature is disabled" }
```

`/api/send-message`, `/api/status`, `/api/senders`, and `/health` keep working
normally — the AI toggle only controls AI-generated reply suggestions and does
not affect manual WhatsApp sending.

### 7. Call the endpoint

```bash
curl -X POST http://localhost:8080/api/ai/reply \
  -u admin:your_password \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Kak promo cuci kiloan masih ada?",
    "phone_number": "628123456789"
  }'
```

Enabled response:

```json
{
  "reply": "Masih kak 😊 Promo Cuci 8KG Rp10.000 berlaku Senin sampai Rabu ya.",
  "intent": "ask_promo",
  "sources": [
    { "id": 1, "title": "Promo Laundry", "content": "Promo Cuci 8KG Rp10.000 berlaku Senin sampai Rabu.", "category": "promo", "score": 0.12 }
  ]
}
```

An empty `message` returns **HTTP 400** `{ "success": false, "message": "message is required" }`.

> **Note:** `ENABLE_AI_AUTO_SEND` is a reserved toggle for a future phase.
> Auto-send is not implemented — the endpoint only returns suggestions.

See [`ai-agent/README.md`](ai-agent/README.md) for full sidecar configuration and Docker usage.

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/fahrudina/whatspoints/issues)
- **Discussions**: [GitHub Discussions](https://github.com/fahrudina/whatspoints/discussions)
- **Documentation**: Check the `/docs` folder for detailed guides

## 🙏 Acknowledgments

- [Whatsmeow](https://github.com/tulir/whatsmeow) - Excellent WhatsApp Web API library
- [Gin](https://github.com/gin-gonic/gin) - Fast HTTP web framework
- [Testify](https://github.com/stretchr/testify) - Testing toolkit
- [Supabase](https://supabase.com) - Open source Firebase alternative
