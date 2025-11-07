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
```bash
# Build the application
go build -o whatspoints

# Run the built binary
./whatspoints
```

### 📱 WhatsApp Setup

#### Single Sender (Default)
1. **First Run**: When you start the application, it will generate a QR code
2. **Scan QR Code**: Use WhatsApp on your phone to scan the QR code
3. **Connection**: Once connected, the service will maintain the session in PostgreSQL

#### Multiple Senders (Advanced)
To register multiple sender phone numbers:

1. **Initial Setup**: Connect the first WhatsApp account as described above
2. **Add Additional Senders**: Use the `-add-sender` command to add more phone numbers:
   ```bash
   # Add a new WhatsApp phone number
   ./whatspoints -add-sender
   
   # Or during development
   go run main.go -add-sender
   ```
   
   This will:
   - Display a QR code in your terminal
   - Wait for you to scan it with the new WhatsApp account
   - Automatically register the sender in the database
   - Make it available for sending messages

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

#### Build Docker Image

```bash
docker build -t whatspoints .
```

#### Run with Docker

```bash
docker run -d --name whatspoints \
  -p 8080:8080 \
  -e SUPABASE_HOST=your-project.supabase.co \
  -e SUPABASE_PORT=6543 \
  -e SUPABASE_USER=postgres \
  -e SUPABASE_PASSWORD=your_password \
  -e SUPABASE_DB=postgres \
  -e SUPABASE_SSLMODE=require \
  -e API_HOST=0.0.0.0 \
  -e API_PORT=8080 \
  -e API_USERNAME=admin \
  -e API_PASSWORD=your_secure_password \
  whatspoints
```

#### Docker Compose

```yaml
version: '3.8'
services:
  whatspoints:
    build: .
    ports:
      - "8080:8080"
    environment:
      - SUPABASE_HOST=your-project.supabase.co
      - SUPABASE_PORT=6543
      - SUPABASE_USER=postgres
      - SUPABASE_PASSWORD=your_password
      - SUPABASE_DB=postgres
      - SUPABASE_SSLMODE=require
      - API_HOST=0.0.0.0
      - API_PORT=8080
      - API_USERNAME=admin
      - API_PASSWORD=your_secure_password
    restart: unless-stopped
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
