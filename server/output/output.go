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
	"fmt"

	"github.com/weaklayer/gateway/server/events"
)

// Output is the interface that all outputs implement
// It allows the top-level output handler to keep a list of
// many outputs to send events to without knowing their implenetation
type Output interface {
	Consume(events []events.Event) error
	// Close performs any nessecary cleanup in an output (e.g. close file descriptor)
	Close()
}

// NewTopOutput creates a NewTopOutput instance
func NewTopOutput(outputs []Output) TopOutput {
	return TopOutput{
		outputs: outputs,
	}
}

// TopOutput is the main destination for events from the sensor API
// It dispatches events to all the different outputs
type TopOutput struct {
	outputs []Output
}

// Close closes all outputs
func (topOutput TopOutput) Close() {
	for _, output := range topOutput.outputs {
		output.Close()
	}
}

// Consume is the main destination for sensor events.
func (topOutput TopOutput) Consume(events []events.Event) error {
	var errors []error = nil
	for _, output := range topOutput.outputs {
		err := output.Consume(events)
		if err != nil {
			errors = append(errors, err)
		}
	}

	if errors != nil {
		return fmt.Errorf("Error(s) encountered consuming events: %v", errors)
	}

	return nil
}
