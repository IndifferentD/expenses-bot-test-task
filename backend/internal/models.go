package internal

import "time"

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

type User struct {
	ID             int64  `db:"id" json:"id"`
	TelegramUserID int64  `db:"telegram_user_id" json:"telegram_user_id"`
	Alias          string `db:"alias" json:"alias"`
}
