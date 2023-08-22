package db

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type CustomDate string

func (c CustomDate) Value() (driver.Value, error) {
	layout := "2006-01-02"
	return time.Parse(layout, string(c))
}

func (c *CustomDate) Scan(src interface{}) error {
	time, ok := src.(time.Time)
	if !ok {
		return fmt.Errorf("error scanning date")
	}

	timeStr := strings.Split(time.String(), " ")[0]

	*c = CustomDate(timeStr)

	return nil
}
