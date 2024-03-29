package models

import (
	"bridge/api/v1/pb"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type User struct {
	*pb.User
}

type UserMeta struct {
	*pb.UserMeta
}

// Value implements driver.Valuer which simply returns the JSON-encoded representation of UserMeta.
func (m *UserMeta) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implement the sql.Scanner which decodes a JSON-encoded value into UserMeta.
func (m *UserMeta) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed - got %T", b)
	}
	return json.Unmarshal(b, &m)
}
