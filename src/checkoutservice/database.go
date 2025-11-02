package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"

    _ "github.com/lib/pq"
	pb "github.com/GoogleCloudPlatform/microservices-demo/src/checkoutservice/genproto"
)

var db *sql.DB

func initDB() error {
    var err error
    host := getEnv("POSTGRES_HOST", "postgres")
    port := getEnv("POSTGRES_PORT", "5432")
    user := getEnv("POSTGRES_USER", "boutique")
    password := getEnv("POSTGRES_PASSWORD", "boutique123")
    dbname := getEnv("POSTGRES_DB", "orders")

    connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname)

    for i := 0; i < 10; i++ {
        db, err = sql.Open("postgres", connStr)
        if err != nil {
            return err
        }

        err = db.Ping()
        if err == nil {
            break
        }
        log.Printf("Waiting for database... attempt %d/10", i+1)
        time.Sleep(2 * time.Second)
    }

    if err != nil {
        return fmt.Errorf("failed to connect to database: %v", err)
    }

    createTableQuery := `
    CREATE TABLE IF NOT EXISTS orders (
        id SERIAL PRIMARY KEY,
        order_id VARCHAR(255) NOT NULL,
        user_id VARCHAR(255) NOT NULL,
        user_currency VARCHAR(10) NOT NULL,
        items JSONB NOT NULL,
        shipping_address JSONB NOT NULL,
        shipping_cost JSONB NOT NULL,
        total_cost JSONB NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`

    _, err = db.Exec(createTableQuery)
    if err != nil {
        return fmt.Errorf("failed to create table: %v", err)
    }

    log.Println("Database initialized successfully")
    return nil
}

func saveOrder(orderID string, userID string, userCurrency string, items []*pb.CartItem, 
    address *pb.Address, shippingCost *pb.Money, totalCost *pb.Money) error {
    
    if db == nil {
        return fmt.Errorf("database not initialized")
    }

    itemsJSON, _ := json.Marshal(items)
    addressJSON, _ := json.Marshal(address)
    shippingCostJSON, _ := json.Marshal(shippingCost)
    totalCostJSON, _ := json.Marshal(totalCost)

    query := `INSERT INTO orders (order_id, user_id, user_currency, items, shipping_address, shipping_cost, total_cost) 
              VALUES ($1, $2, $3, $4, $5, $6, $7)`

    _, err := db.Exec(query, orderID, userID, userCurrency, itemsJSON, addressJSON, shippingCostJSON, totalCostJSON)
    if err != nil {
        log.Printf("Failed to save order: %v", err)
        return err
    }

    log.Printf("Order %s saved to database", orderID)
    return nil
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}