<p align="center">
  <h1 align="center">YARD Backend</h1>
  <h3 align="center">Yet Another Reforge Database - Backend API</h3>
</p>

<p align="center">
  A high-performance Go backend API for <strong>YARD</strong> (Yet Another Reforge Database), providing reforge stone data, item information, and market prices for Hypixel SkyBlock players.
</p>

**Frontend**: [YARD-Frontend](../YARD-Frontend)

## Table of Contents

- [Overview](#overview)
- [Requirements](#requirements)
- [Installation](#installation)
  - [System Dependencies](#system-dependencies)
  - [Go Installation](#go-installation)
  - [Redis Installation](#redis-installation)
- [Configuration](#configuration)
- [Development](#development)
- [API Endpoints](#api-endpoints)
- [Data Sources](#data-sources)
- [Common Issues](#common-issues)
- [Docker](#docker)
  - [Monitoring with Prometheus and Grafana](#monitoring-with-prometheus-and-grafana)

## Overview

YARD Backend is a RESTful API service that:
- Fetches reforge stone data from the Hypixel API
- Retrieves live market prices from SkyCofl (Auction House and Bazaar)
- Loads reforge effects and stats from NotEnoughUpdates repository
- Stores data in Redis for fast access
- Provides item texture rendering and upscaling
- Schedules automatic data updates every 6 hours

## Requirements

- Go 1.21 or later
- Redis 7.0 or later
- Git (for submodule initialization)

## Installation

The following instructions are written for Unix-based systems. Adjust package manager commands accordingly for other distributions.

### System Dependencies

Install essential build tools:

**macOS (using Homebrew):**
```bash
brew install git
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt update
sudo apt install build-essential git
```

**Linux (Arch):**
```bash
sudo pacman -S base-devel git
```

### Go Installation

**macOS (using Homebrew):**
```bash
brew install go
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt install golang-go
```

**Linux (Arch):**
```bash
sudo pacman -S go
```

**Or download from [golang.org](https://golang.org/dl/)**

Verify the installation:
```bash
go version
```

### Redis Installation

**macOS (using Homebrew):**
```bash
brew install redis
brew services start redis
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt install redis-server
sudo systemctl enable redis-server
sudo systemctl start redis-server
```

**Linux (Arch):**
```bash
sudo pacman -S redis
sudo systemctl enable redis
sudo systemctl start redis
```

Verify Redis is running:
```bash
redis-cli ping
```

Expected output: `PONG`

## Configuration

### Environment Variables

The backend uses environment variables for configuration. You can set them directly or use a `.env` file.

Create a `.env` file in the project root directory. Use `.env.example` as a template:

```bash
cp .env.example .env
```

Edit the `.env` file with your configuration:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `REDIS_HOST` | Redis server hostname | `localhost` | No |
| `REDIS_PORT` | Redis server port | `6379` | No |
| `REDIS_PASSWORD` | Redis authentication password | - | No |
| `API_URL` | Hypixel API endpoint | `https://api.hypixel.net/v2/resources/skyblock/items` | No |
| `SKYCOFL_URL` | SkyCofl API base URL | `https://sky.coflnet.com` | No |
| `NEU_REPO_PATH` | Path to NotEnoughUpdates repository | `NotEnoughUpdates-REPO` | No |
| `ALLOWED_ORIGIN` | Allowed CORS origin(s). Use `*` for all origins (dev only) or specific domain(s) comma-separated for production | `*` | No |
| `METRICS_ENABLED` | Enable Prometheus metrics collection. Set to `true` or `1` to enable | `false` | No |
| `METRICS_IP_WHITELIST` | Optional IP whitelist for `/metrics` endpoint. Comma separated IPs or CIDR ranges (e.g., `127.0.0.1,172.18.0.0/24`). Leave empty to allow all IPs | - | No |

### Example .env File

```env
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

API_URL=https://api.hypixel.net/v2/resources/skyblock/items
SKYCOFL_URL=https://sky.coflnet.com

NEU_REPO_PATH=NotEnoughUpdates-REPO

# For production, set to your frontend domain(s):
# ALLOWED_ORIGIN=https://yourdomain.com,https://www.yourdomain.com
# For development, leave as * or omit
ALLOWED_ORIGIN=*

# Optional: Enable Prometheus metrics collection
# METRICS_ENABLED=true
# Optional: Restrict metrics endpoint to specific IPs (comma-separated)
# METRICS_IP_WHITELIST=127.0.0.1,172.18.0.0/24
```

## Development

Clone the repository with submodules:
```bash
git clone --recurse-submodules https://github.com/YARDatabase/YARD-Backend.git
cd YARD-Backend
```

If you have already cloned the repository without submodules:
```bash
git submodule update --init --recursive
```

Download Go dependencies:
```bash
go mod tidy
go mod download
```

**Note**: If you're using the monitoring features, `go mod tidy` will download the Prometheus dependencies.

Run the application:
```bash
go run main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Health Check

**GET** `/health`

Returns the health status of the backend service.

**Response:**
```json
{
  "status": "ok",
  "message": "YARD Backend is running",
  "time": "2026-01-01T12:00:00Z"
}
```

### Get Reforge Stones

**GET** `/api/reforge-stones`

Returns all stored reforge stones with their data.

**Response:**
```json
{
  "success": true,
  "count": 10,
  "lastUpdated": "2026-01-01T12:00:00Z",
  "reforgeStones": [
    {
      "id": "MANDRAA",
      "name": "Mandraa",
      "tier": "RARE",
      "material": "SKULL_ITEM",
      "category": "REFORGE_STONE",
      "npc_sell_price": 1
    }
  ]
}
```

### Get Item Image

**GET** `/api/item/{itemId}`

Returns a PNG image of the specified item. The item ID should be in uppercase with underscores (e.g., `MANDRAA`, `PERFECT_RUBY`).

**Response:** PNG image (256x256 pixels)

**Example:**
```
GET /api/item/MANDRAA
```

### Get Item Image by Data

**GET** `/api/item-data/{itemId}`

Returns a PNG image of the specified item based on item data. Supports custom item data parameters.

**Response:** PNG image (256x256 pixels)

### Metrics Endpoint

**GET** `/metrics`

Returns Prometheus-formatted metrics. Only available when `METRICS_ENABLED=true` in your environment configuration.

**Response:** Prometheus metrics in text format

**Example:**
```
GET /metrics
```

## Data Sources

YARD Backend integrates with multiple data sources:

- **Hypixel API**: Official reforge stone data and item information
- **SkyCofl**: Live market prices from Auction House and Bazaar
- **NotEnoughUpdates Repository**: Reforge effects, stat scaling, and item lore

## Security Features

### CORS Configuration

The API uses configurable CORS (Cross-Origin Resource Sharing) to control which origins can access the API:

- **Development**: Defaults to `*` (allows all origins) for easy local testing
- **Production**: Set `ALLOWED_ORIGIN` to your frontend domain(s) for security

**Example for production:**
```env
ALLOWED_ORIGIN=https://yard.example.com,https://www.yard.example.com
```

Multiple origins can be specified as comma-separated values. The API will only accept requests from these origins.

### Rate Limiting

The API implements rate limiting to prevent abuse:

- **Limit**: 60 requests per minute per client IP
- **Window**: 1 minute sliding window
- **Response**: Returns `429 Too Many Requests` when exceeded
- **Headers**: Includes `Retry-After` header indicating when to retry

Rate limiting is applied to all API endpoints except `/health`. The health check endpoint is excluded to allow monitoring tools to check server status.

## Data Storage

Reforge stones are stored in Redis with the following structure:
- `reforge_stone:{id}` - Individual stone data (JSON)
- `reforge_stones:ids` - Set of all stone IDs
- `reforge_stones:last_updated` - Timestamp of last update
- `reforge_stones:count` - Total count of stones

## Testing

Run tests with:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -cover ./...
```

## Common Issues

### Submodule Not Initialized

If the `NotEnoughUpdates-REPO` directory is empty:
```bash
git submodule update --init --recursive
```

### Redis Connection Error

If you see Redis connection errors:
1. Verify Redis is running: `redis-cli ping`
2. Check the `REDIS_HOST` and `REDIS_PORT` environment variables
3. Ensure Redis is accessible from your network

### Rate Limiting

The backend implements rate limiting for API requests to respect Hypixel API limits. The default minimum delay between requests is 700ms.

## Docker

The backend can be run using Docker Compose, which will automatically set up Redis and the backend service.

### Prerequisites

- Docker and Docker Compose installed on your system

### Quick Start

1. Create a `.env` file in the project root (see [Configuration](#configuration) section for details)

2. Build and start the services:
```bash
docker-compose up --build
```

3. The backend will be available at `http://localhost:8080`

### Docker Compose Commands

- Start services in detached mode:
```bash
docker-compose up -d
```

- Stop services:
```bash
docker-compose down
```

- View logs:
```bash
docker-compose logs -f
```

- Rebuild and restart:
```bash
docker-compose up --build -d
```

### Services

The docker-compose setup includes:
- **Redis**: Data storage service (port 6379)
- **Backend**: API service (port 8080)

The backend service automatically connects to Redis using the service name `redis` as the hostname.

### Monitoring with Prometheus and Grafana

The backend includes optional monitoring support with Prometheus and Grafana for tracking API metrics.

#### Prerequisites

1. Enable metrics in your `.env` file:
   ```env
   METRICS_ENABLED=true
   ```

2. Optionally restrict metrics endpoint access:
   ```env
   METRICS_IP_WHITELIST=127.0.0.1,172.18.0.0/24
   ```

#### Quick Start

Start all services including monitoring:

```bash
docker-compose -f docker-compose.monitoring.yml up -d
```

This will start:
- **Redis** on port 6379
- **Backend** on port 8080
- **Prometheus** on port 9090
- **Grafana** on port 3001

#### Accessing Monitoring

- **Grafana**: http://localhost:3001
  - Default username: `admin`
  - Default password: `admin`
  - Pre-configured dashboard: "YARD API Monitoring"
- **Prometheus**: http://localhost:9090
  - View raw metrics and query interface
- **Metrics Endpoint**: http://localhost:8080/metrics
  - Prometheus-formatted metrics (only available when `METRICS_ENABLED=true`)

#### Available Metrics

- `yard_http_requests_total` - Total HTTP requests by method, endpoint, and status code
- `yard_http_request_errors_total` - Total HTTP errors by method, endpoint, and status code
- `yard_http_requests_by_country_total` - Total requests by country and endpoint

#### Stopping Monitoring Services

```bash
docker-compose -f docker-compose.monitoring.yml down
```

To remove monitoring data volumes:

```bash
docker-compose -f docker-compose.monitoring.yml down -v
```

## License

This project is licensed under the GPL-3.0 License.
