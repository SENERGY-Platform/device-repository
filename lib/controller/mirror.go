/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"fmt"

	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/database"
)

func (this *Controller) MirrorUpdate() error {
	if !this.config.AsMgwMirror {
		return fmt.Errorf("not a mirror")
	}
	if this.mirrorPullCallback == nil {
		return fmt.Errorf("missing mirror pull callback")
	}
	this.mirrorPullCallback(this.config, this.db, true)
	return nil
}

func (this *Controller) SetMirrorPullCallback(pull func(config configuration.Config, db database.Database, checkLastUpdate bool)) {
	this.mirrorPullCallback = pull
}
