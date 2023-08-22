package db

import (
	"database/sql/driver"
	"strings"
)

type CustomStringSlice []string

func (s CustomStringSlice) Value() (driver.Value, error) {
	return strings.Join(s, ","), nil
}

func (s *CustomStringSlice) Scan(src interface{}) error {
	str := src.(string)
	if str == "" {
		*s = []string{}
		return nil
	}

	*s = strings.Split(string(src.(string)), ",")
	return nil
}
