package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config bundles all runtime configuration for the adapter.
type Config struct {
	// MQTT broker settings
	MQTTPort string

	// Device identity
	DeviceID string

	// NATS publisher settings
	NATSURL         string
	PublishTimeout  time.Duration
	EnableJetStream bool

	// Database
	DatabaseDSN string

	// API
	ListenAddr string
}

// MustLoad builds a Config from environment variables or panics if they are invalid.
func MustLoad() Config {
	cfg := Config{}

	// --- MQTT ---
	cfg.MQTTPort = getenv("MQTT_PORT", "1883")

	// --- NATS ---
	cfg.NATSURL = getenv("NATS_URL", "nats://localhost:4222")
	cfg.PublishTimeout = time.Duration(parseInt("PUBLISH_TIMEOUT_SEC", 5)) * time.Second
	cfg.EnableJetStream = parseBool("ENABLE_JETSTREAM", true)

	// --- Database ---
	cfg.DatabaseDSN = getenv("DATABASE_DSN", "postgres://user:password@localhost:5432/servicedb?sslmode=disable")

	// --- API ---
	cfg.ListenAddr = getenv("LISTEN_ADDR", ":9091")

	return cfg
}

// getenv returns the environment variable's value or a fallback.
func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

// parseBool parses an environment variable into a bool, falling back to def on error.
func parseBool(key string, def bool) bool {
	v, err := strconv.ParseBool(getenv(key, strconv.FormatBool(def)))
	if err != nil {
		return def
	}
	return v
}

// parseInt parses an environment variable into an int, falling back to def on error.
func parseInt(key string, def int) int {
	v, err := strconv.Atoi(getenv(key, fmt.Sprintf("%d", def)))
	if err != nil {
		return def
	}
	return v
}
