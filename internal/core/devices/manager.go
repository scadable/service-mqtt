package devices

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"service-mqtt/pkg/rand"
)

type Manager struct {
	db *gorm.DB
	lg zerolog.Logger
}

func NewManager(db *gorm.DB, lg zerolog.Logger) *Manager {
	return &Manager{
		db: db,
		lg: lg.With().Str("component", "device-manager").Logger(),
	}
}

func (m *Manager) AddDevice(deviceType string) (*Device, error) {
	var devID string
	for {
		devID = rand.ID16()
		var count int64
		if err := m.db.Model(&Device{}).Where("id = ?", devID).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("failed to check for existing device ID: %w", err)
		}
		if count == 0 {
			break
		}
		m.lg.Warn().Str("device_id", devID).Msg("generated device ID already exists, retrying...")
	}

	dev := &Device{
		ID:           devID,
		DeviceType:   deviceType,
		MQTTUser:     devID,
		MQTTPassword: rand.Password(16),
		CreatedAt:    time.Now().UTC(),
	}

	if err := m.db.Create(dev).Error; err != nil {
		return nil, fmt.Errorf("create device record in db: %w", err)
	}

	return dev, nil
}

func (m *Manager) Authenticate(username, password string) bool {
	var dev Device
	if err := m.db.Where("mqtt_user = ? AND mqtt_password = ?", username, password).First(&dev).Error; err != nil {
		return false
	}
	return true
}
