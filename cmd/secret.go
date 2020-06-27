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

	"github.com/spf13/cobra"
	"github.com/weaklayer/gateway/common/auth"
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Generate a random 512-bit secret",
	RunE:  secretCmdRun,
}

type secretContainer struct {
	Secret []byte `json:"secret"`
}

func secretCmdRun(cmd *cobra.Command, args []string) error {

	secret, err := auth.NewRandomBytes(512 / 8)
	if err != nil {
		return fmt.Errorf("Failed to generate token signing secret: %w", err)
	}

	sc := secretContainer{
		Secret: secret,
	}

	secretContainerJSONBytes, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return fmt.Errorf("Failed to serialize generated secret to JSON: %w", err)
	}

	secretContainerJSON := string(secretContainerJSONBytes)

	printedBytes, err := fmt.Println(secretContainerJSON)
	if err != nil {
		return fmt.Errorf("Failed to print generated secret: %w", err)
	}

	if printedBytes < len(secretContainerJSON) {
		return fmt.Errorf("Failed to print entire secret")
	}

	return nil
}

func init() {
	rootCmd.AddCommand(secretCmd)
}
