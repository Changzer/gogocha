package main

import (
	_ "context"
	"database/sql"
	_ "encoding/json"
	_ "fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
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

	r := mux.NewRouter()
	handler := cors.AllowAll().Handler(r)

	// Create database tables if they don't exist
	if err := createTables(db); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// Routes using mux
	r.HandleFunc("/customers", CreateCustomerHandler(db)).Methods("POST")
	r.HandleFunc("/customers", GetCustomersHandler(db)).Methods("GET") // Use a closure to pass db to the handler
	r.HandleFunc("/orders", CreateOrderHandler(db)).Methods("POST")
	r.HandleFunc("/orders", GetOrdersHandler(db)).Methods("GET")
	r.HandleFunc("/order_items", CreateOrderItemHandler(db)).Methods("POST")
	r.HandleFunc("/order_items", GetOrderItemHandler(db)).Methods("GET")
	r.HandleFunc("/products", CreateProductsHandler(db)).Methods("POST")
	r.HandleFunc("/products", GetProductsHandler(db)).Methods("GET")
	r.HandleFunc("/product", GetProductByIDHandler(db)).Methods("GET")

	// Start the server with the mux router
	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))

}
