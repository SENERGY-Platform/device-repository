package controller

import (
	"context"
	"errors"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"net/http"
	"time"
)

func (this *Controller) ReadService(id string, jwt jwt_http_router.Jwt) (result model.Service, err error, errCode int) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	deviceType, exists, err := this.db.GetDeviceTypeWithService(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, err, http.StatusNotFound
	}
	for _, service := range deviceType.Services {
		if service.Id == id {
			return service, nil, http.StatusOK
		}
	}
	return result, errors.New("found device-type without service in search for service"), http.StatusInternalServerError
}
