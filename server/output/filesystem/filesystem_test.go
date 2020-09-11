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

package filesystem

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/weaklayer/gateway/server/events"
)

func TestWritingEvents(t *testing.T) {
	group, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("Failed to generate UUID: %v", err)
	}
	sensor, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("Failed to generate UUID: %v", err)
	}

	filesystemOutput, err := NewFilesystemOutput(".", 60*time.Second, 100000000)

	event1 := events.SensorEvent{
		Type:   "Unknown",
		Time:   1,
		Sensor: sensor,
		Group:  group,
		Data:   make(map[string]interface{}),
	}

	event2 := events.SensorEvent{
		Type:   "Unknown",
		Time:   1,
		Sensor: sensor,
		Group:  group,
		Data:   make(map[string]interface{}),
	}

	err = filesystemOutput.Consume([]events.SensorEvent{event1, event2})
	if err != nil {
		t.Fatalf("Failed to write events to filesystem: %v", err)
	}

	filesystemOutput.Close()
	// wait for the file to close
	time.Sleep(1 * time.Second)

	dirPath := "./" + group.String()
	dir, err := os.Open(dirPath)
	if err != nil {
		t.Fatalf("Failed to open directory for reading: %v", err)
	}
	files, err := dir.Readdir(0)
	if err != nil {
		t.Fatalf("Failed to open directory for reading: %v", err)
	}

	for _, f := range files {
		fileName := f.Name()
		path := filepath.Join(dirPath, fileName)
		fileBytes, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		fileEvents := make([]events.SensorEvent, 0)
		json.Unmarshal(fileBytes, &fileEvents)
		if err != nil {
			t.Fatalf("Could not deserialize file contents into array of events: %v", err)
		}

		if len(fileEvents) != 2 {
			t.Fatalf("Wrong number of events found in file")
		}

		for _, event := range fileEvents {
			if group.String() != event.Group.String() || sensor.String() != event.Sensor.String() {
				t.Fatalf("Event identifiers do not match")
			}
		}
		os.Remove(path)
	}
	os.Remove(dirPath)
}
