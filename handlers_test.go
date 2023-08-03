package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/shopspring/decimal"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAddToOrderItemHandler(t *testing.T) {
	// Create a test database connection
	db := createTestDBConnection(t)
	defer db.Close()

	// Prepare a request with the order item data
	item := OrderItem{
		OrderID:   1, // Replace this with the order ID you want to add the item to
		ProductID: 1, // Replace this with the product ID you want to add
		Quantity:  3,
	}

	// Calculate the price per item
	productDetail := getProductByID(db, item.ProductID)
	if productDetail == nil {
		t.Fatal("Product not found")
	}

	item.PricePerItem = productDetail.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))

	// Convert the order item to JSON
	itemJSON, err := json.Marshal(item)
	if err != nil {
		t.Fatal("Failed to marshal order item:", err)
	}

	// Prepare the request
	req, err := http.NewRequest("POST", "/order_items", bytes.NewBuffer(itemJSON))
	if err != nil {
		t.Fatal("Failed to create request:", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Create the handler and serve the request
	handler := http.HandlerFunc(AddToOrderItemHandler(db))
	handler.ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected status code %d, but got %d", http.StatusCreated, rr.Code)
	}

	// Additional checks, if needed
	// ...
}

// getOrderByID retrieves an order from the database by its ID.
func getOrderByID(t *testing.T, db *sql.DB, orderID int) Order {
	var order Order
	row := db.QueryRow("SELECT * FROM orders WHERE order_id = $1", orderID)
	err := row.Scan(&order.OrderID, &order.CustomerID, &order.OrderDate, &order.TotalAmount)
	if err != nil {
		t.Fatalf("Failed to fetch order with ID %d: %v", orderID, err)
	}
	return order
}

func createTestOrder(t *testing.T, db *sql.DB) int {
	var orderID int
	err := db.QueryRow("INSERT INTO orders (customer_id, order_date, total_amount) VALUES ($1, $2, $3) RETURNING order_id",
		1, // Replace with the appropriate customer_id for the test
		time.Now().Format("2006-01-02 15:04:05"),
		decimal.NewFromFloat(0), // Start with a total_amount of 0 for the test order
	).Scan(&orderID)

	if err != nil {
		t.Fatalf("Failed to create test order: %v", err)
	}

	return orderID
}

func createTestDBConnection(t *testing.T) *sql.DB {
	connStr := getConnectionString()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("error connecting to the database: %v", err)
	}
	return db
}
