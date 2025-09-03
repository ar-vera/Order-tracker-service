package main

import (
	"Order-tracker-service/internal/domain"
	"Order-tracker-service/internal/repository"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"os"
)

func main() {
	if err := godotenv.Load("/Users/veraryabova/Desktop/Go/Order-tracker-service/.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	pgDsn := os.Getenv("PG_DSN")

	fmt.Println(pgDsn)

	dataBase, err := sqlx.Connect("postgres", pgDsn)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer dataBase.Close()

	// db.RunMigrations(dataBase, "/Users/veraryabova/Desktop/Go/Order-tracker-service/migrations/")

	repo := repository.NewOrderRepository(dataBase)

	// ⚡ Тестовые данные
	order := domain.GetTestOrder()

	// --- Create ---
	//err = repo.Create(order)
	if err != nil {
		log.Fatal("Ошибка при сохранении заказа:", err)
	}
	fmt.Println("✅ Order успешно сохранён в БД")

	// --- GetById ---
	orderFromDb, err := repo.GetById(order.OrderUID)
	if err != nil {
		log.Fatal("Ошибка при получении заказа по ID:", err)
	}
	fmt.Printf("📦 Order по UID=%s:\n%+v\n\n", order.OrderUID, orderFromDb)

	// --- GetAll ---
	allOrders, err := repo.GetAll()
	if err != nil {
		log.Fatal("Ошибка при получении всех заказов:", err)
	}
	fmt.Println("📜 Все заказы из БД:")
	for _, o := range allOrders {
		fmt.Printf("- %s (%s)\n", o.OrderUID, o.CustomerID)
	}
}
