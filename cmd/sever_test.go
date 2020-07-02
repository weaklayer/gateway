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

var serverC = &cobra.Command{
	Use:   "server",
	Short: "test",
}

func TestServerNoConfig(t *testing.T) {
	configFilePath = ""
	err := serverCmdRun(serverC, make([]string, 0))
	if err == nil {
		t.Fatalf("Secret command did not produce error")
	}
	t.Logf("%v", err)
}

func TestServerBadConfigPath(t *testing.T) {
	configFilePath = "test-configs/doesnt_exist.yaml"
	err := serverCmdRun(serverC, make([]string, 0))
	if err == nil {
		t.Fatalf("Secret command did not produce error")
	}
	t.Logf("%v", err)
}

func TestServerBadYamlConfig(t *testing.T) {
	configFilePath = "test-configs/invalid_yaml.yaml"
	err := serverCmdRun(serverC, make([]string, 0))
	if err == nil {
		t.Fatalf("Secret command did not produce error")
	}
	t.Logf("%v", err)
}

func TestServerMissingTokenSecretConfig(t *testing.T) {
	configFilePath = "test-configs/missing_current_secret.yaml"
	err := serverCmdRun(serverC, make([]string, 0))
	if err == nil {
		t.Fatalf("Secret command did not produce error")
	}
	t.Logf("%v", err)
}

func TestServerBadTokenSecretConfig(t *testing.T) {
	configFilePath = "test-configs/bad_current_secret.yaml"
	err := serverCmdRun(serverC, make([]string, 0))
	if err == nil {
		t.Fatalf("Secret command did not produce error")
	}
	t.Logf("%v", err)
}

func TestServerBadInstallKeyChecksumsConfig(t *testing.T) {
	configFilePath = "test-configs/bad_checksum.yaml"
	err := serverCmdRun(serverC, make([]string, 0))
	if err == nil {
		t.Fatalf("Secret command did not produce error")
	}
	t.Logf("%v", err)
}
