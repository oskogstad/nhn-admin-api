package main

import (
	"net/http"
)

// RootHandler godoc
// @Summary Say hello
// @Description Say hello
// @Accept json
// @Produce json
// @Success 201
// @Failure 400
// @Router / [get]
func RootHandler(responseWriter http.ResponseWriter, r *http.Request) {
	responseWriter.Header().Set("Server", "A Go Web Server")
	responseWriter.Write([]byte("Hello, Go"))
}
