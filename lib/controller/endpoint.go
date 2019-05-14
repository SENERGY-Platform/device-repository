package controller

import "github.com/SENERGY-Platform/iot-device-repository/lib/model"

func (this *Controller) removeEndpointsOfDevice(device model.DeviceInstance) error {
	panic("todo") //TODO
}

func (this *Controller) updateEndpointsOfDevice(oldDevice, newDevice model.DeviceInstance) error {
	panic("todo") //TODO
	if oldDevice.Url != newDevice.Url {
		/*
			    deviceType, err := this.GetDeviceTypeById(device.DeviceType, 3)
				if err != nil {
					return err
				}

				//delete old
				endpoints, err := this.getEndpointsByDevice(device.Id)
				if err != nil {
					return err
				}
				for _, endpoint := range endpoints {
					this.ordf.Delete(endpoint)
					if err != nil {
						return err
					}
				}

				//create new
				for _, service := range deviceType.Services {
					endpoint := model.Endpoint{
						ProtocolHandler: service.Protocol.ProtocolHandlerUrl,
						Service:         service.Id,
						Device:          device.Id,
						Endpoint:        createEndpointString(service.EndpointFormat, device.Url, service.Url, device.Config),
					}
					if endpoint.Endpoint != "" {
						tempErr := this.ordf.SetIdDeep(&endpoint)
						if tempErr != nil {
							err = tempErr
						} else {
							_, tempErr = this.ordf.Insert(endpoint)
							if tempErr != nil {
								err = tempErr
							}
						}
					}
				}
				return err
		*/
	}
	return nil
}
