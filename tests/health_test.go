package test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"

	"auth-service/internal/handler"
)

func TestLivenessProbe(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Mock DB and Redis (not needed for liveness - it only checks if app is alive)
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	redisClient, _ := redismock.NewClientMock()

	// Create handler
	healthHandler := handler.NewHealthHandler(db, redisClient)

	// Create test router
	router := gin.New()
	router.GET("/health/live", healthHandler.LivenessProbe)

	// Test liveness probe
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/live", nil)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "alive", response["status"])
	assert.NotEmpty(t, response["timestamp"])
}

func TestReadinessProbe_AllHealthy(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Mock DB
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer db.Close()

	// Expect DB ping
	mock.ExpectPing()

	// Mock Redis
	redisClient, redisMock := redismock.NewClientMock()
	redisMock.ExpectPing().SetVal("PONG")

	// Create handler
	healthHandler := handler.NewHealthHandler(db, redisClient)

	// Create test router
	router := gin.New()
	router.GET("/health/ready", healthHandler.ReadinessProbe)

	// Test readiness probe
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/ready", nil)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ready", response["status"])
	assert.NotEmpty(t, response["timestamp"])

	dependencies := response["dependencies"].(map[string]interface{})
	assert.Equal(t, "healthy", dependencies["database"])
	assert.Equal(t, "healthy", dependencies["redis"])

	// Verify all expectations met
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestReadinessProbe_DatabaseUnhealthy(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Mock DB with error
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer db.Close()

	// Expect DB ping to fail
	mock.ExpectPing().WillReturnError(sql.ErrConnDone)

	// Mock Redis (healthy)
	redisClient, redisMock := redismock.NewClientMock()
	redisMock.ExpectPing().SetVal("PONG")

	// Create handler
	healthHandler := handler.NewHealthHandler(db, redisClient)

	// Create test router
	router := gin.New()
	router.GET("/health/ready", healthHandler.ReadinessProbe)

	// Test readiness probe
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/ready", nil)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "not_ready", response["status"])

	dependencies := response["dependencies"].(map[string]interface{})
	assert.Equal(t, "unhealthy", dependencies["database"])
	assert.Equal(t, "healthy", dependencies["redis"])

	// Verify all expectations met
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestReadinessProbe_RedisUnhealthy(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Mock DB (healthy)
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectPing()

	// Mock Redis with error
	redisClient, redisMock := redismock.NewClientMock()
	redisMock.ExpectPing().SetErr(redis.Nil)

	// Create handler
	healthHandler := handler.NewHealthHandler(db, redisClient)

	// Create test router
	router := gin.New()
	router.GET("/health/ready", healthHandler.ReadinessProbe)

	// Test readiness probe
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/ready", nil)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "not_ready", response["status"])

	dependencies := response["dependencies"].(map[string]interface{})
	assert.Equal(t, "healthy", dependencies["database"])
	assert.Equal(t, "unhealthy", dependencies["redis"])

	// Verify all expectations met
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestHealthCheck_Comprehensive(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Mock DB
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectPing()

	// Mock Redis
	redisClient, redisMock := redismock.NewClientMock()
	redisMock.ExpectPing().SetVal("PONG")

	// Create handler
	healthHandler := handler.NewHealthHandler(db, redisClient)

	// Create test router
	router := gin.New()
	router.GET("/health", healthHandler.HealthCheck)

	// Test health check
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.NotEmpty(t, response["timestamp"])
	assert.NotEmpty(t, response["version"])

	dependencies := response["dependencies"].(map[string]interface{})
	assert.Equal(t, "healthy", dependencies["database"])
	assert.Equal(t, "healthy", dependencies["redis"])

	// Check details are present
	details := response["details"].(map[string]interface{})
	assert.NotNil(t, details["database"])
	assert.NotNil(t, details["redis"])

	// Verify all expectations met
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestHealthCheck_Timeout(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create a real DB connection that will timeout
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer db.Close()

	// Make ping delay longer than health check timeout
	mock.ExpectPing().WillDelayFor(10 * time.Second)

	// Mock Redis
	redisClient, redisMock := redismock.NewClientMock()
	redisMock.ExpectPing().SetVal("PONG")

	// Create handler
	healthHandler := handler.NewHealthHandler(db, redisClient)

	// Create test router
	router := gin.New()
	router.GET("/health/ready", healthHandler.ReadinessProbe)

	// Test readiness probe with timeout
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/ready", nil)

	// Use a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	router.ServeHTTP(w, req)

	// The response should indicate unhealthy due to timeout
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "not_ready", response["status"])
}

func BenchmarkLivenessProbe(b *testing.B) {
	gin.SetMode(gin.TestMode)

	db, _, _ := sqlmock.New()
	defer db.Close()

	redisClient, _ := redismock.NewClientMock()

	healthHandler := handler.NewHealthHandler(db, redisClient)

	router := gin.New()
	router.GET("/health/live", healthHandler.LivenessProbe)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health/live", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkReadinessProbe(b *testing.B) {
	gin.SetMode(gin.TestMode)

	db, mock, _ := sqlmock.New(sqlmock.MonitorPingsOption(true))
	defer db.Close()

	redisClient, redisMock := redismock.NewClientMock()

	healthHandler := handler.NewHealthHandler(db, redisClient)

	router := gin.New()
	router.GET("/health/ready", healthHandler.ReadinessProbe)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectPing()
		redisMock.ExpectPing().SetVal("PONG")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health/ready", nil)
		router.ServeHTTP(w, req)
	}
}
