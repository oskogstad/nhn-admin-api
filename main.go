package main

import (
	"fmt"
	"log"
	"net/http"

	_ "api.nhn.no/admin/docs" // docs is generated by Swag CLI, you have to import it.
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title NHN API Admin Service
// @description Various administrative functions for the API-gateway and k8s cluster
// @version 1.0
// @host localhost:8181
// @BasePath /admin/
func main() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/", RootHandler).Methods("GET")
	router.PathPrefix("/admin/docs").Handler(httpSwagger.WrapHandler)
	var portNumber = 8181
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", portNumber), router))
}
