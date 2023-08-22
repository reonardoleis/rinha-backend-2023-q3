package db

import (
	"database/sql/driver"
)

type CustomUUID string

func (c CustomUUID) Value() (driver.Value, error) {
	return string(c), nil
}

func (c *CustomUUID) Scan(src interface{}) error {
	uuid_uint8 := src.([]uint8)
	*c = CustomUUID(uuid_uint8)
	return nil
}
