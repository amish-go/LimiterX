# LimiterX

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)
![Performance](https://img.shields.io/badge/Performance-High-red?style=for-the-badge)

A high-performance API rate limiter with fixed and sliding window algorithms.

</div>

---

## Features

- Fixed Window & Sliding Window algorithms
- Per-IP rate limiting
- In-memory cache with automatic cleanup
- Metrics and health check endpoints
- Configurable via environment variables

## Installation

```bash
git clone <repository-url>
cd limiterx
go mod init limiterx
go mod tidy
go run main.go
```

## Configuration

- `PORT` - Server port (default: `8080`)
- `ALGORITHM` - `fixed` or `sliding` (default: `sliding`)
- `REQUESTS_LIMIT` - Max requests per window (default: `100`)
- `WINDOW_SECONDS` - Time window in seconds (default: `60`)

```bash
# Example
ALGORITHM=fixed REQUESTS_LIMIT=50 WINDOW_SECONDS=30 go run main.go
```

## API Endpoints

**GET /api/resource** - Protected resource with rate limiting

**GET /metrics** - Rate limiter statistics

**GET /health** - Health check



## License

MIT License

## Contact

Email: contact.amish@yahoo.com
