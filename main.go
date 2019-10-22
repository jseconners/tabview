
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"log"


	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	"github.com/joho/sqltocsv"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

type config struct { 
	Databases []db
}

type db struct { 
	Label string `json:"label"`
	Name  string `json:"name"`
	Port  int    `json:"port"`
	Host  string `json:"host"`
	User  string `json:"user"`
	Pass  string `json:"pass"`
}

type connectionPool struct {
	DataSources map[string]datasource
}

type datasource struct {
	DB *sql.DB
	Tables map[string]bool
}

var Conf config
var ConnPool connectionPool


func processConfig() {
	viper.SetConfigName("conf")
	viper.AddConfigPath(".")

	readErr := viper.ReadInConfig()
	if readErr != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", readErr))
	}

	decodeErr := viper.Unmarshal(&Conf)
	if decodeErr != nil {
		panic(fmt.Errorf("unable to decode into struct, %s", decodeErr))
	}	
}

func dbHandle(connString string) (*sql.DB) {
	db, err := sql.Open("mysql", connString)
	if err != nil {
		panic(err.Error()) 
	}
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	return db
}

func processDataSources() {
	conLen := len(Conf.Databases)

	ConnPool.DataSources = make(map[string]datasource, conLen)

	for _, dbConf := range Conf.Databases {
		
		// create database handle
		connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbConf.User, dbConf.Pass, dbConf.Host, dbConf.Port, dbConf.Name)
		db := dbHandle(connString)

		// populate available tables for database
		// just get them all for testing
		tables := make(map[string]bool)
		rows, err := db.Query("select table_name from information_schema.tables where table_type = 'BASE TABLE' and table_schema = database() order by table_name")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var table string
		for rows.Next() {
			rows.Scan(&table)
			tables[table] = true
		}

		// Add datasource to connection pool
		ConnPool.DataSources[dbConf.Label] = datasource {
			DB: db,
			Tables: tables,
		}
	}
}


func dbList(w http.ResponseWriter, r *http.Request) {
	labels := []string{}
	for label, _ := range ConnPool.DataSources {
    	labels = append(labels, label)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(labels)

}

func tableList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	
	cp, found := ConnPool.DataSources[vars["dbLabel"]]
	if !found {
		w.WriteHeader(404)
		return
	}

	tables := []string{}
	for tname, _ := range cp.Tables {
		tables = append(tables, tname)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tables)

}

func data(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	
	cp, found := ConnPool.DataSources[vars["dbLabel"]]
	if !found {
		w.WriteHeader(404)
		return
	}

	// Table must exist in the data source table list
	// generated independently of user input. vars["tableName"] is 
	// safe to use in query
	_, found = cp.Tables[vars["tableName"]]
	if !found {
		w.WriteHeader(404)
		return
	}

	rows, _ := cp.DB.Query(fmt.Sprintf("SELECT * FROM %s", vars["tableName"]))
	defer rows.Close()

	w.Header().Set("Content-type", "text/plain")
    // w.Header().Set("Content-Disposition", "attachment; filename=\"report.csv\"")
	sqltocsv.Write(w, rows)
}


func main() {

	processConfig()
	processDataSources()

	// router
	router := mux.NewRouter().StrictSlash(true)
	
	// routes
	router.HandleFunc("/", dbList).Methods("GET")
	router.HandleFunc("/{dbLabel}/", tableList).Methods("GET")
	router.HandleFunc("/{dbLabel}/{tableName}/", data).Methods("GET")

	// start server
	log.Fatal(http.ListenAndServe(":8080", router))
}