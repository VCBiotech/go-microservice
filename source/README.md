# File Manager Microservice

A high-performance Go microservice for file management with multi-cloud storage support, template rendering, and PDF generation capabilities.

## 🚀 Features

- **Multi-Cloud Storage**: Support for AWS S3 and Google Cloud Storage with configurable default provider
- **Template Rendering**: Dynamic HTML template rendering with JSON data injection using Go templates
- **PDF Generation**: Convert rendered templates to PDF using Gotenberg service
- **Server Authentication**: PIN-based authentication system with bcrypt hashing for server-to-server communication
- **Role-Based Access Control**: Fine-grained permissions with admin, calculator, and analytics roles
- **Metadata Management**: In-memory metadata store for file information (PostgreSQL integration planned)
- **Telemetry & Logging**: Structured logging with colored output and OpenTelemetry support
- **Rate Limiting**: Built-in rate limiting for API protection (1000 requests per time window)
- **Health Checks**: Ready-to-use health check endpoints
- **Docker Support**: Containerized deployment with multi-stage Docker builds
- **Hot Reload Development**: Air configuration for automatic reloading during development
- **ECR Integration**: AWS ECR deployment pipeline with automated builds

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client Apps   │    │  Load Balancer  │    │  File Manager   │
│                 │───▶│                 │───▶│   Service       │
│ (Servers/APIs)  │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                       ┌────────────────────────────────┼────────────────────────────────┐
                       │                                │                                │
                       ▼                                ▼                                ▼
              ┌─────────────────┐              ┌─────────────────┐              ┌─────────────────┐
              │   AWS S3        │              │ Google Cloud    │              │   Gotenberg     │
              │   Storage       │              │   Storage       │              │  PDF Service    │
              └─────────────────┘              └─────────────────┘              └─────────────────┘
```

## 📋 Prerequisites

- **Go**: 1.24.2 or later
- **Docker**: For containerized deployment
- **Gotenberg**: PDF conversion service (runs on port 3001)
- **AWS Account**: For S3 storage (optional)
- **Google Cloud Account**: For GCS storage (optional)
- **Air**: For hot reload development (optional, install with `go install github.com/air-verse/air@latest`)

## 🔧 Installation & Setup

### 1. Clone the Repository

```bash
git clone <repository-url>
cd go-microservice/source
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configuration

Create your environment configuration by copying the example file and updating it:

```bash
cp config.env.example .env
```

The service supports configuration via environment variables or a JSON secrets string.

#### Environment Variables

Set the following environment variables in your `.env` file or export them:

```bash
# Server Configuration
SERVER_PORT=3000

# AWS S3 Configuration
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-aws-access-key-here
AWS_SECRET_ACCESS_KEY=your-aws-secret-key-here

# Google Cloud Storage Configuration
GCP_PROJECT_ID=your-gcp-project-id
GCP_CREDENTIALS_FILE=/path/to/your/gcp-credentials.json

# Storage Configuration
DEFAULT_CLOUD=aws
BUCKET_NAME=your-s3-bucket-name
REPLICATE_TO_ALL_CLOUDS=false

# Database Configuration (Optional - for future PostgreSQL integration)
DATABASE_URL=postgres://username:password@localhost:5432/database_name

# Gotenberg PDF Service (External dependency)
GOTENBERG_URL=http://localhost:3001

# Development Settings
LOG_LEVEL=info
DEBUG=false
```

#### JSON Secrets (Alternative)

Use this method for containerized deployments:

```bash
export SECRETS='{"AWS_ACCESS_KEY_ID":"your-key","AWS_SECRET_ACCESS_KEY":"your-secret","BUCKET_NAME":"your-bucket","GCP_PROJECT_ID":"your-project","DEFAULT_CLOUD":"aws"}'
```

### 4. Start Gotenberg Service (Required for PDF Generation)

```bash
docker run --rm -p 3001:3000 gotenberg/gotenberg:7
```

### 5. Run the Service

```bash
# Development mode with hot reload using Air
make dev

# Or build and run manually
go build -o bin/file-manager .
./bin/file-manager
```

## 🐳 Docker Deployment

### Build Docker Image

```bash
make build
```

### Run with Docker

```bash
make run
```

### AWS ECR Deployment

```bash
# Login to ECR, build, tag and push
make build-push
```

## 📚 API Documentation

### Authentication

All API endpoints (except `/health`) require server authentication:

**Required Headers:**

- `X-Server-ID`: Server identifier (`calculator-server`, `analytics-server`, `admin-server`)
- `X-PIN`: Server PIN (default: `123`, `456`, `789` respectively)

### Endpoints

#### Health Check

```http
GET /health
```

**Response:**

```
Status: 200 OK
Body: Ok
```

#### Service Info

```http
GET /
```

**Response:**

```
Status: 200 OK
Body: File Manager Service.
```

#### Template Rendering & PDF Generation

```http
POST /v1/files/render-template
Content-Type: multipart/form-data
X-Server-ID: calculator-server
X-PIN: 123
```

**Form Data:**

- `template`: HTML template file
- `jsonData`: JSON string with template variables

**Example Template (invoice.html):**

```html
<!DOCTYPE html>
<html>
  <head>
    <title>Invoice</title>
    <style>
      body {
        font-family: Arial, sans-serif;
      }
      .header {
        background-color: #f0f0f0;
        padding: 20px;
      }
      .amount {
        font-size: 24px;
        font-weight: bold;
        color: #2c5aa0;
      }
    </style>
  </head>
  <body>
    <div class="header">
      <h1>Invoice #{{.InvoiceNumber}}</h1>
      <p>Date: {{.Date}}</p>
    </div>
    <div class="content">
      <h2>Bill To:</h2>
      <p>{{.CustomerName}}</p>
      <p>{{.CustomerAddress}}</p>
      <h2>Items:</h2>
      <table>
        {{range .Items}}
        <tr>
          <td>{{.Description}}</td>
          <td>${{.Amount}}</td>
        </tr>
        {{end}}
      </table>
      <div class="amount">Total: ${{.Total}}</div>
    </div>
  </body>
</html>
```

**Example JSON Data:**

```json
{
  "InvoiceNumber": "INV-2024-001",
  "Date": "2024-01-15",
  "CustomerName": "John Doe",
  "CustomerAddress": "123 Main St, City, State",
  "Items": [
    { "Description": "Web Development", "Amount": "1500.00" },
    { "Description": "Design Services", "Amount": "500.00" }
  ],
  "Total": "2000.00"
}
```

**Response:**

```json
{
  "message": "File has been successfully created and uploaded to S3."
}
```

## 💡 Usage Examples

### 1. cURL Example

```bash
curl -X POST http://localhost:3000/v1/files/render-template \
  -H "X-Server-ID: calculator-server" \
  -H "X-PIN: 123" \
  -F "template=@examples/invoice-template.html" \
  -F 'jsonData={"InvoiceNumber":"INV-001","CustomerName":"John Doe","Total":"2000.00"}'
```

### 2. Test Script

Use the provided test script to validate your setup:

```bash
# Run from project root
./examples/test-service.sh
```

This script will:

- Check if the service is running
- Start Gotenberg if needed
- Test the health endpoint
- Test template rendering with example files
- Test authentication validation
- Clean up resources

### 3. Go Client Example

```go
package main

import (
    "bytes"
    "io"
    "mime/multipart"
    "net/http"
    "os"
)

func uploadTemplate() error {
    url := "http://localhost:3000/v1/files/render-template"

    var b bytes.Buffer
    writer := multipart.NewWriter(&b)

    file, err := os.Open("template.html")
    if err != nil {
        return err
    }
    defer file.Close()

    fw, err := writer.CreateFormFile("template", "template.html")
    if err != nil {
        return err
    }
    io.Copy(fw, file)

    writer.WriteField("jsonData", `{"name":"John","amount":"100"}`)
    writer.Close()

    req, err := http.NewRequest("POST", url, &b)
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", writer.FormDataContentType())
    req.Header.Set("X-Server-ID", "calculator-server")
    req.Header.Set("X-PIN", "123")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}
```

### 4. Python Client Example

```python
import requests

def upload_template():
    url = "http://localhost:3000/v1/files/render-template"

    headers = {
        "X-Server-ID": "calculator-server",
        "X-PIN": "123"
    }

    files = {
        "template": open("template.html", "rb")
    }

    data = {
        "jsonData": '{"name": "John", "amount": "100"}'
    }

    response = requests.post(url, headers=headers, files=files, data=data)
    return response.json()

result = upload_template()
print(result)
```

## 🛠️ Development

### Available Make Commands

```bash
# Show all available commands
make help

# Start development server with hot reload (uses Air)
make dev

# Build Docker image
make build

# Run tests
make test

# ECR operations
make ecr-login    # Login to AWS ECR
make tag          # Tag Docker image for ECR
make push         # Push image to ECR
make build-push   # Complete ECR deployment pipeline

# Run application locally in Docker
make run

# Database operations (future PostgreSQL integration)
make init-db      # Initialize database
make migrate      # Apply migrations
make new-migration args=<name>  # Create new migration

# Code quality (placeholders for future implementation)
make lint         # Lint code
make format       # Format code
make type-check   # Type checking

# Cleanup
make clean        # Clean build artifacts
```

### Hot Reload Development

The project uses [Air](https://github.com/air-verse/air) for hot reload during development:

1. Install Air: `go install github.com/air-verse/air@latest`
2. Run `make dev` to start with hot reload
3. Air will automatically rebuild and restart the service when files change

Configuration is in `.air.toml`:

- Watches `.go`, `.tpl`, `.tmpl`, `.html` files
- Excludes test files and temporary directories
- Builds to `./tmp/main`

### Project Structure

```
source/
├── .air.toml             # Air hot reload configuration
├── .dockerignore         # Docker ignore patterns
├── .gitignore           # Git ignore patterns
├── config.env.example   # Environment configuration template
├── Dockerfile           # Multi-stage Docker build
├── go.mod              # Go module dependencies
├── go.sum              # Go module checksums
├── main.go             # Application entry point
├── Makefile            # Build and deployment commands
├── README.md           # This file
├── application/         # Application layer
│   ├── app.go          # Main app initialization & server setup
│   ├── config.go       # Config wrapper for backward compatibility
│   └── routes.go       # Route definitions and handlers
├── auth/               # Authentication & authorization
│   └── server_auth.go  # Server auth middleware with bcrypt
├── config/             # Configuration management
│   └── config.go       # Environment and secrets configuration
├── domain/             # Domain/business logic layer
│   └── file/          # File domain
│       ├── hanlder.go  # File handlers (template rendering)
│       ├── model.go    # File models
│       └── repo.go     # File repository logic
├── examples/           # Example files and test scripts
│   ├── invoice-data.json      # Sample JSON data
│   ├── invoice-template.html  # Sample HTML template
│   └── test-service.sh       # Comprehensive test script
├── metadata/           # Metadata management
│   └── metadata.go     # In-memory metadata store interface
├── storage/            # Storage layer
│   ├── aws_s3.go      # AWS S3 adapter implementation
│   ├── gcs.go         # Google Cloud Storage adapter
│   ├── manager.go     # Multi-cloud storage manager
│   └── storage.go     # Storage interface definition
├── telemetry/          # Observability
│   └── logger.go      # Colored structured logging
├── tmp/               # Air build directory (gitignored)
└── utils/             # Utility functions
    ├── aws_helper.go   # AWS S3 utilities
    ├── clerk_helper.go # Clerk auth utilities (commented)
    └── parse_template.go # Template parsing and PDF generation
```

## 🔐 Security

### Server Authentication

The service uses a PIN-based authentication system with bcrypt hashing:

- **calculator-server**: PIN `123` (calculator role)
- **analytics-server**: PIN `456` (analytics role)
- **admin-server**: PIN `789` (admin role)

> ⚠️ **Production Note**: Change default PINs in production and store them securely

### Role-Based Access Control (RBAC)

- **Admin role**: Full access to all operations
- **Calculator role**: Access to template rendering operations
- **Analytics role**: Access to analytics operations

### Rate Limiting

Built-in rate limiting allows 1000 requests per time window per client using Echo's memory store.

### Authentication Headers

All endpoints (except `/health`) require:

- `X-Server-ID`: Must match one of the configured server IDs
- `X-PIN`: Must match the corresponding PIN for the server ID

## 🌐 Multi-Cloud Storage

### Supported Providers

- **AWS S3**: Primary cloud storage with full SDK integration
- **Google Cloud Storage**: Alternative/backup storage with authentication

### Configuration

- Set `DEFAULT_CLOUD` to `aws` or `gcp` to choose your primary storage provider
- Set `REPLICATE_TO_ALL_CLOUDS=true` to replicate files to all configured clouds
- Each cloud provider requires its own authentication configuration

### Storage Manager

The `StorageManager` handles multiple cloud adapters:

- Automatically initializes available adapters based on configuration
- Provides unified interface for all storage operations
- Supports fallback and replication strategies

## 📊 Monitoring & Logging

### Structured Logging

- **Colored Console Output**: Easy-to-read colored logs for development
- **JSON Logging**: Structured logs for production environments
- **Request Tracing**: Unique request IDs with start/end logging
- **Contextual Logging**: Include request metadata in all logs

### Health Monitoring

- **Health Check Endpoint**: `/health` returns service status
- **Request/Response Logging**: Automatic logging of all HTTP requests
- **Error Tracking**: Detailed error logging with context

### Telemetry Integration

- Ready for OpenTelemetry integration
- Request timing and performance metrics
- Custom attributes for request tracking

## 🧪 Testing

### Test Script

Run the comprehensive test script:

```bash
./examples/test-service.sh
```

The script tests:

- Service health and availability
- Gotenberg PDF service integration
- Template rendering with real files
- Authentication validation
- Error handling

### Unit Tests

```bash
make test
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes and test thoroughly
4. Run the test script: `./examples/test-service.sh`
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions:

1. Check the [Issues](../../issues) page
2. Create a new issue with detailed information
3. Include logs and error messages when reporting bugs
4. Use the test script to validate your setup

## 🗺️ Roadmap

- [x] Multi-cloud storage support (AWS S3, GCS)
- [x] Template rendering with Go templates
- [x] PDF generation via Gotenberg
- [x] Server authentication with bcrypt
- [x] Role-based access control
- [x] Docker containerization
- [x] Hot reload development setup
- [x] Comprehensive test suite
- [ ] PostgreSQL integration for persistent metadata
- [ ] File versioning system
- [ ] Webhook notifications
- [ ] Advanced RBAC with custom roles
- [ ] Metrics and monitoring dashboard
- [ ] File compression and optimization
- [ ] Batch operations support
- [ ] API rate limiting per user/server
- [ ] File encryption at rest
- [ ] Audit logging
- [ ] GraphQL API support
