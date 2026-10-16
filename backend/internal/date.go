package internal

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrInvalidDate = errors.New("invalid date")

// Date хранит календарный день без времени и часового пояса.
type Date struct {
	value time.Time
}

func NewDate(value time.Time) Date {
	year, month, day := value.Date()
	return Date{value: time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

func (d Date) IsZero() bool {
	return d.value.IsZero()
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("failed to decode date: %w", err)
	}
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return fmt.Errorf("%w: expected YYYY-MM-DD: %v", ErrInvalidDate, err)
	}
	*d = NewDate(parsed)
	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.value.Format(time.DateOnly))
}

func (d *Date) Scan(value any) error {
	parsed, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("failed to scan date: unsupported value %T", value)
	}
	*d = NewDate(parsed)
	return nil
}

func (d Date) Value() (driver.Value, error) {
	return d.value, nil
}
