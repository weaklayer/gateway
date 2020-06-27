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
	"testing"

	"github.com/google/uuid"
	"github.com/weaklayer/gateway/common/auth"
)

func TestValidPageLoadEvent(t *testing.T) {
	const validWindowLocationEvent = `{
		"type": "WindowLocation",
		"time": 45678,
		"protocol": "https",
		"hostname": "weaklayer.com",
		"port": 443,
		"path": "",
		"search": "",
		"hash": "",
		"windowReference": 1
	}`

	event := testValidParseEvent(t, validWindowLocationEvent)
	if event.GetType() != WindowLocation {
		t.Fatalf("Parsed Page Load event as %s", event.GetType())
	}
}

func TestValidUnknownEvent(t *testing.T) {
	const validUnknownEvent = `{
		"type": "SuperAwesomeFutureEvent",
		"time": 45678452345,
		"cooooool": {
			"asdfas": "adfgsdfgsdfg"
		},
		"beans": [1,2,3,4,5,6,7,8],
		"yo": "yo"
	}`

	event := testValidParseEvent(t, validUnknownEvent)
	if event.GetType() != Unknown {
		t.Fatalf("Parsed Unknown event as %s", event.GetType())
	}
}

func testValidParseEvent(t *testing.T, data string) Event {
	sensor, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("Failed to generate UUID: %v", err)
	}
	group, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("Failed to generate UUID: %v", err)
	}

	event, err := ParseEvent([]byte(data), sensor, group)
	if err != nil {
		t.Fatalf("Failed to parse valid event: %v", err)
	}

	if !auth.UUIDEquals(sensor, event.GetSensor()) {
		t.Fatalf("Sensor UUIDs don't match")
	}

	if !auth.UUIDEquals(group, event.GetGroup()) {
		t.Fatalf("Group UUIDs don't match")
	}

	return event
}
