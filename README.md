# Go Bazel Starter

A Go project demonstrating Bazel build system, HTTP client with retry logic, and Redis integration for caching and rate limiting.

## Features

- **HTTP Client**: Configurable HTTP client with retry logic and exponential backoff
- **Redis Integration**: HTTP response caching, rate limiting, and statistics
- **Bazel Build**: Fast, reproducible builds with dependency management
- **Testing**: Comprehensive test coverage for all packages
- **CI/CD**: GitLab CI configuration with Redis testing

## Prerequisites

- Go 1.22+
- Bazel 8.3+
- Redis (optional, for caching features)

## Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd go-bazel-starter
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run `make bazel.gazelle` to generate/update BUILD files.

## Building

### Using Bazel (Recommended)
```bash
# Build all packages (excluding OCI images)
bazel build //...

# Build only Go packages (faster)
bazel build //pkg/... //cmd/tool:tool

# Build specific package
bazel build //pkg/cache

# Build and run the tool
bazel run //cmd/tool:tool -- -url https://httpbin.org/get
```

### Using Go
```bash
# Build
go build ./cmd/tool

# Run
go run ./cmd/tool/main.go -url https://httpbin.org/get
```

## Testing

### Using Bazel
```bash
# Run all tests
bazel test //...

# Run specific package tests
bazel test //pkg/cache:all
```

### Using Go
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./pkg/cache/...
```

## Usage

### Basic HTTP Request
```bash
# Simple request
bazel run //cmd/tool:tool -- -url https://httpbin.org/get

# With custom retries
bazel run //cmd/tool:tool -- -url https://httpbin.org/get -retries 5
```

### Redis Caching
```bash
# Start Redis (if not running)
docker run -d --name redis -p 6379:6379 redis:7-alpine

# Use Redis cache
bazel run //cmd/tool:tool -- -url https://httpbin.org/json -cache

# Custom Redis server
bazel run //cmd/tool:tool -- -url https://httpbin.org/json -cache -redis localhost:6380
```

### Rate Limiting
```bash
# Set custom rate limit (requests per minute)
bazel run //cmd/tool:tool -- -url https://httpbin.org/delay/1 -cache -rate-limit 5
```

## Project Structure

```
go-bazel-starter/
├── cmd/
│   └── tool/           # Main application
├── pkg/
│   ├── cache/          # Redis caching and rate limiting
│   ├── httpx/          # HTTP client with retry logic
│   └── retry/          # Retry mechanism with backoff
├── BUILD.bazel         # Root build file
├── MODULE.bazel        # Bazel module configuration
└── go.mod              # Go module dependencies
```

## Redis Features

### HTTP Response Caching
- Automatic caching of HTTP responses
- Configurable TTL (Time To Live)
- Cache hit/miss statistics
- Automatic cache invalidation

### Rate Limiting
- Per-URL rate limiting
- Sliding window implementation
- Configurable limits and time windows
- Redis-based persistence

### Cache Management
- Cache statistics and monitoring
- Manual cache clearing
- Request count tracking
- Performance metrics

## CI/CD

The project includes GitLab CI configuration with:
- Multi-stage pipeline (test, build, artifact, redis-test)
- Redis service container for testing
- Automated testing and validation
- Artifact collection and storage

## Dependencies

- **github.com/redis/go-redis/v9**: Redis client
- **github.com/stretchr/testify**: Testing framework
- **Standard library**: HTTP, context, time, etc.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## Troubleshooting

### OCI Image Build Issues
If you encounter errors with `bazel build //...` related to OCI images:
```bash
# Use this command instead to build only Go packages
bazel build //pkg/... //cmd/tool:tool

# Or comment out the oci_image target in cmd/tool/BUILD.bazel
```

### Redis Connection Issues
If Redis is not available:
```bash
# The tool will continue working without cache
bazel run //cmd/tool:tool -- -url https://httpbin.org/get

# Start Redis for caching features
docker run -d --name redis -p 6379:6379 redis:7-alpine
```

## License

This project is licensed under the MIT License.