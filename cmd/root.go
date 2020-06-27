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
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "weaklayer-gateway",
	Short: "Weaklayer is a software system for securing the web browser",
	Long: `
Welcome to Weaklayer Gateway
This program contains the Weaklayer Gateway Server and associated admin utilities
More information and documentation at https://weaklayer.com
`,
	RunE: rootCmdRun,
}

func rootCmdRun(cmd *cobra.Command, args []string) error {

	message := `
Use "weaklayer-gateway help" to display usage information
`

	printedBytes, err := fmt.Println(message)
	if err != nil {
		return fmt.Errorf("Failed to display message: %w", err)
	}

	if printedBytes < len(message) {
		return fmt.Errorf("Failed to display entire message")
	}

	return nil
}

// Execute is the main entry point for this program
func Execute() error {
	return rootCmd.Execute()
}
