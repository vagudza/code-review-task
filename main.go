package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/jackc/pgx/v4"
)

type bookModel struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Cost   int    `json:"cost"`
}

func getBooksByAuthor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	author := vars["author"]

	// Connection
	conn, err := pgx.Connect(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	// Query
	query := fmt.Sprintf("SELECT id, title, author, cost FROM books WHERE author = '%s'", author)
	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var books []bookModel
	// Loop through rows, using Scan to assign column data to struct fields
	for rows.Next() {
		var book bookModel
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Cost); err != nil {
			panic(err)
		}
		books = append(books, book)
	}

	fmt.Println(books)
	fmt.Println("Успешно выполнен запрос, заполнено записей: ", len(books))
	json.NewEncoder(w).Encode(books)
}

func main() {
	// definitions in config file:
	os.Setenv("DB_USERNAME", "postgres")
	os.Setenv("DB_PASSWORD", "root")
	os.Setenv("DB_HOST", "127.0.0.1")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "books")
	os.Setenv("DB_URL", fmt.Sprintf("postgres://%s:%s@%s:%s/%s", os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME")))

	fmt.Println("Запуск сервера...")
	r := mux.NewRouter()
	r.HandleFunc("/books/{author}", getBooksByAuthor).Methods("GET")
	http.ListenAndServe(":8080", r)
}
