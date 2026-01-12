package handler

import "elibrary/internal/service"

type LocationHandler struct {
	Service *service.LocationService
}

func NewLocationHandler(service *service.LocationService) *LocationHandler {
	return &LocationHandler{Service: service}
}
