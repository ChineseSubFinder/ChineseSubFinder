package models

import (
	"database/sql/driver"
	"encoding/json"
)

type StringList []string

func (p StringList) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *StringList) Scan(data interface{}) error {
	return json.Unmarshal(data.([]byte), &p)
}
