package controller

import "github.com/SENERGY-Platform/iot-device-repository/lib/model"

func (this *Controller) updateHubOfDevice(oldDevice model.DeviceInstance, newDevice model.DeviceInstance) error {
	panic("implement me") //TODO
	/*

		if old.Gateway != "" && (old.Url != deviceInstance.Url || tagRemovedOrChanged(old.Tags, deviceInstance.Tags)) {
			//reset gateway hash
			gw, err := this.GetGateway(old.Gateway)
			if err != nil {
				return err
			}
			gw.Hash = ""
			err = eventsourcing.PublishGateway(gw, "")
			if err != nil {
				return err
			}
		}
		deviceInstance.Gateway = old.Gateway
		_, err = this.ordf.Update(old, deviceInstance)
		if err != nil {
			return
		}
	*/
}

func (this *Controller) resetHubOfDevice(device model.DeviceInstance) error {
	panic("implement me") //TODO
}
