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
	http.HandleFunc("/books/update", updateBookForm)
	http.HandleFunc("/books/update/process", booksUpdateProcess)
	http.HandleFunc("/books/delete/process", booksDeleteProcess)

	fmt.Println("The Server is running.......")
	http.ListenAndServe(":8081", nil)
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

	row := db.QueryRow("SELECT * FROM books WHERE published = $1", published)

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

func updateBookForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
	}

	published := r.FormValue("published")
	if published == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	row := db.QueryRow("SELECT * FROM books WHERE isbn = $1", published)

	bk := Book{}
	err := row.Scan(&bk.Name, &bk.Author, &bk.Published)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	tpl.ExecuteTemplate(w, "update.gohtml", bk)
}

func booksUpdateProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	// get form values
	bk := Book{}
	bk.Name = r.FormValue("name")
	bk.Author = r.FormValue("author")
	bk.Published = r.FormValue("published")

	// validate form values
	if bk.Name == "" || bk.Author == "" || bk.Published == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	// insert values
	_, err := db.Exec("UPDATE books SET id = $1, name=$2, author=$3, published=$4 WHERE name=$1;", bk.Id, bk.Name, bk.Author, bk.Published)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	// confirm insertion
	tpl.ExecuteTemplate(w, "updated.gohtml", bk)
}

func booksDeleteProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	published := r.FormValue("published")
	if isbn == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	// delete book
	_, err := db.Exec("DELETE FROM books WHERE published=$1;", published)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/books", http.StatusSeeOther)
}
