package user

type User struct {
	ID             int64  `db:"id" json:"id"`
	TelegramUserID int64  `db:"telegram_user_id" json:"telegram_user_id"`
	Alias          string `db:"alias" json:"alias"`
}
