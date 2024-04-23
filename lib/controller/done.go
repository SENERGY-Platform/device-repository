package controller

import (
	"github.com/SENERGY-Platform/service-commons/pkg/donewait"
)

func (this *Controller) SendDone(done donewait.DoneMsg) error {
	return this.producer.SendDone(done)
}
