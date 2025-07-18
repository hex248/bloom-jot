package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Jot struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Datetime    time.Time `json:"dateTime"`
	CreatedAt   time.Time `json:"createdAt"`
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	jots, err := getAllJots()
	if err != nil {
		http.Error(w, "Error fetching Jots: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "All Jots:\n")
	for _, j := range jots {
		fmt.Fprintf(w, "Name: %s\nDescription: %s\nDate: %s\nCreated: %s\n\n", j.Name, j.Description, j.Datetime.Format(time.RFC822), j.CreatedAt.Format(time.RFC822))
	}
}

func allHandler(w http.ResponseWriter, r *http.Request) {
	jots, err := getAllJots()
	if err != nil {
		http.Error(w, "Error fetching Jots: "+err.Error(), http.StatusInternalServerError)
		return
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

	dateTime := r.URL.Query().Get("dateTime")

	if dateTime == "" {
		dateTime = "01 Jan 70 00:00 +0000"
	}

	// replace second space with '+' if there are two consecutive spaces
	// this is a workaround for Apple Shortcuts' URL formatting for '+'
	if len(dateTime) > 0 {
		for i := range len(dateTime) - 1 {
			if dateTime[i] == ' ' && dateTime[i+1] == ' ' {
				dateTime = dateTime[:i+1] + "+" + dateTime[i+2:]
				break
			}
		}
	}

	parsedDateTime, err := time.Parse(time.RFC822Z, dateTime)
	if err != nil {
		http.Error(w, "Invalid dateTime format: "+err.Error(), http.StatusBadRequest)
		return
	}
	if parsedDateTime.IsZero() {
		http.Error(w, "Invalid dateTime value", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", "./jot.db")
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	err = createJot(db, name, description, parsedDateTime)
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

		log.Println("Database file created: jot.db")
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
		description TEXT NOT NULL,
		dateTime DATETIME NOT NULL,
		createdAt DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	statement, err := db.Prepare(createJotTableSQL)
	if err != nil {
		log.Fatal(err.Error())
	}

	statement.Exec()
}

func getAllJots() ([]Jot, error) {
	db, err := sql.Open("sqlite3", "./jot.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM jots")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jots []Jot
	for rows.Next() {
		var j Jot
		if err := rows.Scan(&j.Id, &j.Name, &j.Description, &j.Datetime, &j.CreatedAt); err != nil {
			return nil, err
		}
		jots = append(jots, j)
	}

	return jots, nil
}

func createJot(db *sql.DB, name string, description string, dateTime time.Time) error {
	insertSQL := `INSERT INTO jots (name, description, dateTime) VALUES (?, ?, ?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec(name, description, dateTime)
	return err
}
