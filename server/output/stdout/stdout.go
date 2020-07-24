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
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/weaklayer/gateway/server/events"
)

// NewStdoutOutput creates an StdoutOutput instance
func NewStdoutOutput() StdoutOutput {
	eventStrings := make(chan string, 10000)
	stdoutput := StdoutOutput{
		eventStrings: eventStrings,
	}

	go process(eventStrings)

	return stdoutput
}

// StdoutOutput is an event output that writes events to stdout
type StdoutOutput struct {
	eventStrings chan<- string
}

// Close does nothing for StdoutOutput
// It is implemented to conform with the Output interface
func (stdoutOutput StdoutOutput) Close() {
}

// Consume takes the events and writes them to a channel for processing
func (stdoutOutput StdoutOutput) Consume(events []events.Event) error {
	var encounteredError = false

	for _, event := range events {
		serializedBytes, err := json.Marshal(event)
		if err != nil {
			encounteredError = true
			log.Info().Err(err).Msg("Failed to serialized event. Discarding Event.")
			continue
		}

		select {
		case stdoutOutput.eventStrings <- string(serializedBytes):
		default:
			encounteredError = true
			log.Info().Msgf("Event queue for stdout output full. Discarding Event.")
			continue
		}
	}

	if encounteredError {
		return fmt.Errorf("Encountered errors serializing events for stdout")
	}

	return nil
}

func process(eventStrings <-chan string) {
	for eventString := range eventStrings {
		n, err := fmt.Println(eventString)
		if err != nil {
			log.Info().Err(err).Msg("Error printing event to stdout.")
		}

		if n < len(eventString) {
			log.Info().Msg("Failed to print all event bytes to stdout.")
		}
	}
}
