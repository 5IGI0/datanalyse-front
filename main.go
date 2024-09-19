package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

var GlobalContext = struct {
	Database     *sqlx.DB
	DatabaseName string
	Tables       map[string]TableDescription
	Finders      map[string]Finder
}{
	Tables: make(map[string]TableDescription),
	Finders: map[string]Finder{
		"username":    &UsernameFinder{},
		"realnames":   &RealnameFinder{},
		"phone":       &PhoneFinder{},
		"email":       &EmailFinder{},
		"facebook_id": &ExactFinder{TargetAnalyzer: "fbid_analyzer"}}}

func main() {
	GlobalContext.Database = sqlx.MustConnect("mysql", os.Getenv("DATABASE_URI"))
	GlobalContext.Database.SetConnMaxLifetime(time.Minute * 3)
	GlobalContext.Database.SetMaxOpenConns(10)
	GlobalContext.Database.SetMaxIdleConns(10)

	GlobalContext.DatabaseName = strings.Split(os.Getenv("DATABASE_URI"), "/")[1]

	log.Println("loading datasets...")
	rows, err := GlobalContext.Database.Query("SHOW TABLES")
	AssertError(err)
	// the close is done below

	for rows.Next() {
		var table string
		AssertError(rows.Scan(&table))

		log.Println("loading", table)
		td, err := DescribeTable(table)
		if err != nil {
			log.Println("an error occured when loading", table, ":", err.Error())
			continue
		}
		GlobalContext.Tables[table] = td
	}
	log.Println("datasets loaded")
	rows.Close()

	log.Println("init finders...")
	for finder_name, finder := range GlobalContext.Finders {
		log.Println("init", finder_name)
		finder.Init()
	}
	log.Println("finder initialisations done.")

	r := mux.NewRouter()

	r.HandleFunc("/api/v1/status", ApiEndpoint(ApiStatus))
	r.HandleFunc("/api/v1/search/{finder}/{colkey}", ApiEndpoint(ApiSearch))

	panic(http.ListenAndServe(os.Getenv("LISTEN_ADDR"), r))
}
