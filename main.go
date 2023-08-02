package main

import (
	_ "context"
	"database/sql"
	_ "encoding/json"
	_ "fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	_ "github.com/shopspring/decimal"
	"log"
	"net/http"
	_ "os"
	_ "strconv"
	_ "strings"
)

func main() {
	connStr := getConnectionString()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("error connecting to the database", err)
	}
	defer db.Close()

	runTest(db)

	r := mux.NewRouter()

	// Create database tables if they don't exist
	if err := createTables(db); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// Routes using mux
	r.HandleFunc("/customers", createCustomerHandler(db)).Methods("POST")
	r.HandleFunc("/customers", getCustomersHandler(db)).Methods("GET") // Use a closure to pass db to the handler
	r.HandleFunc("/orders", createOrderHandler(db)).Methods("POST")
	r.HandleFunc("/orders", getOrdersHandler(db)).Methods("GET")
	r.HandleFunc("/order_items", createOrderItemHandler(db)).Methods("POST")
	r.HandleFunc("/order_items", getOrderItemHandler(db)).Methods("GET")
	r.HandleFunc("/products", createProductsHandler(db)).Methods("POST")
	r.HandleFunc("/products", getProductsHandler(db)).Methods("GET")
	r.HandleFunc("/product", getProductByIDHandler(db)).Methods("GET")

	// Start the server with the mux router
	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func runTest(db *sql.DB) {
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

		// Insert the order item into the database
		result, err := db.Exec("INSERT INTO order_items (order_id, product_id, quantity, price_per_item) VALUES ($1, $2, $3, $4)",
			item.OrderID, item.ProductID, item.Quantity, item.PricePerItem.String())
		if err != nil {
			log.Fatal("Failed to add order item:", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Fatal("Failed to add order item:", err)
		}

		if rowsAffected != 1 {
			log.Fatal("Failed to add order item.")
		}
	}

	log.Println("Order items added successfully!")
}
