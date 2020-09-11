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

package output

import (
	"testing"

	"github.com/google/uuid"
	"github.com/weaklayer/gateway/server/events"
)

type dummyConsumer struct {
	eventsReceived *int
}

func (dummyConsumer dummyConsumer) Consume(events []events.SensorEvent) error {
	*dummyConsumer.eventsReceived = *dummyConsumer.eventsReceived + len(events)
	return nil
}

func (dummyConsumer dummyConsumer) Close() {
}

func TestOutputConsume(t *testing.T) {

	sensor, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("Failed to generate UUID: %v", err)
	}
	group, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("Failed to generate UUID: %v", err)
	}

	event1 := events.SensorEvent{
		Type:   "Unknown",
		Time:   1,
		Sensor: sensor,
		Group:  group,
	}

	event2 := events.SensorEvent{
		Type:   "Unknown",
		Time:   1,
		Sensor: sensor,
		Group:  group,
	}

	events := []events.SensorEvent{event1, event2}

	er1 := 0
	er2 := 0
	output1 := dummyConsumer{eventsReceived: &er1}
	output2 := dummyConsumer{eventsReceived: &er2}

	topOutput := TopOutput{outputs: []Output{output1, output2}}

	err = topOutput.Consume(events)
	if err != nil {
		t.Fatalf("Error consuming event: %v", err)
	}

	if *output1.eventsReceived != 2 {
		t.Fatalf("Output 1 received %d events instead of the expected 2", *output1.eventsReceived)
	}
	if *output2.eventsReceived != 2 {
		t.Fatalf("Output 2 received %d events instead of the expected 2", *output2.eventsReceived)
	}
}
