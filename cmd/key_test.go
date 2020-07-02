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

var kC = &cobra.Command{
	Use:   "key",
	Short: "test",
}

func TestGoodUUID(t *testing.T) {
	groupArg = "0c805482-41b5-4ed3-a425-ff2694498293"
	err := keyCmdRun(kC, make([]string, 0))
	if err != nil {
		t.Fatalf("Key command execution failed with supplied UUID: %v", err)
	}
}

func TestBadUUID(t *testing.T) {
	groupArg = "0c805482-41b5-4ed3-a425-f94498293"
	err := keyCmdRun(kC, make([]string, 0))
	if err == nil {
		t.Fatalf("Key command execution succeed with bad UUID")
	}
}

func TestNoUUID(t *testing.T) {
	groupArg = defaultGroupArg
	err := keyCmdRun(kC, make([]string, 0))
	if err != nil {
		t.Fatalf("Key command execution failed with no supplied UUID: %v", err)
	}
}
