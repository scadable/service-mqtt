package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	httpSwagger "github.com/swaggo/http-swagger"
	"service-mqtt/internal/core/devices"
)

type Handler struct {
	mgr *devices.Manager
	lg  zerolog.Logger
}

type addDeviceRequest struct {
	Type string `json:"type"`
}

func New(m *devices.Manager, lg zerolog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	h := &Handler{mgr: m, lg: lg}

	r.Post("/devices", h.handleAdd)

	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusMovedPermanently)
	})
	r.Get("/docs/*", httpSwagger.WrapHandler)

	return r
}

// handleAdd creates a new device.
// @Summary      Add a new device
// @Description  Creates a new device with MQTT credentials.
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        device  body      addDeviceRequest     true  "Device Type"
// @Success      200     {object}  devices.Device
// @Failure      400     {string}  string "Bad Request"
// @Failure      500     {string}  string "Internal Server Error"
// @Router       /devices [post]
func (h *Handler) handleAdd(w http.ResponseWriter, r *http.Request) {
	var req addDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Type == "" {
		http.Error(w, `{"error": "body must be {\"type\":\"<deviceType>\"}"}`, http.StatusBadRequest)
		return
	}

	dev, err := h.mgr.AddDevice(req.Type)
	if err != nil {
		h.lg.Error().Err(err).Msg("add device")
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, dev)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
