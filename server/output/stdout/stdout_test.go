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

package stdout

import (
	"testing"

	"github.com/google/uuid"
	"github.com/weaklayer/gateway/server/events"
)

func TestStdoutOutputConsume(t *testing.T) {

	sensor, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("Failed to generate UUID: %v", err)
	}
	group, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("Failed to generate UUID: %v", err)
	}

	event := events.SensorEvent{
		Type:   events.Unknown,
		Time:   1,
		Sensor: sensor,
		Group:  group,
	}

	stdoutput := NewStdoutOutput()

	err = stdoutput.Consume([]events.Event{event})
	if err != nil {
		t.Fatalf("Error consuming event: %v", err)
	}
}
