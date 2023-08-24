package db

import (
	"database/sql/driver"
)

type CustomUUID string

func (c CustomUUID) Value() (driver.Value, error) {
	return string(c), nil
}

func (c *CustomUUID) Scan(src interface{}) error {
	uuid_str := src.(string)
	*c = CustomUUID(uuid_str)
	return nil
}
