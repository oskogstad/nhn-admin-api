package main

import (
	"encoding/json"
	"net/http"
	"regexp"
)

// Service ...
type Service struct {
	Name                string
	Port                int
	GatewayEndpoint     string
	ContainerRepository string
	ImageTag            string
}

// IsValidName checks if a string contains only asci letters , a-zA-Z
var IsValidName = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString

// NewAPIRegistration ...
// @Sumary Register new API
// @Accept json
// @Param data body main.Service true "Service info"
// @Produce json
// @success 201
// @Failure 400
// @Router /api/new [post]
func NewAPIRegistration(responseWriter http.ResponseWriter, request *http.Request) {
	var service Service

	err := json.NewDecoder(request.Body).Decode(&service)
	if err != nil {
		http.Error(responseWriter, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !IsValidName(service.GatewayEndpoint) || !IsValidName(service.Name) {
		http.Error(responseWriter, "Invalid name/endpoint, must match \"^[a-zA-Z]+$\"", http.StatusBadRequest)
		return
	}

	CreateNewServiceConfig(service)
	responseWriter.Write([]byte(service.Name + " created"))
}
