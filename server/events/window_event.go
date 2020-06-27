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
var windowEventJSONSchemaString = fmt.Sprintf(`{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"title": "WindowEvent",
	"type": "object",
	"allOf": [
		%s,
		{
			"properties": {
				"type": {
					"type": "string",
					"enum": ["%s"],
					"description": "Indicator for the type of event this is."
				}
			}
		}
	]
}`, sensorEventJSONSchemaString, Window)

// Window is the event type string for Window events
const Window EventType = "Window"

var windowEventJSONSchema *gojsonschema.Schema

// WindowEvent is a SensorEvent that indicates a web page was loaded on the sensor
type WindowEvent struct {
	SensorEvent
}

// GetType returns the event Type field
func (event WindowEvent) GetType() EventType {
	return Window
}

// GetTime returns the event Time field
func (event WindowEvent) GetTime() int64 {
	return event.Time
}

// GetSensor returns the event Sensor field
func (event WindowEvent) GetSensor() uuid.UUID {
	return event.Sensor
}

// GetGroup returns the event Group field
func (event WindowEvent) GetGroup() uuid.UUID {
	return event.Group
}

func parseWindowEvent(data json.RawMessage, sensor uuid.UUID, group uuid.UUID) (Event, error) {
	var windowEvent WindowEvent

	err := schemaValidate(data, windowEventJSONSchema)
	if err != nil {
		return windowEvent, fmt.Errorf("Invalid Window event: %w", err)
	}

	json.Unmarshal(data, &windowEvent)
	if err != nil {
		return windowEvent, fmt.Errorf("Could not parse request body JSON entry into Window event: %w", err)
	}

	windowEvent.Sensor = sensor
	windowEvent.Group = group
	windowEvent.Type = Window

	return windowEvent, nil
}

func init() {
	schemaLoader := gojsonschema.NewStringLoader(windowEventJSONSchemaString)
	s, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load JSON schema for Window event type")
	}

	windowEventJSONSchema = s

	eventParserMap[Window] = parseWindowEvent
}
