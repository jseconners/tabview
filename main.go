
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"log"


	"database/sql"
	_ "github.com/go-sql-driver/mysql"

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
	DataSources []datasource
	Labels map[string]bool
}

type datasource struct {
	DB *sql.DB
	Label string
	Tables []string
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
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	return db
}

func processDataSources() {
	conLen := len(Conf.Databases)

	ConnPool.DataSources = make([]datasource, conLen)
	ConnPool.Labels = make(map[string]bool, conLen)

	for i, dbConf := range Conf.Databases {
		ConnPool.DataSources[i].Label = dbConf.Label
		ConnPool.Labels[dbConf.Label] = true

		connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbConf.User, dbConf.Pass, dbConf.Host, dbConf.Port, dbConf.Name)
		ConnPool.DataSources[i].DB = dbHandle(connString)
	}
}


func dbList(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	
	labels := []string{}
	for label, _ := range ConnPool.Labels {
    	labels = append(labels, label)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(labels)
}

/**
func tableList(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Tables!")
}

func dataList(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Data!")
}
**/

func main() {

	processConfig()
	processDataSources()



	// router
	router := mux.NewRouter().StrictSlash(true)
	
	// routes
	router.HandleFunc("/", dbList).Methods("GET")


	// start server
	log.Fatal(http.ListenAndServe(":8080", router))
	/**
	**/
}