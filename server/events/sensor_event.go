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

// sensorEventJSONSchema specifies conditions on the JSON representation of a SensorEvent
// Note: sensor and group are not marked as required. We fill these fields from the auth
// token the sensor presents so it does not need to provide them in the event
var sensorEventJSONSchemaString = `{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"title": "SensorEvent",
	"type": "object",
	"required": ["type", "time"],
	"properties": {
		"type": {
			"type": "string",
			"pattern": "^[a-zA-z]*$",
			"maxLength" : 255,
			"description": "Indicator for the type of event this is."
		},
		"time": {
			"type": "integer",
			"minimum": 0,
			"description": "When this event happened in microseconds since the UTC epoch."
		},
		"sensor": {
			"type": "string",
			"format": "uuid",
			"description": "The sensor this event concerns."
		},
		"group": {
			"type": "string",
			"format": "uuid",
			"description": "The group that this sensor belongs to."
		}
	}
}`

var sensorEventJSONSchema *gojsonschema.Schema

// EventType provides a restrction on the types of values that are allowed in the event "type" field
type EventType string

// Unknown is the event type used when we encounter an event
// type that the weaklayer gateway doesn't know about
const Unknown EventType = "Unknown"

// Event is an interface that the weaklayer events implement
type Event interface {
	GetType() EventType
	GetTime() int64
	GetSensor() uuid.UUID
	GetGroup() uuid.UUID
}

// SensorEvent gives all the fields common to sensor events
// It isn't a valid event on its own as there is no EventType for sensor event
// However, very useful for partial deserialiation and code deduplication
type SensorEvent struct {
	Type   EventType `json:"type"`
	Time   int64     `json:"time"`
	Sensor uuid.UUID `json:"sensor"`
	Group  uuid.UUID `json:"group"`
}

// GetType returns the event Type field
func (event SensorEvent) GetType() EventType {
	return event.Type
}

// GetTime returns the event Time field
func (event SensorEvent) GetTime() int64 {
	return event.Time
}

// GetSensor returns the event Sensor field
func (event SensorEvent) GetSensor() uuid.UUID {
	return event.Sensor
}

// GetGroup returns the event Group field
func (event SensorEvent) GetGroup() uuid.UUID {
	return event.Group
}

func parseSensorEvent(data json.RawMessage, sensor uuid.UUID, group uuid.UUID) (SensorEvent, error) {
	var sensorEvent SensorEvent

	err := schemaValidate(data, sensorEventJSONSchema)
	if err != nil {
		return sensorEvent, fmt.Errorf("Invalid sensor event: %w", err)
	}

	json.Unmarshal(data, &sensorEvent)
	if err != nil {
		return sensorEvent, fmt.Errorf("Could not parse request body JSON entry into sensor event: %w", err)
	}

	sensorEvent.Sensor = sensor
	sensorEvent.Group = group
	return sensorEvent, nil
}

func init() {
	schemaLoader := gojsonschema.NewStringLoader(sensorEventJSONSchemaString)
	s, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load JSON schema for sensor events")
	}

	sensorEventJSONSchema = s
}

type eventParseFucntion func(data json.RawMessage, sensor uuid.UUID, group uuid.UUID) (Event, error)

var eventParserMap map[EventType]eventParseFucntion = make(map[EventType]eventParseFucntion)

// ParseEvent validates json data and deserializes the json data into the proper event instance
func ParseEvent(data json.RawMessage, sensor uuid.UUID, group uuid.UUID) (Event, error) {

	sensorEvent, err := parseSensorEvent(data, sensor, group)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse as generic sensor event: %w", err)
	}

	parseFunction, ok := eventParserMap[sensorEvent.GetType()]
	if !ok {
		// This means that it is an unknown type
		// we will just use the sensor event and set the type to unknown
		sensorEvent.Type = Unknown
		return sensorEvent, nil
	}

	event, err := parseFunction(data, sensor, group)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse as %s: %w", sensorEvent.GetType(), err)
	}

	return event, nil
}

func schemaValidate(data json.RawMessage, schema *gojsonschema.Schema) error {
	eventLoader := gojsonschema.NewGoLoader(data)
	result, err := schema.Validate(eventLoader)
	if err != nil {
		return fmt.Errorf("Failed to validate event against JSON schema: %w", err)
	}
	if !result.Valid() {
		return fmt.Errorf("Event did not match JSON schema")
	}

	return nil
}
