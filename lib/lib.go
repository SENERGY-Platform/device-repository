package lib

import (
	"github.com/SENERGY-Platform/device-repository/lib/api"
	"github.com/SENERGY-Platform/device-repository/lib/com"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/device-repository/lib/source"
	"log"
)

func Start(conf config.Config) (stop func(), err error) {
	db, err := database.New(conf)
	if err != nil {
		log.Println("ERROR: unable to connect to database", err)
		return stop, err
	}

	perm, err := com.NewSecurity(conf)
	if err != nil {
		log.Println("ERROR: unable to create permission handler", err)
		return stop, err
	}

	ctrl, err := controller.New(conf, db, perm)
	if err != nil {
		db.Disconnect()
		log.Println("ERROR: unable to start control", err)
		return stop, err
	}

	sourceStop, err := source.Start(conf, ctrl)
	if err != nil {
		db.Disconnect()
		ctrl.Stop()
		log.Println("ERROR: unable to start source", err)
		return stop, err
	}

	err = api.Start(conf, ctrl)
	if err != nil {
		sourceStop()
		db.Disconnect()
		ctrl.Stop()
		log.Println("ERROR: unable to start api", err)
		return stop, err
	}

	return func() {
		sourceStop()
		db.Disconnect()
		ctrl.Stop()
	}, err
}
