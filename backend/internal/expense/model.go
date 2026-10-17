package expense

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Category string

const (
	CategoryFood          Category = "food"
	CategoryTransport     Category = "transport"
	CategoryHousing       Category = "housing"
	CategoryEntertainment Category = "entertainment"
	CategoryHealth        Category = "health"
	CategoryShopping      Category = "shopping"
	CategoryOther         Category = "other"
)

func (c Category) Valid() bool {
	switch c {
	case CategoryFood, CategoryTransport, CategoryHousing, CategoryEntertainment, CategoryHealth, CategoryShopping, CategoryOther:
		return true
	default:
		return false
	}
}

type Expense struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Amount    int64     `db:"amount" json:"amount"`
	Category  Category  `db:"category" json:"category"`
	SpentAt   Date      `db:"spent_at" json:"spent_at"`
	Comment   *string   `db:"comment" json:"comment"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ExpenseChanges: nil не меняет поле; пустой Comment очищает его.
type ExpenseChanges struct {
	Amount   *int64
	Category *Category
	SpentAt  *Date
	Comment  *string
}

type Summary struct {
	Count int64 `db:"count"`
	Total int64 `db:"total"`
}

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

func (d Date) Before(other Date) bool {
	return d.value.Before(other.value)
}

func (d Date) String() string {
	return d.value.Format(time.DateOnly)
}

func (d Date) Format(layout string) string {
	return d.value.Format(layout)
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
