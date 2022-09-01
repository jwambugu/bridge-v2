package models

import "time"

type Branch struct {
	ID        uint64     `json:"id,omitempty" db:"id"`
	Name      string     `json:"name,omitempty" db:"name"`
	SapID     uint8      `json:"sap_id,omitempty" db:"sap_id"`
	CreatedAt *time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}
