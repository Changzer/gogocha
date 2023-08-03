package main

import (
	"database/sql"
	"encoding/json"
	"github.com/shopspring/decimal"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Customer struct {
	CustomerID int    `json:"customer_id"`
	Name       string `json:"name"`
	Cpf        string `json:"cpf"`
	Email      string `json:"email"`
}

type Order struct {
	OrderID      int             `json:"order_id"`
	CustomerID   int             `json:"customer_id"`
	CustomerName string          `json:"customer_name"`
	OrderDate    string          `json:"order_date"`
	TotalAmount  decimal.Decimal `json:"total_amount"`
}

type OrderItem struct {
	OrderID      int             `json:"order_id"`
	ItemID       int             `json:"item_id"`
	ProductID    int             `json:"product_id"`
	Quantity     int             `json:"quantity"`
	PricePerItem decimal.Decimal `json:"price_per_item"`
}

func GetCustomersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT customer_id, name, cpf, email FROM customers")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		customers := []Customer{}
		for rows.Next() {
			var customer Customer
			if err := rows.Scan(&customer.CustomerID, &customer.Name, &customer.Cpf, &customer.Email); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			customers = append(customers, customer)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(customers)
	}
}

func CreateCustomerHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var customer Customer
		if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
			http.Error(w, "invalid request payload", http.StatusBadRequest)
			return
		}

		if customer.Name == "" {
			http.Error(w, "Name field is required", http.StatusBadRequest)
			return
		}

		result, err := db.Exec("INSERT INTO customers (name, cpf, email) VALUES ($1, $2, $3)",
			customer.Name, customer.Cpf, customer.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if rowsAffected != 1 {
			http.Error(w, "error creating customer", http.StatusInternalServerError)
			return
		}

		// Respond with a JSON object containing a success message
		response := map[string]string{
			"message": "Customer registered successfully!",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

func GetOrdersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT o.order_id, c.customer_id, c.name, o.order_date, o.total_amount FROM orders o JOIN customers c on o.customer_id = c.customer_id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		orders := []Order{}
		for rows.Next() {
			var order Order
			if err := rows.Scan(&order.OrderID, &order.CustomerID, &order.CustomerName, &order.OrderDate, &order.TotalAmount); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			orders = append(orders, order)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orders)
	}
}

func CreateOrderHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var orderData struct {
			Customer struct {
				Name  string `json:"name"`
				CPF   string `json:"cpf"`
				Email string `json:"email"`
			} `json:"customer"`

			Products []struct {
				ProductID int             `json:"product_id"`
				Quantity  int             `json:"quantity"`
				Price     decimal.Decimal `json:"price"`
			} `json:"products"`

			TotalAmount decimal.Decimal `json:"totalAmount"`
			OrderDate   time.Time       `json:"order_date"`
		}

		if err := json.NewDecoder(r.Body).Decode(&orderData); err != nil {
			http.Error(w, "invalid request payload", http.StatusBadRequest)
			return
		}

		if orderData.Customer.Name == "" {
			http.Error(w, "Name field is required", http.StatusBadRequest)
			return
		}

		if len(orderData.Products) == 0 {
			http.Error(w, "Please add at least one product to the order", http.StatusBadRequest)
			return
		}

		// Insert the customer details into the database
		var customerID int
		err := db.QueryRow("INSERT INTO customers (name, cpf, email) VALUES ($1, $2, $3) RETURNING customer_id",
			orderData.Customer.Name, orderData.Customer.CPF, orderData.Customer.Email).Scan(&customerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert the order into the database
		var orderID int
		err = db.QueryRow("INSERT INTO orders (customer_id, total_amount, order_date) VALUES ($1, $2, $3) RETURNING order_id",
			customerID, orderData.TotalAmount, time.Now()).Scan(&orderID) // Use orderData.TotalAmount
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert the order items into the database
		for _, product := range orderData.Products {
			_, err = db.Exec("INSERT INTO order_items (order_id, product_id, quantity, price_per_item) VALUES ($1, $2, $3, $4)",
				orderID, product.ProductID, product.Quantity, product.Price)
			if err != nil {
				log.Printf("Error inserting product with ID %d: %v", product.ProductID, err) // Log the problematic product ID
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		response := map[string]int{
			"order_id": orderID,
		}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			log.Println("Error marshaling response:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonResponse)
	}
}

func GetOrderItemHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT oi.order_id, oi.item_id, p.product_id, oi.quantity, p.price FROM order_items oi JOIN products p on oi.product_id = p.product_id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		orderitems := []OrderItem{}
		for rows.Next() {
			var order_items OrderItem
			if err := rows.Scan(&order_items.OrderID, &order_items.ItemID, &order_items.ProductID, &order_items.Quantity, &order_items.PricePerItem); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			orderitems = append(orderitems, order_items)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orderitems)
	}
}

func CreateOrderItemHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var orderItems OrderItem
		if err := json.NewDecoder(r.Body).Decode(&orderItems); err != nil {
			http.Error(w, "invalid request payload", http.StatusBadRequest)
			return
		}
		if orderItems.OrderID == 0 || orderItems.ProductID == 0 || orderItems.Quantity == 0 {
			http.Error(w, "invalid or missing data in the request ", http.StatusBadRequest)
			return
		}

		product := getProductByID(db, orderItems.ItemID)
		if product == nil {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}

		pricePerItem := product.Price.Mul(decimal.NewFromInt(int64(orderItems.Quantity)))

		orderItems.PricePerItem = pricePerItem

		result, err := db.Exec("INSERT INTO order_items (order_id, product_id, quantity, price_per_item) VALUES ($1, $2, $3, $4)",
			orderItems.OrderID, orderItems.ProductID, orderItems.Quantity, orderItems.PricePerItem)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if rowsAffected != 1 {
			http.Error(w, "error creating order item list", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func GetProductByIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the product_id from the URL query parameter
		productIDStr := r.URL.Query().Get("product_id")
		productID, err := strconv.Atoi(productIDStr)
		if err != nil {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			return
		}

		// Fetch the product from the database
		product := getProductByID(db, productID)
		if product == nil {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}

		// Return the product as JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(product)
	}
}

func GetProductsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT product_id, product_name, price FROM products")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		products := []Product{}
		for rows.Next() {
			var product Product
			if err := rows.Scan(&product.ProductID, &product.ProductName, &product.Price); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			products = append(products, product)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
	}
}

func CreateProductsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var product Product
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			http.Error(w, "invalid request payload", http.StatusBadRequest)
			return
		}

		if product.ProductName == "" {
			http.Error(w, "Name field is required", http.StatusBadRequest)
			return
		}

		result, err := db.Exec("INSERT INTO products (product_name, price) VALUES ($1, $2)",
			product.ProductName, product.Price)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if rowsAffected != 1 {
			http.Error(w, "error creating product", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func AddToOrderItemHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var item OrderItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, "invalid request payload", http.StatusBadRequest)
			return
		}

		// Fetch the product price from the products table based on product_id
		productDetail := getProductByID(db, item.ProductID)
		if productDetail == nil {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}

		// Calculate the price per item
		item.PricePerItem = productDetail.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))

		// Insert the order item into the database
		_, err := db.Exec("INSERT INTO order_items (order_id, product_id, quantity, price_per_item) VALUES ($1, $2, $3, $4)",
			item.OrderID, item.ProductID, item.Quantity, item.PricePerItem.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update the total amount for the order in the orders table
		_, err = db.Exec("UPDATE orders SET total_amount = total_amount + $1 WHERE order_id = $2",
			item.PricePerItem.String(), item.OrderID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func CalculateTotalAmount(db *sql.DB, products []struct {
	ProductID int `json:"productID"`
	Quantity  int `json:"quantity"`
}) (decimal.Decimal, error) {
	// Calculate the total amount based on the items in the products list
	var totalAmount decimal.Decimal
	for _, product := range products {
		price, err := getProductPriceByID(db, product.ProductID)
		if err != nil {
			return decimal.Zero, err
		}
		totalAmount = totalAmount.Add(price.Mul(decimal.NewFromInt(int64(product.Quantity))))
	}
	return totalAmount, nil
}

func getProductPriceByID(db *sql.DB, productID int) (decimal.Decimal, error) {
	var price decimal.Decimal
	err := db.QueryRow("SELECT price FROM products WHERE product_id = $1", productID).Scan(&price)
	if err != nil {
		return decimal.Zero, err
	}
	return price, nil
}
