package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type Jot struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./jot.db")
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, name, description FROM jots")
	if err != nil {
		http.Error(w, "Error fetching Jots: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var jots []Jot
	for rows.Next() {
		var j Jot
		if err := rows.Scan(&j.Id, &j.Name, &j.Description); err != nil {
			http.Error(w, "Error scanning Jot: "+err.Error(), http.StatusInternalServerError)
			return
		}
		jots = append(jots, j)
	}

	fmt.Fprintf(w, "All Jots:\n")
	for _, j := range jots {
		fmt.Fprintf(w, "Name: %s\nDescription: %s\n\n", j.Name, j.Description)
	}
}

func allHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./jot.db")
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, name, description FROM jots")
	if err != nil {
		http.Error(w, "Error fetching Jots: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var jots []Jot
	for rows.Next() {
		var j Jot
		if err := rows.Scan(&j.Id, &j.Name, &j.Description); err != nil {
			http.Error(w, "Error scanning Jot: "+err.Error(), http.StatusInternalServerError)
			return
		}
		jots = append(jots, j)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jots)

}

func newHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Name parameter is required", http.StatusBadRequest)
		return
	}
	description := r.URL.Query().Get("description")
	if description == "" {
		http.Error(w, "Description parameter is required", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", "./jot.db")
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	err = createJot(db, name, description)
	if err != nil {
		http.Error(w, "Error creating Jot: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Created Jot: %+v\n", name)
	fmt.Fprintf(w, "Created Jot: %s", name)
}

func init() {
	if _, err := os.Stat("jot.db"); os.IsNotExist(err) {
		file, err := os.Create("jot.db")
		if err != nil {
			log.Fatal(err.Error())
		}
		file.Close()
	}

	db, err := sql.Open("sqlite3", "./jot.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable(db)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/all", allHandler)
	http.HandleFunc("/new", newHandler)

	log.Println("jot is ready")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createTable(db *sql.DB) {
	createJotTableSQL := `
	CREATE TABLE IF NOT EXISTS jots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT NOT NULL
	);
	`

	statement, err := db.Prepare(createJotTableSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec()
}

func createJot(db *sql.DB, name string, description string) error {
	insertSQL := `INSERT INTO jots (name, description) VALUES (?, ?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec(name, description)
	return err
}
