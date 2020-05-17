package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
)

type Book struct {
	Id        string `json: "id"`
	Name      string `json: "name"`
	Author    string `json: "author"`
	Published string `json: "pub"`
}

var bks []Book
var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://philip:start123@localhost/bookservice?sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("You connected to your database.")
}

func main() {
	http.HandleFunc("/books", getBooks)

	fmt.Println("The Server is running.......")
	http.ListenAndServe(":8081", nil)

	// Routing
	// func handleRequest() {
	// 	// r := mux.NewRouter().StrictSlash(true)

	// 	http.HandleFunc("/books", getBooks)
	// 	// r.HandleFunc("api/v1/books").Methods("POST")
	// 	// r.HandleFunc("api/v1/books").Methods("PUT")
	// 	// r.HandleFunc("api/v1/books").Methods("DELETE")

	// 	fmt.Println("The Server is running.......")
	// 	http.ListenAndServe(":8081", nil)
	// }
}
func getBooks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT * FROM books;")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()

	for rows.Next() {
		bk := Book{}
		err := rows.Scan(&bk.Id, &bk.Name, &bk.Author, &bk.Published) // order matters
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		bks = append(bks, bk)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// for _, bk := range bks {
	// 	fmt.Fprintf(w, "%s, %s, %s, %s\n", bk.Id, bk.Name, bk.Author, bk.Published)
	// }

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(bks)
}
