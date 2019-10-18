
package main

import (
	"fmt"
	"net/http"
	"log" 

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)


func landing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Testing!")
}


func main() {
	viper.SetConfigName("conf")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", landing)
	log.Fatal(http.ListenAndServe(":8080", router))
}