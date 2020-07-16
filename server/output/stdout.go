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
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/weaklayer/gateway/server/events"
)

// NewStdoutOutput creates an StdoutOutput instance
func NewStdoutOutput() StdoutOutput {
	stdoutput := StdoutOutput{
		eventStrings: make(chan string, 1000),
	}

	go stdoutput.process()

	return stdoutput
}

// StdoutOutput is an event output that writes events to stdout
type StdoutOutput struct {
	eventStrings chan string
}

// Consume takes the events and writes them to a channel for processing
func (stdoutOutput StdoutOutput) Consume(events []events.Event) error {
	var encounteredError = false

	for _, event := range events {
		serializedBytes, err := json.Marshal(event)
		if err != nil {
			encounteredError = true
			log.Warn().Err(err).Msg("Failed to serialized event. Discarding Event.")
			continue
		}

		stdoutOutput.eventStrings <- string(serializedBytes)
	}

	if encounteredError {
		return fmt.Errorf("Encountered errors serializing events for stdout")
	}

	return nil
}

func (stdoutOutput StdoutOutput) process() {
	for eventString := range stdoutOutput.eventStrings {
		n, err := fmt.Println(eventString)
		if err != nil {
			log.Warn().Err(err).Msg("Error printing event to stdout.")
		}

		if n < len(eventString) {
			log.Warn().Msg("Failed to print all event bytes to stdout.")
		}
	}
}
