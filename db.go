package main

import (
	"database/sql"
	"github.com/shopspring/decimal"
)

type Product struct {
	ProductID   int             `json:"product_id"`
	ProductName string          `json:"product_name"`
	Price       decimal.Decimal `json:"price"`
}

func createTables(db *sql.DB) error {
	// Create customer table
	_, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS customers (
    customer_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    cpf VARCHAR(255),
    email VARCHAR(255) )`)
	if err != nil {
		return err
	}

	// create orders table
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS orders (
    order_id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL,
    order_date TIMESTAMP NOT NULL,
    total_amount NUMERIC(10, 2) NOT NULL,
    FOREIGN KEY (customer_id) REFERENCES customers(customer_id) ) `)
	if err != nil {
		return err
	}

	// create products table

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS products(
	    product_id SERIAL PRIMARY KEY,
	    product_name VARCHAR(255) NOT NULL,
	    price NUMERIC(10, 2)
	)`)
	if err != nil {
		return err
	}

	// Create order_items table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS order_items (
	    item_id SERIAL PRIMARY KEY,
	    order_id INTEGER NOT NULL,
	    product_id INTEGER NOT NULL,
	    quantity INTEGER NOT NULL,
	    price_per_item NUMERIC (10, 2),
	    FOREIGN KEY (order_id) REFERENCES orders(order_id),
	    FOREIGN KEY (product_id) REFERENCES products(product_id)) `)
	if err != nil {
		return err
	}

	return nil
}

// Get a product by its ID
func getProductByID(db *sql.DB, productID int) *Product {
	var product Product
	row := db.QueryRow("SELECT product_id, product_name, price FROM products WHERE product_id = $1", productID)
	if err := row.Scan(&product.ProductID, &product.ProductName, &product.Price); err != nil {
		return nil
	}
	return &product
}
