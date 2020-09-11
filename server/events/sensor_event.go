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
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// SensorEvent gives all the fields common to sensor events
// Other fields fall into a map
type SensorEvent struct {
	Type   string                 `json:"type"`
	Time   int64                  `json:"time"`
	Sensor uuid.UUID              `json:"sensor"`
	Group  uuid.UUID              `json:"group"`
	Data   map[string]interface{} `json:"-"`
}

// ParseEvent parses incoming events in JSON form and inject the sensor and group ids
func ParseEvent(data json.RawMessage, sensor uuid.UUID, group uuid.UUID) (SensorEvent, error) {
	var sensorEvent SensorEvent

	err := json.Unmarshal(data, &sensorEvent)
	if err != nil {
		return sensorEvent, fmt.Errorf("Could not parse request body JSON entry: %w", err)
	}

	if sensorEvent.Time <= 0 {
		return sensorEvent, fmt.Errorf("Invalid or unspecified time value in sensor event")
	}

	err = json.Unmarshal(data, &sensorEvent.Data)
	if err != nil {
		return sensorEvent, fmt.Errorf("Could not parse request body JSON entry: %w", err)
	}

	delete(sensorEvent.Data, "type")
	delete(sensorEvent.Data, "time")
	delete(sensorEvent.Data, "sensor")
	delete(sensorEvent.Data, "group")

	sensorEvent.Sensor = sensor
	sensorEvent.Group = group

	return sensorEvent, nil
}

// MarshalJSON produces the desired json serialization for sensor events
func (sensorEvent SensorEvent) MarshalJSON() ([]byte, error) {
	dataMap := sensorEvent.Data
	dataMap["type"] = sensorEvent.Type
	dataMap["time"] = sensorEvent.Time
	dataMap["sensor"] = sensorEvent.Sensor
	dataMap["group"] = sensorEvent.Group

	return json.Marshal(dataMap)
}
