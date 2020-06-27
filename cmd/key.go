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

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/weaklayer/gateway/common/auth"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var groupArg string

const defaultGroupArg = "<New Random UUID>"

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Generate an install key and corresponding verifier",
	RunE:  keyCmdRun,
}

func keyCmdRun(cmd *cobra.Command, args []string) error {

	var groupUUID uuid.UUID
	var err error
	if groupArg != defaultGroupArg {
		groupUUID, err = uuid.Parse(groupArg)
		if err != nil {
			return fmt.Errorf(`Could not parse "%s" as a UUID: %w`, groupArg, err)
		}
	} else {
		groupUUID, err = uuid.NewRandom()
		if err != nil {
			return fmt.Errorf("Failed to generate random group UUID: %w", err)
		}
	}

	key, err := auth.NewKey(groupUUID)
	if err != nil {
		return fmt.Errorf("Failed to generate install key: %w", err)
	}

	verifier, err := auth.NewVerifier(key)
	if err != nil {
		return fmt.Errorf("Failed to create verifier from install key: %w", err)
	}

	if !auth.Verify(key, verifier) {
		return fmt.Errorf("Generated key and verifier do not match")
	}

	keyJSONBytes, err := json.MarshalIndent(key, "  ", "  ")
	if err != nil {
		return fmt.Errorf("Failed to serialize generated key to JSON: %w", err)
	}

	verifierJSONBytes, err := json.MarshalIndent(verifier, "  ", "  ")
	if err != nil {
		return fmt.Errorf("Failed to serialize generated verifier to JSON: %w", err)
	}

	printString := fmt.Sprintf(`{
  "key": %s,
  "verifier": %s
}`, string(keyJSONBytes), string(verifierJSONBytes))

	printedBytes, err := fmt.Println(printString)
	if err != nil {
		return fmt.Errorf("Failed to print generated key and verifier: %w", err)
	}

	if printedBytes < len(printString) {
		return fmt.Errorf("Failed to print entire key and verifier")
	}

	return nil
}

func init() {
	keyCmd.Flags().StringVar(&groupArg, "group", defaultGroupArg, "Valid UUID to indentify the sensor group")
	rootCmd.AddCommand(keyCmd)
}
