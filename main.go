package main

import (
	"fmt"
	"os"
	"strconv"

	dotenv "github.com/joho/godotenv"
)

func main() {
	fmt.Println("Program started")

	err := dotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	// queries := []string{"French Revolution wikipedia"}
	query := os.Args[2:]
	limit, err := strconv.Atoi(os.Args[1])
	if err != nil {
		return
	}
	fullSearch(query, limit)

}
