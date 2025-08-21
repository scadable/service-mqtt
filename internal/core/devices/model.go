package devices

import "time"

// Device represents a single device with MQTT credentials.
type Device struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	DeviceType   string    `json:"type"`
	MQTTUser     string    `json:"mqtt_user"`
	MQTTPassword string    `json:"mqtt_password"`
	CreatedAt    time.Time `json:"created_at"`
}
