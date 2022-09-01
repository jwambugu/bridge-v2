package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type User struct {
	ID          uint64     `json:"id,omitempty" db:"id"`
	Name        string     `json:"name,omitempty" db:"name"`
	Email       string     `json:"email,omitempty" db:"email"`
	PhoneNumber string     `json:"phone_number,omitempty"`
	Password    string     `json:"password,omitempty" db:"password"`
	Meta        UserMeta   `json:"meta,omitempty" db:"meta"`
	CreatedAt   *time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

type KYCData struct {
	IDNumber string `json:"id_number,omitempty"`
	KRAPin   string `json:"kra_pin,omitempty"`
}

type UserMeta struct {
	KYCData KYCData `json:"kyc_data,omitempty"`
}

// Value implements driver.Valuer which simply returns the JSON-encoded representation of UserMeta.
func (m *UserMeta) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implement the sql.Scanner which decodes a JSON-encoded value into UserMeta.
func (m *UserMeta) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("user meta type assertion to []byte failed")
	}

	return json.Unmarshal(b, &m)
}
