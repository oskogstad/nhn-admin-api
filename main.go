package main

import (
	"fmt"
	"log"
	"net/http"
)

func rootHandler(responseWriter http.ResponseWriter, request *http.Request) {
	fmt.Fprint(responseWriter, "Hello world")
}

func main() {
	http.HandleFunc("/", rootHandler)

	var portNumber = 8181

	fmt.Printf("Starting http listener on port %d ...\n", portNumber)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", portNumber), nil))
}
