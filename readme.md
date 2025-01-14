# Sensor Data API

This project provides an API to accept sensor data from different devices, validate the message type, and store valid data into a Redis database. 

# Dependencies

- **[Echo](https://echo.labstack.com/)**: Web framework.
- **[Go-Redis](https://github.com/go-redis/redis)**: Redis client for Go.

## Install dependencies
- For windows:
```
.\init.cmd
```

- For Linux/Mac:
```
.\init.sh
```

- Manual
```
go get github.com/labstack/echo/v4
go get github.com/redis/go-redis/v9
```

## Prerequisites

- **Redis Server**: A running Redis instance is required.

    - Use temporary docker server for test purpose 
    ```
    docker run --name testRedis -p 6379:6379 -d redis
    ```

## Configuration

- `--redis-url`: Address of the Redis server (default: `localhost:6379`).
- `--redis-password`: Redis password (can be set via the `REDIS_PASSWORD` environment variable). Empty by default.

## Running
Start the Application

```bash
go run main.go --redis-url=localhost:6379 --redis-password=yourpassword
```


Start the Application with docker

```bash
docker run --name testRedis -p 6379:6379 -d redis
go run main.go
```

## Endpoints

### 1. **POST /process**
  Process and stores sensor data in Redis.

#### Request Body example
```json
{
  "time": "2025-01-01T10:00:00Z",
  "device_id": "1234",
  "device_type": "A",
  "uptime": 123,
  "temp": 23.5
}
```

### 2. **GET /getDataById?id=id**
  Get sensor data by device ID