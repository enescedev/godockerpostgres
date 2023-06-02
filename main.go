package main

// go mod init gohttpdb
// go get github.com/lib/pq

//http://localhost:9006/table

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Data struct {
	ID  int
	ID2 int
}

func main() {

	//.env dosyasını yükle
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// PostgreSQL bağlantı ayarları

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// PostgreSQL bağlantısını kontrol et
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("PostgreSQL bağlantısı başarılı")

	tmpl := template.Must(template.ParseFiles("template.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		rows, err := db.Query("SELECT * FROM test")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var data []Data
		for rows.Next() {
			var d Data
			err := rows.Scan(&d.ID, &d.ID2)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			data = append(data, d)
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r)
			return
		}

		id := r.FormValue("id")
		id2 := r.FormValue("id2")

		stmt, err := db.Prepare("INSERT INTO test (id, id2) VALUES ($1, $2)")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(id, id2)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	})

	fmt.Println("Web sunucusu başlatıldı. Formu görmek için <http://localhost:9006> adresini ziyaret edebilirsiniz.")
	log.Fatal(http.ListenAndServe(":9006", nil))
}
