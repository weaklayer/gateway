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
	"strings"

	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/weaklayer/gateway/common/auth"
	"github.com/weaklayer/gateway/server"
	"github.com/xeipuuv/gojsonschema"
)

type stringsConfig struct {
	Sensor struct {
		API struct {
			Host string `json:"host,omitempty"`
			Port int32  `json:"port,omitempty"`
		} `json:"api,omitempty"`
		Token struct {
			Duration int64 `json:"duration,omitempty"`
			Secrets  struct {
				Current string   `json:"current,omitempty"`
				Past    []string `json:"past,omitempty"`
			} `json:"secrets,omitempty"`
		} `json:"token,omitempty"`
		Install struct {
			Verifiers []struct {
				Group    string `json:"group,omitempty"`
				Salt     string `json:"salt,omitempty"`
				Hash     string `json:"hash,omitempty"`
				Checksum string `json:"checksum,omitempty"`
			} `json:"verifiers,omitempty"`
		} `json:"install,omitempty"`
	} `json:"sensor,omitempty"`
}

var configJSONSchema = fmt.Sprintf(`
{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"title": "Config",
	"type": "object",
	"required": ["sensor"],
	"properties": {
		"sensor": {
			"type": "object",
			"required": ["token", "api"],
			"properties": {

				"token": {
					"type": "object",
					"required": ["secrets", "duration"],
					"properties": {
						"duration": {
							"type": "integer",
							"minimum": 1,
							"description": "The number of microseconds that issued tokens are valid for."
						},
						"secrets": {
							"type": "object",
							"required": ["current"],
							"properties": {
								"current": {
									"type": "string",
									"pattern": "^[a-zA-z0-9+\/]{86}==$",
									"example": "4QIuEulkopN+QbyMXOMVCKZBps0JjutFpI7U1OI6FgUFY6587Be9xLRPUSOK7fajM+RokZSwv3F31t5XAbTXnQ==",
									"description": "A 512-bit secret value that is used to sign and verify sensor auth tokens."
								},
								"past": {
									"type": "array",
									"items": {
										"type": "string",
										"pattern": "^[a-zA-z0-9+\/]{86}==$",
										"example": "4QIuEulkopN+QbyMXOMVCKZBps0JjutFpI7U1OI6FgUFY6587Be9xLRPUSOK7fajM+RokZSwv3F31t5XAbTXnQ==",
										"description": "A 512-bit secret value that is used to sign and verify sensor auth tokens."
									}
								}
							}
						}
					}
				},

				"install": {
					"type": "object",
					"properties": {
						"verifiers": {
							"type": "array",
							"items": %s
						}
					}
				},

				"api": {
					"type": "object",
					"required": ["host", "port"],
					"properties": {
						"host": {
							"type": "string",
							"format": "hostname",
							"example": "localhost",
							"description": "The host that the sensor API listens on."
						},
						"port": {
							"type": "integer",
							"minimum": 0,
							"maximum": 65535,
							"example": 8080,
							"description": "The port that the sensor API listens on."
						}
					}
				}
			}
		}
	}
}
`, auth.VerifierJSONSchema)

var configFilePath = ""

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the Weaklayer Gateway Server",
	RunE:  serverCmdRun,
}

func serverCmdRun(cmd *cobra.Command, args []string) error {

	// Configs where a default makes sense
	viper.SetDefault("sensor.api.host", "localhost")
	viper.SetDefault("sensor.api.port", 8080)
	viper.SetDefault("sensor.token.duration", 2419200000000)

	if configFilePath != "" {
		viper.SetConfigFile(configFilePath)
	} else {
		return fmt.Errorf("Must specify a config file")
	}

	viper.SetEnvPrefix("WEAKLAYER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("Failed to read config: %w", err)
	}

	var mergedConfig stringsConfig

	err = viper.Unmarshal(&mergedConfig)
	if err != nil {
		return fmt.Errorf("Failed to unmarshall config into struct: %w", err)
	}

	// We serialize the config from viper to json and back
	// This is primarily done because viper wont unmarshal a base64 string into a byte slice
	// json unmarshal does so we can marshal to json and back for the desired effect
	// we get the added bonus now of being able to use json schema validation
	mergedConfigBytes, err := json.Marshal(mergedConfig)
	if err != nil {
		return fmt.Errorf("Failed to convert config into desired format: %w", err)
	}

	err = validateConfigJSON(mergedConfigBytes)
	if err != nil {
		return fmt.Errorf("Config validation failed: %w", err)
	}

	var finalConfig server.Config
	err = json.Unmarshal(mergedConfigBytes, &finalConfig)
	if err != nil {
		return fmt.Errorf("Failed to convert config into desired format: %w", err)
	}

	err = validateConfigStruct(finalConfig)
	if err != nil {
		return fmt.Errorf("Config validation failed: %w", err)
	}

	return server.Run(finalConfig)
}

func validateConfigJSON(jsonBytes []byte) error {

	schemaLoader := gojsonschema.NewStringLoader(configJSONSchema)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return fmt.Errorf("Failed to load key JSON schema: %w", err)
	}

	documentLoader := gojsonschema.NewBytesLoader(jsonBytes)
	result, err := schema.Validate(documentLoader)
	if err != nil {
		return fmt.Errorf("Failed to validate against config JSON schema: %w", err)
	}

	if !result.Valid() {
		return fmt.Errorf("Config JSON schema validation returned not valid")
	}

	return nil
}

func validateConfigStruct(config server.Config) error {
	// The assumption here is that the json version of the config was already validated
	// This is for some extra stuff that can't be done with the json schema.
	// for example, verifier checksums

	for i, verifier := range config.Sensor.Install.Verifiers {
		if !auth.IsVerifierValid(verifier) {
			return fmt.Errorf("The checksum of install verfier at index %d did not match", i)
		}
	}

	return nil
}

func init() {

	serverCmd.Flags().StringVar(&configFilePath, "config", "", `Path to the desired config file
Permitted formats are YAML, JSON, TOML, HCL, envfile and Java properties config files`)

	rootCmd.AddCommand(serverCmd)
}
