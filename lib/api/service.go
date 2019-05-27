package api

import (
	"github.com/SENERGY-Platform/device-repository/lib/config"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"net/http"
)

func init() {
	endpoints = append(endpoints, ServiceEndpoints)
}

func ServiceEndpoints(config config.Config, control Controller, router *jwt_http_router.Router) {

	resource := "/services"

	router.GET(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})
}
