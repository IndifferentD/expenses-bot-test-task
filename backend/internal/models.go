package internal

type Category string

const (
	CategoryFood          Category = "еда"
	CategoryTransport     Category = "транспорт"
	CategoryHousing       Category = "жильё"
	CategoryEntertainment Category = "развлечения"
	CategoryHealth        Category = "здоровье"
	CategoryShopping      Category = "покупки"
	CategoryOther         Category = "другое"
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
	ID        int64    `db:"id" json:"id"`
	UserID    int64    `db:"user_id" json:"user_id"`
	Amount    int64    `db:"amount" json:"amount"`
	Category  Category `db:"category" json:"category"`
	SpentAt   string   `db:"spent_at" json:"spent_at"`
	Comment   *string  `db:"comment" json:"comment"`
	CreatedAt string   `db:"created_at" json:"created_at"`
	UpdatedAt string   `db:"updated_at" json:"updated_at"`
}
