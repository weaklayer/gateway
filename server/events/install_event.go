// SPDX-License-Identifier: AGPL-3.0-or-later

// Copyright (C) 2020 Mitchell Wasson

// This file is part of Weaklayer Gateway.

// Weaklayer Gateway is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package events

import (
	"github.com/google/uuid"
)

// Install is the event type string for Install events
const Install EventType = "Install"

// InstallEvent is a SensorEvent that indicates the sensor installation
type InstallEvent struct {
	SensorEvent
	Label string `json:"label"`
}

// GetType returns the event Type field
func (event InstallEvent) GetType() EventType {
	return Install
}

// GetTime returns the event Time field
func (event InstallEvent) GetTime() int64 {
	return event.Time
}

// GetSensor returns the event Sensor field
func (event InstallEvent) GetSensor() uuid.UUID {
	return event.Sensor
}

// GetGroup returns the event Group field
func (event InstallEvent) GetGroup() uuid.UUID {
	return event.Group
}
