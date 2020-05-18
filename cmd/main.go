package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
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

var tpl *template.Template

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

	tpl = template.Must(template.ParseGlob("../templates/*.gohtml"))
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/books", getBooks)
	http.HandleFunc("/books/show", getBook)
	http.HandleFunc("/books/create", booksCreateForm)
	http.HandleFunc("/books/create/process", createBookProcess)

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

func index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/books", http.StatusSeeOther)
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

	w.Header().Set("content-type", "text/html")
	json.NewEncoder(w).Encode(bks)

	tpl.ExecuteTemplate(w, "books.gohtml", bks)

}

// To get a specific book through it's Author
func getBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "Get" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	published := r.FormValue("published")
	if published == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
	}

	row := db.QueryRow("SELECT * FROM books WHERE authur = $1", published)

	bk := Book{}
	err := row.Scan(&bk.Id, &bk.Name, &bk.Author, &bk.Published)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	// json.NewEncoder(w).Encode(bk)
	w.Header().Set("content-type", "text/html; charset=utf8")
	tpl.ExecuteTemplate(w, "show.gohtml", bk)
}

func booksCreateForm(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "create.gohtml", nil)
}

func createBookProcess(w http.ResponseWriter, r *http.Request) {
	//check the method
	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
	}

	//Get the form value
	bk := Book{}
	//bk.Id = r.FormValue("id")
	bk.Name = r.FormValue("name")
	bk.Author = r.FormValue("author")
	bk.Published = r.FormValue("published")

	//validate the form values
	if bk.Name == "" || bk.Author == "" || bk.Published == "" {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	//insert the nform values
	_, err := db.Exec("INSERT INTO books (name, author, published) VALUES ($1, $2, $3)", bk.Name, bk.Author, bk.Published)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// // convert into JSON
	// w.Header().Set("content-type", "application/json")
	// json.NewEncoder(w).Encode(bk)
	// confirm execution
	tpl.ExecuteTemplate(w, "created.gohtml", bk)
}
