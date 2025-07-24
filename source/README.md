# File Manager Microservice

A high-performance Go microservice for file management with multi-cloud storage support, template rendering, and PDF generation capabilities.

## 🚀 Features

- **Multi-Cloud Storage**: Support for AWS S3 and Google Cloud Storage with configurable default provider
- **Template Rendering**: Dynamic HTML template rendering with JSON data injection
- **PDF Generation**: Convert rendered templates to PDF using Gotenberg service
- **Server Authentication**: PIN-based authentication system for server-to-server communication
- **Role-Based Access Control**: Fine-grained permissions with admin, calculator, and analytics roles
- **Metadata Management**: In-memory metadata store for file information
- **Telemetry & Logging**: Structured logging with OpenTelemetry support
- **Rate Limiting**: Built-in rate limiting for API protection
- **Health Checks**: Ready-to-use health check endpoints
- **Docker Support**: Containerized deployment with Docker

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

Create your environment configuration. The service supports configuration via environment variables or a JSON secrets string.

#### Environment Variables

```bash
export AWS_REGION="us-east-1"
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export GCP_PROJECT_ID="your-project-id"
export GCP_CREDENTIALS_FILE="path/to/credentials.json"
export DEFAULT_CLOUD="aws"  # or "gcp"
export BUCKET_NAME="your-bucket-name"
export REPLICATE_TO_ALL_CLOUDS="false"
```

#### JSON Secrets (Alternative)

```bash
export SECRETS='{"AWS_ACCESS_KEY_ID":"your-key","AWS_SECRET_ACCESS_KEY":"your-secret","BUCKET_NAME":"your-bucket"}'
```

### 4. Start Gotenberg Service (Required for PDF Generation)

```bash
docker run --rm -p 3001:3000 gotenberg/gotenberg:7
```

### 5. Run the Service

```bash
# Development mode with hot reload
make dev

# Or build and run manually
go build -o bin/file-manager
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
# Build and push to ECR
make build-push
```

## 📚 API Documentation

### Authentication

All API endpoints (except health check) require server authentication:

**Headers:**

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
  -F "template=@invoice.html" \
  -F 'jsonData={"InvoiceNumber":"INV-001","CustomerName":"John Doe","Total":"2000.00"}'
```

### 2. Go Client Example

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

    // Create multipart form
    var b bytes.Buffer
    writer := multipart.NewWriter(&b)

    // Add template file
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

    // Add JSON data
    writer.WriteField("jsonData", `{"name":"John","amount":"100"}`)
    writer.Close()

    // Create request
    req, err := http.NewRequest("POST", url, &b)
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", writer.FormDataContentType())
    req.Header.Set("X-Server-ID", "calculator-server")
    req.Header.Set("X-PIN", "123")

    // Send request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}
```

### 3. Python Client Example

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
# Start development server with hot reload
make dev

# Build Docker image
make build

# Run tests
make test

# Build and push to ECR
make build-push

# Run locally
make run

# Show all available commands
make help
```

### Project Structure

```
source/
├── application/          # Application layer
│   ├── app.go           # Main app initialization
│   ├── config.go        # Config wrapper
│   └── routes.go        # Route definitions
├── auth/                # Authentication
│   └── server_auth.go   # Server auth middleware
├── config/              # Configuration
│   └── config.go        # Config management
├── domain/              # Domain layer
│   └── file/           # File domain
│       ├── hanlder.go  # File handlers
│       ├── model.go    # File models
│       └── repo.go     # File repository
├── storage/             # Storage layer
│   ├── aws_s3.go       # AWS S3 adapter
│   ├── gcs.go          # Google Cloud Storage adapter
│   ├── manager.go      # Storage manager
│   └── storage.go      # Storage interface
├── telemetry/           # Observability
│   └── logger.go       # Logging configuration
└── utils/               # Utilities
    ├── aws_helper.go    # AWS utilities
    ├── clerk_helper.go  # Clerk utilities
    └── parse_template.go # Template parsing
```

## 🔐 Security

### Server Authentication

The service uses a PIN-based authentication system:

- **calculator-server**: PIN `123` (calculator role)
- **analytics-server**: PIN `456` (analytics role)
- **admin-server**: PIN `789` (admin role)

> ⚠️ **Production Note**: Change default PINs and use secure password hashing

### Rate Limiting

Built-in rate limiting allows 1000 requests per time window per client.

## 🌐 Multi-Cloud Storage

### Supported Providers

- **AWS S3**: Primary cloud storage
- **Google Cloud Storage**: Alternative/backup storage

### Configuration

Set `DEFAULT_CLOUD` to `aws` or `gcp` to choose your primary storage provider.

## 📊 Monitoring & Logging

The service includes:

- **Structured Logging**: JSON-formatted logs with context
- **Request Tracing**: Unique request IDs for tracking
- **Health Checks**: `/health` endpoint for monitoring
- **OpenTelemetry**: Ready for distributed tracing

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions:

1. Check the [Issues](../../issues) page
2. Create a new issue with detailed information
3. Include logs and error messages when reporting bugs

## 🗺️ Roadmap

- [ ] PostgreSQL integration for persistent metadata
- [ ] File versioning system
- [ ] Webhook notifications
- [ ] Advanced RBAC with custom roles
- [ ] Metrics and monitoring dashboard
- [ ] File compression and optimization
- [ ] Batch operations support
