package main

import "fmt"

func getConnectionString() string {
	host := "localhost"
	port := "5432"
	dbname := "gogocha"
	user := "postgres"
	password := "postgres"

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
}
