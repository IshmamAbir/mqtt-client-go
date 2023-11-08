package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SensorData struct {
	Id          uuid.UUID       `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Time        time.Time       `json:"time"`
	SensorId    int             `json:"sensor_id"`
	SensorValue json.RawMessage `json:"sensor_value" gorm:"type:jsonb"`
}

func (sd SensorData) TableName() string {
	return "sensor_data"
}

type Packet struct {
	Enc    bool        `json:"enc"`
	Buffer interface{} `json:"buffer"`
}
