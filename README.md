# CSV Processing Service - Go Backend

A high-performance REST API service built with Go and Gin framework that processes CSV files by adding email validation flags to each row.

## 📋 Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Running the Application](#running-the-application)
- [API Documentation](#api-documentation)
- [Testing](#testing)
- [Usage Examples](#usage-examples)
- [Docker Deployment](#docker-deployment)
- [Troubleshooting](#troubleshooting)

## ✨ Features

- **Asynchronous Processing**: Non-blocking CSV file processing with job tracking
- **Email Validation**: Regex-based email validation for each row
- **Thread-Safe Operations**: Concurrent request handling with proper synchronization
- **Job Status Tracking**: Real-time job progress monitoring
- **File Management**: Automatic storage and cleanup of processed files
- **Error Handling**: Comprehensive error responses with appropriate HTTP status codes
- **RESTful API**: Clean API design following REST principles
- **High Performance**: Built with Go's efficient concurrency model

## 📋 Prerequisites

- **Go**: Version 1.19 or higher
- **Git**: For version control
- **Postman/cURL**: For API testing (optional)

### Check Prerequisites

```bash
# Check Go version
go version
# Should return: go version go1.19+

# Check Git
git --version
```

## 🔧 Installation

### Step 1: Clone Repository

```bash
git clone https://github.com/wanimudasir/demandscience.git
cd demandscience
```

### Step 2: Initialize Go Module

```bash
# Initialize module (if not already done)
go mod init demandscience

# Install dependencies
go get github.com/gin-gonic/gin@v1.9.1
go get github.com/google/uuid@v1.3.0
go get github.com/gin-contrib/cors@v1.4.0

# For testing
go get github.com/stretchr/testify/assert
```

### Step 3: Create Directory Structure

```bash
mkdir -p cmd internal/{handlers,services,models} processed_files
```

### Default Configuration

- **Server Port**: 8080
- **Storage Directory**: `processed_files/`
- **Max File Size**: 10MB
- **Supported Formats**: CSV

## 🚀 Running the Application

### Development Mode

```bash
# Run directly with Go
go run cmd/main.go

# Output:
# [GIN-debug] POST   /API/upload    --> demandscience/internal/handlers.(*CsvProcessorHandler).UploadFile-fm
# [GIN-debug] GET    /API/download/:id --> demandscience/internal/handlers.(*CsvProcessorHandler).DownloadFile-fm
# Starting server on :8080
```

### Production Mode

```bash
# Build binary
go build -o bin/demandscience cmd/main.go

# Run binary
./bin/demandscience

```

### Verify Server is Running

```bash
# Health check
curl http://localhost:8080/API/upload
# Expected: {"error":"No file provided"} with 400 status

```

## 📚 API Documentation

### Base URL

```
http://localhost:8080
```

### Endpoints Overview

| Method | Endpoint             | Description                       | Status Codes  |
| ------ | -------------------- | --------------------------------- | ------------- |
| POST   | `/API/upload`        | Upload CSV file for processing    | 200, 400      |
| GET    | `/API/download/{id}` | Check job status or download file | 200, 400, 423 |

---

#### 1. Upload CSV File

**Endpoint**: `POST /API/upload`

**Description**: Upload a CSV file for processing. The service will add a `has_email` column to each row.

**Request Format**:

```bash
curl -X POST \
  -F "file=@your-file.csv" \
  http://localhost:8080/API/upload
```

**Request Parameters**:

- `file` (required): CSV file via multipart/form-data

**Success Response** (200 OK):

```json
{
  "id": "a225eb00-0907-4273-92ca-5faadeefae5f"
}
```

**Error Responses** (400 Bad Request):

```json
{
  "error": "No file provided"
}
```

```json
{
  "error": "File is empty"
}
```

```json
{
  "error": "Invalid file type. Only CSV and text files are allowed"
}
```

---

#### 2. Check Job Status / Download File

**Endpoint**: `GET /API/download/{id}`

**Description**: Check processing status or download the processed file.

**Status Check Request**:

```bash
curl -X GET http://localhost:8080/API/download/a225eb00-0907-4273-92ca-5faadeefae5f
```

**Download File Request**:

```bash
curl -X GET \
  -o processed_file.csv \
  "http://localhost:8080/API/download/a225eb00-0907-4273-92ca-5faadeefae5f"
```

**Response Types**:

**In Progress** (423 Locked):

```json
{
  "status": "in_progress",
  "message": "Job is still in progress",
  "completed": false
}
```

**Completed** (200 OK):

```json
{
  "id": "a225eb00-0907-4273-92ca-5faadeefae5f",
  "status": "completed",
  "message": "Job completed successfully"
}
```

**File Download** (200 OK):

```json
{
  "id": "a225eb00-0907-4273-92ca-5faadeefae5f",
  "status": "completed",
  "message": "Job completed successfully",
  "file_data": "blob data......."
}
```

**Error Responses** (400 Bad Request):

```json
{
  "error": "Invalid job ID"
}
```

## 🧪 Testing

### Run Unit Tests

```bash
# Run all tests
go test ./internal/handlers -v

# Run specific test
go test ./internal/handlers -run TestUploadValidFile -v

# Run tests with coverage
go test ./internal/handlers -cover -v

```

### Manual Testing

### Testing with Postman

#### Import Collection

1. **Create Postman Collection**: "CSV Processing API"
2. **Add Environment Variable**:
   - Variable: `base_url`
   - Value: `http://localhost:8080`

#### Test Requests

**1. Upload File**

- Method: `POST`
- URL: `{{base_url}}/API/upload`
- Body: form-data, key=`file`, type=File
- Tests Script:

```javascript
const response = pm.response.json();
pm.globals.set("job_id", response.id);

pm.test("Upload successful", function () {
  pm.response.to.have.status(200);
  pm.expect(response).to.have.property("id");
});
```

**2. Check Status**

- Method: `GET`
- URL: `{{base_url}}/API/download/{{job_id}}`
- Tests Script:

```javascript
pm.test("Valid status response", function () {
  pm.expect(pm.response.code).to.be.oneOf([200, 423, 400]);
});
```

**3. Download File**

- Method: `GET`
- URL: `{{base_url}}/API/download/{{job_id}}`

## 📊 Usage Examples

### Example 1: Basic CSV Processing

**Input CSV** (`input.csv`):

```csv
name,email,phone
John Doe,john.doe@example.com,123-456-7890
Jane Smith,invalid-email,098-765-4321
Bob Wilson,bob@company.org,555-123-4567
```

**API Call**:

```bash
# Upload
curl -X POST -F "file=@input.csv" http://localhost:8080/API/upload

# Response: {"id":"job-123"}

# Check status
curl -X GET http://localhost:8080/API/download/job-123

# Download when ready
curl -X GET -o output.csv "http://localhost:8080/API/download/job-123"
```

**Output CSV** (`output.csv`):

```csv
name,email,phone,has_email
John Doe,john.doe@example.com,123-456-7890,true
Jane Smith,invalid-email,098-765-4321,false
Bob Wilson,bob@company.org,555-123-4567,true
```

### Example 2: Error Handling

```bash
# Test empty file
curl -X POST -F "file=@empty.csv" http://localhost:8080/API/upload
# Response: {"error":"File is empty"}

# Test invalid job ID
curl -X GET http://localhost:8080/API/download/invalid-id
# Response: {"error":"Invalid job ID"}

# Test invalid file type
curl -X POST -F "file=@document.pdf" http://localhost:8080/API/upload
# Response: {"error":"Invalid file type. Only CSV and text files are allowed"}
```

### Example 3: Automation Script

```bash
#!/bin/bash

# Automated CSV processing script
process_csv() {
    local file=$1
    echo "Processing: $file"

    # Upload file
    RESPONSE=$(curl -s -X POST -F "file=@$file" http://localhost:8080/API/upload)
    JOB_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

    if [ -z "$JOB_ID" ]; then
        echo "Upload failed: $RESPONSE"
        return 1
    fi

    echo "Job ID: $JOB_ID"

    # Wait for completion
    while true; do
        STATUS=$(curl -s http://localhost:8080/API/download/$JOB_ID)
        if echo "$STATUS" | grep -q "completed.*true"; then
            echo "Processing completed!"
            break
        elif echo "$STATUS" | grep -q "error"; then
            echo "Processing failed: $STATUS"
            return 1
        else
            echo "Still processing..."
            sleep 2
        fi
    done

    # Download processed file
    OUTPUT_FILE="${file%.*}_processed.csv"
    curl -s -o "$OUTPUT_FILE" "http://localhost:8080/API/download/$JOB_ID"
    echo "Downloaded: $OUTPUT_FILE"
}

# Usage
process_csv "test-data/sample.csv"
```

## 🐳 Docker Deployment

### Build Docker Image

```bash
# Create Dockerfile
cat << 'EOF' > Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o demandscience cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/demandscience .
RUN mkdir -p processed_files

EXPOSE 8080
CMD ["./demandscience"]
EOF

# Build image
docker build -t demandscience .

# Run container
docker run -p 8080:8080 demandscience
```

### Docker Compose

```yaml
# docker-compose.yml
version: "3.8"

services:
  demandscience:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./processed_files:/root/processed_files
    environment:
      - GIN_MODE=release
    restart: unless-stopped

volumes:
  processed_files:
```

```bash
# Run with Docker Compose
docker-compose up --build

# Run in background
docker-compose up -d --build

# View logs
docker-compose logs -f
```

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
go test ./... -v

# Run handler tests only
go test ./internal/handlers -v

# Run with coverage
go test ./internal/handlers -cover

# Run specific test
go test ./internal/handlers -run TestUploadValidFile -v

# Run tests with race detection
go test ./internal/handlers -race -v
```

### Integration Tests

```bash
# Create integration test script
cat << 'EOF' > integration_test.sh
#!/bin/bash

echo "=== CSV Processing API Integration Tests ==="

# Start server in background
go run cmd/main.go &
SERVER_PID=$!
sleep 3

# Test 1: Valid file upload
echo "Test 1: Valid File Upload"
RESPONSE=$(curl -s -X POST -F "file=@test-data/sample.csv" http://localhost:8080/API/upload)
JOB_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

if [ -n "$JOB_ID" ]; then
    echo "✅ Upload successful: $JOB_ID"
else
    echo "❌ Upload failed: $RESPONSE"
fi

# Test 2: Status check
echo "Test 2: Status Check"
STATUS=$(curl -s http://localhost:8080/API/download/$JOB_ID)
echo "Status: $STATUS"

# Test 3: Download after completion
sleep 5
echo "Test 3: File Download"
curl -s -o test_output.csv "http://localhost:8080/API/download/$JOB_ID?download=true"

if [ -f test_output.csv ]; then
    echo "✅ File downloaded successfully"
    echo "Content preview:"
    head -3 test_output.csv
else
    echo "❌ Download failed"
fi

# Cleanup
kill $SERVER_PID
rm -f test_output.csv
EOF

chmod +x integration_test.sh
./integration_test.sh
```

### Load Testing

```bash
# Create load test
cat << 'EOF' > load_test.go
package main

import (
    "bytes"
    "fmt"
    "mime/multipart"
    "net/http"
    "sync"
    "time"
)

func main() {
    const numRequests = 10
    const serverURL = "http://localhost:8080"

    csvContent := "name,email\nJohn,john@test.com\nJane,jane@test.com"

    var wg sync.WaitGroup
    start := time.Now()

    for i := 0; i < numRequests; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            var buf bytes.Buffer
            writer := multipart.NewWriter(&buf)
            part, _ := writer.CreateFormFile("file", fmt.Sprintf("test%d.csv", id))
            part.Write([]byte(csvContent))
            writer.Close()

            resp, err := http.Post(serverURL+"/API/upload",
                writer.FormDataContentType(), &buf)
            if err != nil {
                fmt.Printf("Request %d failed: %v\n", id, err)
                return
            }
            defer resp.Body.Close()

            fmt.Printf("Request %d: Status %d\n", id, resp.StatusCode)
        }(i)
    }

    wg.Wait()
    fmt.Printf("Load test completed in %v\n", time.Since(start))
}
EOF

# Run load test
go run load_test.go
```

## 📝 Usage Examples

### Example 1: Processing Contact List

```bash
# Sample contact list
cat << EOF > contacts.csv
first_name,last_name,email,phone,department
John,Doe,john.doe@company.com,123-456-7890,Engineering
Jane,Smith,jane.smith.invalid,098-765-4321,Marketing
Bob,Wilson,bob@company.org,555-123-4567,Sales
Mary,Johnson,,999-888-7777,HR
EOF

# Process the file
RESPONSE=$(curl -s -X POST -F "file=@contacts.csv" http://localhost:8080/API/upload)
JOB_ID=$(echo $RESPONSE | jq -r '.id')

# Wait for completion
while true; do
    STATUS=$(curl -s http://localhost:8080/API/download/$JOB_ID)
    if echo "$STATUS" | grep -q '"completed":true'; then
        break
    fi
    echo "Processing..."
    sleep 2
done

# Download result
curl -s -o contacts_processed.csv "http://localhost:8080/API/download/$JOB_ID?download=true"
```

### Example 2: Batch Processing

```bash
#!/bin/bash

# Process multiple CSV files
for file in *.csv; do
    echo "Processing: $file"

    RESPONSE=$(curl -s -X POST -F "file=@$file" http://localhost:8080/API/upload)
    JOB_ID=$(echo $RESPONSE | jq -r '.id')

    # Store job ID for later download
    echo "$JOB_ID:$file" >> job_list.txt
done

echo "All uploads initiated. Checking completion..."

# Check all jobs
while read line; do
    JOB_ID=$(echo $line | cut -d':' -f1)
    FILENAME=$(echo $line | cut -d':' -f2)

    # Wait for completion
    while true; do
        STATUS=$(curl -s http://localhost:8080/API/download/$JOB_ID)
        if echo "$STATUS" | grep -q '"completed":true'; then
            curl -s -o "processed_$FILENAME" "http://localhost:8080/API/download/$JOB_ID?download=true"
            echo "✅ $FILENAME processed"
            break
        fi
        sleep 1
    done
done < job_list.txt

rm job_list.txt
```

## 🔧 Configuration Options

### Custom Server Port

```bash
# Method 1: Environment variable
export SERVER_PORT=9090
go run cmd/main.go

# Method 2: Command line flag (requires code modification)
go run cmd/main.go -port=9090
```

### Custom Storage Directory

```bash
# Create custom storage
mkdir custom_storage

# Set environment variable
export STORAGE_DIR=custom_storage
go run cmd/main.go
```

### Enable Debug Mode

```bash
# Set Gin to debug mode
export GIN_MODE=debug
go run cmd/main.go

# Enable verbose logging
export LOG_LEVEL=debug
go run cmd/main.go
```

## 🚨 Troubleshooting

### Common Issues

#### 1. Server Won't Start

```bash
# Check if port is already in use
lsof -i :8080

# Kill existing process
kill -9 $(lsof -t -i:8080)

# Try different port
SERVER_PORT=9090 go run cmd/main.go
```

#### 2. File Upload Fails

```bash
# Check file permissions
ls -la your-file.csv

# Check file size
du -h your-file.csv

# Verify file format
file your-file.csv
```

#### 3. Job Status Stuck

```bash
# Check server logs
# Look for error messages in console output

# Check processed_files directory
ls -la processed_files/

# Test with smaller file
echo "name,email\ntest,test@test.com" > small_test.csv
```

### Debugging

#### Enable Detailed Logging

```go
// Add to main.go
import "log"

func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)

    // ... rest of code
}
```

#### Check File Processing

```bash
# Monitor processed_files directory
watch -n 1 'ls -la processed_files/'

# Check file creation timestamps
stat processed_files/*
```

## 🔒 Security Considerations

### File Upload Security

- **File Size Limit**: 10MB maximum
- **File Type Validation**: Only CSV files
- **Path Traversal Prevention**: Safe file naming
- **Memory Management**: Streaming file processing
