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

package processing

import (
	"encoding/json"
	"fmt"

	"github.com/weaklayer/gateway/server/events"
)

// EventProcessor takes events and sends them to their next destinations
type EventProcessor struct{}

// Consume is the entry point for processing events.
// This process is asynchronous as there are multiple events and multiple destinations
func (eventProcessor EventProcessor) Consume(events []events.Event) error {
	for _, event := range events {
		serializedBytes, err := json.Marshal(event)
		if err != nil {
			return err
		}
		_, err = fmt.Println(string(serializedBytes))
		if err != nil {
			return err
		}
	}

	return nil
}

func (eventProcessor EventProcessor) consume(events []events.Event, errorChan chan error) {

	close(errorChan)
}
