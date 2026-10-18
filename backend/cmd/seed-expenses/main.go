package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"
	_ "time/tzdata"

	"expenses-bot/backend/internal/db"
	"expenses-bot/backend/internal/expense"
	"expenses-bot/backend/internal/user"

	"github.com/caarlos0/env/v11"
)

const expenseCount = 100

type config struct {
	DatabaseURL string `env:"DATABASE_URL,required"`
	Timezone    string `env:"TIMEZONE" envDefault:"Europe/Moscow"`
}

type sample struct {
	category  expense.Category
	comments  []string
	minAmount int64
	maxAmount int64
}

var samples = []sample{
	{expense.CategoryFood, []string{"продукты", "кофе", "овощи", "обед"}, 100, 2500},
	{expense.CategoryTransport, []string{"метро", "такси", "бензин", "автобус"}, 50, 4000},
	{expense.CategoryHousing, []string{"коммунальные услуги", "интернет", "бытовая химия"}, 300, 9000},
	{expense.CategoryEntertainment, []string{"кино", "книги", "концерт"}, 200, 5000},
	{expense.CategoryHealth, []string{"аптека", "врач", "витамины"}, 150, 7000},
	{expense.CategoryShopping, []string{"одежда", "канцелярия", "подарок"}, 200, 8000},
	{expense.CategoryOther, []string{"услуги", "подписка", "прочее"}, 100, 3000},
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatalf("failed to seed expenses: %v", err)
	}
}

func run(ctx context.Context) error {
	settings, err := env.ParseAs[config]()
	if err != nil {
		return fmt.Errorf("failed to load database configuration: %w", err)
	}
	database, err := db.ConnectPostgres(ctx, settings.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	users, err := user.NewRepository(database).List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}
	if len(users) == 0 {
		fmt.Println("Warning: 0 users in database; no expenses added.")
		return nil
	}

	location, err := time.LoadLocation(settings.Timezone)
	if err != nil {
		return fmt.Errorf("failed to load timezone %q: %w", settings.Timezone, err)
	}
	// Вычисляем «сегодня» в часовом поясе бота, затем работаем с календарными днями UTC.
	now := time.Now().In(location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	from := today.AddDate(0, -6, 0)
	days := int(today.Sub(from).Hours() / 24)
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	transaction, err := database.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin expense seed transaction: %w", err)
	}
	defer transaction.Rollback()
	for i := range expenseCount {
		template := samples[random.Intn(len(samples))]
		comment := template.comments[random.Intn(len(template.comments))]
		amount := template.minAmount + random.Int63n(template.maxAmount-template.minAmount+1)
		spentAt := expense.NewDate(from.AddDate(0, 0, random.Intn(days+1)))
		if _, err := transaction.ExecContext(ctx, `
			INSERT INTO expenses (user_id, amount, category, spent_at, comment)
			VALUES ($1, $2, $3, $4, $5)`,
			users[i%len(users)].ID, amount, template.category, spentAt, comment); err != nil {
			return fmt.Errorf("failed to create sample expense %d: %w", i+1, err)
		}
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("failed to commit sample expenses: %w", err)
	}
	fmt.Printf("Added %d sample expenses for %d users (%s through %s).\n", expenseCount, len(users), from.Format(time.DateOnly), today.Format(time.DateOnly))
	return nil
}
