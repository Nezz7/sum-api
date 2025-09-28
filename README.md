# Sum API

A lightweight Go HTTP server that calculates the sum of two numbers.

## Build & Run

```bash

# Run the server
make run

# Run linting and formatting
make lint

# Run tests with benchmarks
make test

# Run tests with coverage
make coverage

# Run SAST (Static Application Security Testing)
make sast

# Run DAST (Dynamic Application Security Testing)
make dast

# Test API endpoints
make curl

# Build the application
make build

# Clean build artifacts
make clean
```

## API Documentation

### Endpoints

#### GET /sum
Calculates the sum of two numbers.

**Parameters:**
- `a` (required): First number
- `b` (required): Second number

**Example:**
```bash
curl "http://localhost:8080/sum?a=5&b=3"
```

**Response:**
```json
{"result": 8}
```

**Error Responses:**
- `400 Bad Request`: Missing or invalid parameters
  ```json
  {"error": "missing parameter 'a'"}
  ```
- `400 Bad Request`: Integer overflow
  ```json
  {"error": "integer overflow: 9223372036854775807 + 1 exceeds maximum value"}
  ```
- `405 Method Not Allowed`: Invalid HTTP method
  ```json
  {"error": "method not allowed"}
  ```