package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"

	_ "service-mqtt/docs"
	"service-mqtt/internal/adapters/gorm"
	"service-mqtt/internal/adapters/mqtt"
	"service-mqtt/internal/adapters/nats"
	"service-mqtt/internal/config"
	"service-mqtt/internal/core/devices"
	api "service-mqtt/internal/delivery/http"
)

// @title           service-mqtt API
// @version         1.0
// @description     This is the API for the service-mqtt device management service.
// @BasePath        /
func main() {
	// 1. ---- Structured Logger ----
	logger := zerolog.
		New(os.Stdout).
		With().
		Timestamp().
		Str("service", "service-mqtt").
		Logger()

	// 2. ---- Load Runtime Config ----
	cfg := config.MustLoad()
	logger.Info().
		Str("mqtt_port", cfg.MQTTPort).
		Str("nats_url", cfg.NATSURL).
		Str("listen_addr", cfg.ListenAddr).
		Msg("configuration loaded")

	// 3. ---- Graceful Shutdown Context ----
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 4. ---- Database ----
	db, err := gorm.New(cfg.DatabaseDSN, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot connect to database")
	}

	deviceManager := devices.NewManager(db, logger)

	// 5. ---- NATS Publisher ----
	publisher, err := nats.New(cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot connect to NATS")
	}
	defer publisher.Close()
	logger.Info().Msg("NATS publisher initialized")

	// 6. ---- MQTT Server ----
	server, err := mqtt.NewServer(ctx, cfg, logger, publisher, deviceManager)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to setup MQTT server")
	}

	// 7. ---- HTTP API ----
	handler := api.New(deviceManager, logger)
	srv := &http.Server{Addr: cfg.ListenAddr, Handler: handler}

	// 8. ---- Start Server and Wait for Shutdown ----
	go func() {
		logger.Info().Msgf("MQTT server starting on port %s", cfg.MQTTPort)
		if err := server.Serve(); err != nil {
			logger.Fatal().Err(err).Msg("MQTT server failed")
		}
	}()

	go func() {
		logger.Info().Str("listen", cfg.ListenAddr).Msg("HTTP up")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("http")
		}
	}()

	<-ctx.Done()

	// 9. ---- Perform Graceful Shutdown ----
	logger.Info().Msg("shutdown signal received, closing connections")
	server.Close()
	_ = srv.Shutdown(context.Background())
	logger.Info().Msg("adapter shut down gracefully")
}
