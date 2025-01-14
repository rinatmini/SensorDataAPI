package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// SensorData represents the structure of the sensor data received from the client.
type SensorData struct {
	Time       string  `json:"time"`        // Timestamp of the sensor data
	DeviceId   string  `json:"device_id"`   // Unique identifier for the device
	DeviceType string  `json:"device_type"` // Type of the device (A or B)
	Uptime     int     `json:"uptime"`      // Uptime of the device in seconds
	Temp       float32 `json:"temp"`        // Temperature recorded by the sensor
}

// IsValidType checks if the device type is either A or B.
func (s SensorData) IsValidType() bool {
	return s.DeviceType == "A" || s.DeviceType == "B"
}

func main() {

	redisAddress := flag.String("redis-url", "localhost:6379", "Redis server address")
	redisPassword := flag.String("redis-password", os.Getenv("REDIS_PASSWORD"), "Redis server password")

	flag.Parse()

	rdb, err := getRedisClient(*redisPassword, *redisAddress)

	if err != nil {
		log.Fatalf("Failed to initialize Redis client: %v", err)
		flag.PrintDefaults()
		os.Exit(1)
	}

	e := echo.New()
	e.POST("/process", func(c echo.Context) error {
		return saveSensor(c, rdb)
	})
	e.GET("/getDataById", func(c echo.Context) error {
		return getSensor(c, rdb)
	})
	e.Logger.Fatal(e.Start(":8080"))
}

// saveSensor processes the incoming sensor data, validates it, and stores it in Redis
func saveSensor(c echo.Context, redisClient *redis.Client) error {
	sensorDataToProcess := new(SensorData)

	err := c.Bind(sensorDataToProcess)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unable to get sensor data from the request body: %v", err))
	}

	err = validateSensorData(sensorDataToProcess)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = saveToRedis(redisClient, sensorDataToProcess, c.Request().Context())

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error on saving user in the cache: %v", err)
	}

	return c.NoContent(http.StatusCreated)
}

// validateSensorData checks if the sensor data is valid based on device type
func validateSensorData(s *SensorData) (e error) {
	if !s.IsValidType() {
		return fmt.Errorf("device type %s is not supported", s.DeviceType)
	}

	return nil
}

// saveToRedis serializes the sensor data and stores it in Redis
func saveToRedis(rdb *redis.Client, sensorData *SensorData, ctx context.Context) (e error) {
	dataToSave, err := json.Marshal(sensorData)

	if err != nil {
		return fmt.Errorf("fatal error on marshalling the sensor data for device %s: %v", sensorData.DeviceId, err)
	}

	err = rdb.Set(ctx, sensorData.DeviceId, dataToSave, 0).Err()

	if err != nil {
		return fmt.Errorf("fatal error on saving the device id %s data in the cache: %v", sensorData.DeviceId, err)
	}

	return nil
}

// getRedisClient initializes a Redis client with the provided credentials
func getRedisClient(password, url string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       0,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return rdb, nil
}

// getSensor handles the GET request to retrieve sensor data by device ID
func getSensor(c echo.Context, rdb *redis.Client) error {
	deviceId := c.QueryParam("id")

	if deviceId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Device 'id' is missing")
	}

	sensorData, err := getSensorDataById(deviceId, rdb, c.Request().Context())

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Couldn't get the Sensor data for device %s from the cache. %v", deviceId, err))
	}

	return c.JSON(http.StatusOK, sensorData)
}

// getSensorDataById retrieves sensor data from Redis by device ID
func getSensorDataById(id string, rdb *redis.Client, ctx context.Context) (*SensorData, error) {

	fromDB, err := rdb.Get(ctx, id).Bytes()

	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("sensor data for device id with %s not found", id)
		}

		return nil, fmt.Errorf("fatal error on retrieiving the sensor data for device id %s from the cache: %v", id, err)
	}

	var sensorData SensorData

	err = json.Unmarshal(fromDB, &sensorData)

	if err != nil {
		return nil, fmt.Errorf("fatal error on reading the sensor data for device id %s from cache: %v", id, err)
	}

	return &sensorData, nil
}
