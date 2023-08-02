package main

import (
	_ "context"
	"database/sql"
	_ "encoding/json"
	_ "fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	_ "os"
	_ "strconv"
	_ "strings"

	"github.com/gorilla/mux"
	_ "github.com/shopspring/decimal"
)

func main() {
	connStr := getConnectionString()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("error connecting to the database", err)
	}
	defer db.Close()

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
