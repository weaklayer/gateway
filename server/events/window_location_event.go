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
	"github.com/rs/zerolog/log"
	"github.com/xeipuuv/gojsonschema"
)

// The type field must be this event type
var windowLocaionEventJSONSchemaString = fmt.Sprintf(`{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"title": "WindowLocationEvent",
	"type": "object",
	"allOf": [
		%s,
		{
			"required": ["protocol", "hostname", "port", "path", "search", "hash", "windowReference"],
			"properties": {
				"type": {
					"type": "string",
					"enum": ["%s"],
					"description": "Indicator for the type of event this is."
				},
				"protocol": {
					"type": "string",
					"description": "The protocol scheme of the URL for the window"
				},
				"hostname": {
					"type": "string",
					"anyOf": [
						{"format":"idn-hostname"},
						{"format":"hostname"},
						{"format":"ipv4"},
						{"format":"ipv6"}
					],
					"description": "The hostname of the URL for the window. Can be an IP address."
				},
				"port": {
					"type": "integer",
					"minimum": 0,
					"maximum": 65535,
					"description": "The port of the URL for the window. A zero value indicates no port was set in the browser. The default port for the protocol can be assumed if it is a networked protocol. For example, port is not applicable to the file protocol"
				},
				"path": {
					"type": "string",
					"description": "The path section of the URL for the web page"
				},
				"search": {
					"type": "string",
					"description": "The search section of the URL for the web page"
				},
				"hash": {
					"type": "string",
					"description": "The hash section of the URL for the web page"
				},
				"windowReference": {
					"type": "integer",
					"minimum": 0,
					"description": "Refernce to the WindowEvent for the window this happened in"
				}
			}
		}
	]
}`, sensorEventJSONSchemaString, WindowLocation)

// WindowLocation is the event type string for PageLoad events
const WindowLocation EventType = "WindowLocation"

var windowLocationEventJSONSchema *gojsonschema.Schema

// WindowLocationEvent is a SensorEvent that indicates a web page was loaded on the sensor
type WindowLocationEvent struct {
	SensorEvent
	Protocol        string `json:"protocol"`
	Hostname        string `json:"hostname"`
	Port            int64  `json:"port"`
	Path            string `json:"path"`
	Search          string `json:"search"`
	Hash            string `json:"hash"`
	WindowReference int64  `json:"windowReference"`
}

// GetType returns the event Type field
func (event WindowLocationEvent) GetType() EventType {
	return WindowLocation
}

// GetTime returns the event Time field
func (event WindowLocationEvent) GetTime() int64 {
	return event.Time
}

// GetSensor returns the event Sensor field
func (event WindowLocationEvent) GetSensor() uuid.UUID {
	return event.Sensor
}

// GetGroup returns the event Group field
func (event WindowLocationEvent) GetGroup() uuid.UUID {
	return event.Group
}

func parseWindowLocationEvent(data json.RawMessage, sensor uuid.UUID, group uuid.UUID) (Event, error) {
	var windowLocationEvent WindowLocationEvent

	err := schemaValidate(data, windowLocationEventJSONSchema)
	if err != nil {
		return windowLocationEvent, fmt.Errorf("Invalid WindowLocation event: %w", err)
	}

	json.Unmarshal(data, &windowLocationEvent)
	if err != nil {
		return windowLocationEvent, fmt.Errorf("Could not parse request body JSON entry into WindowLocation event: %w", err)
	}

	windowLocationEvent.Sensor = sensor
	windowLocationEvent.Group = group
	windowLocationEvent.Type = WindowLocation

	return windowLocationEvent, nil
}

func init() {
	schemaLoader := gojsonschema.NewStringLoader(windowLocaionEventJSONSchemaString)
	s, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load JSON schema for WindowLocation event type")
	}

	windowLocationEventJSONSchema = s

	eventParserMap[WindowLocation] = parseWindowLocationEvent
}
