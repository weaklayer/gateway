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
	"testing"

	"github.com/spf13/cobra"
)

var sC = &cobra.Command{
	Use:   "secret",
	Short: "test",
}

func TestSecretRun(t *testing.T) {
	err := secretCmdRun(sC, make([]string, 0))
	if err != nil {
		t.Fatalf("Secret command execution failed: %v", err)
	}
}
