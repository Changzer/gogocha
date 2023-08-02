package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"log"
	"net/http"
	"time"
)

// ...

func main() {
	// Database connection
	connStr := getConnectionString()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("error connecting to the database", err)
	}
	defer db.Close()

	// Create the customer
	customerID, err := createCustomer(db, "John Doe", "john@example.com", "123456789")
	if err != nil {
		log.Fatal("Failed to create customer:", err)
	}

	// Create an order for the customer
	orderID, err := createOrder(db, customerID, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Fatal("Failed to create order:", err)
	}

	// Select a few products to add to the order
	productsToAdd := []struct {
		ProductID int
		Quantity  int
	}{
		{ProductID: 1, Quantity: 2},
		{ProductID: 2, Quantity: 1},
		{ProductID: 3, Quantity: 3},
	}

	// Add products to the order
	for _, product := range productsToAdd {
		item := OrderItem{
			OrderID:   orderID,
			ProductID: product.ProductID,
			Quantity:  product.Quantity,
		}

		// Fetch the product price from the products table based on product_id
		productDetail := getProductByID(db, item.ProductID)
		if productDetail == nil {
			log.Fatal("Product not found")
		}

		// Calculate the price per item
		item.PricePerItem = productDetail.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))

		// Convert the order item to JSON
		itemJSON, err := json.Marshal(item)
		if err != nil {
			log.Fatal("Failed to marshal order item:", err)
		}

		// Send a POST request to add the order item
		resp, err := http.Post("http://localhost:8080/order_items", "application/json", bytes.NewBuffer(itemJSON))
		if err != nil {
			log.Fatal("Failed to add order item:", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			log.Fatal("Failed to add order item. Status:", resp.StatusCode)
		}
	}

	fmt.Println("Order items added successfully!")
}
