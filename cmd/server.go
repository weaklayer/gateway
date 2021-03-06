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
			Host  string `json:"host,omitempty"`
			Port  int32  `json:"port,omitempty"`
			HTTPS struct {
				Certificate string `json:"certificate,omitempty"`
				Key         string `json:"key,omitempty"`
				Password    string `json:"password,omitempty"`
			} `json:"https,omitempty"`
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
	Outputs []struct {
		Type      string `json:"type,omitempty"`
		Directory string `json:"directory,omitempty"`
		Age       int64  `json:"age,omitempty"`
		Size      int    `json:"size,omitempty"`
	} `json:"outputs,omitempty"`
}

var configJSONSchema = fmt.Sprintf(`
{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"title": "Config",
	"type": "object",
	"required": ["sensor", "outputs"],
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
						},
						"https": {
							"type": "object",
							"description": "Config values for enabling https on the sensor api.",
							"properties": {
								"certificate": {
									"type": "string",
									"example": "/home/weaklayer/certificate.pem",
									"description": "Path to a TLS certificate (PEM format)"
								},
								"key": {
									"type": "string",
									"example": "/home/weaklayer/key.pem",
									"description": "Path to the TLS certificate private key (PEM format)"
								},
								"password": {
									"type": "string",
									"example": "examplekeypassword",
									"description": "Password for decrypting the private key (if applicable)."
								}
							}
						}
					}
				}
			}
		},
		"outputs": {
			"type": "array",
			"items": {
				"type": "object",
				"required": ["type"],
				"properties": {
					"type": {
						"type": "string",
						"enum": ["stdout", "filesystem"],
						"example": "stdout",
						"description": "The type of output to configure"
					},
					"directory": {
						"type": "string",
						"example": "./weaklayer-events",
						"description": "Directory that the gateway will write events to"
					},
					"age": {
						"type": "integer",
						"minimum": 1,
						"example": 3600,
						"description": "The file age, in microseconds, that the filesystem output will close files at"
					},
					"size": {
						"type": "integer",
						"minimum": 1,
						"example": 250000000,
						"description": "The file size, in bytes, that the filesystem output will close files at"
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
		return fmt.Errorf("Must specify a config file with --config")
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
		return fmt.Errorf("Failed to convert config into normalized format: %w", err)
	}

	err = validateConfigJSON(mergedConfigBytes)
	if err != nil {
		return fmt.Errorf("Config validation failed: %w", err)
	}

	var finalConfig server.Config
	err = json.Unmarshal(mergedConfigBytes, &finalConfig)
	if err != nil {
		return fmt.Errorf("Failed to convert config into normalized format: %w", err)
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

	// check that either both or neither of certificate and private key are specified for https
	if (config.Sensor.API.HTTPS.Certificate != "" && config.Sensor.API.HTTPS.Key == "") ||
		(config.Sensor.API.HTTPS.Certificate == "" && config.Sensor.API.HTTPS.Key != "") {
		return fmt.Errorf("Both a certificate and key must be specified to enable https")
	}

	if len(config.Outputs) == 0 {
		return fmt.Errorf("Must configure at least one output")
	}
	for i, configOutput := range config.Outputs {
		if configOutput.Type == "stdout" {
			// no fields for stdoutput
		} else if configOutput.Type == "filesystem" {
			directory := configOutput.Directory
			if directory == "" {
				return fmt.Errorf("Must specify a directory for filesystem output at location %d in outputs array", i)
			}
			age := configOutput.Age
			if age <= 0 {
				return fmt.Errorf("Must specify a strictly positive age for filesystem output at location %d in outputs array", i)
			}
			size := configOutput.Size
			if size <= 0 {
				return fmt.Errorf("Must specify a strictly positive size for filesystem output at location %d in outputs array", i)
			}
		} else {
			return fmt.Errorf("Unknown output type %s at at location %d in outputs array", configOutput.Type, i)
		}
	}

	return nil
}

func init() {

	serverCmd.Flags().StringVar(&configFilePath, "config", "", `Path to the desired config file
Permitted formats are YAML, JSON, TOML, HCL, envfile and Java properties config files`)

	rootCmd.AddCommand(serverCmd)
}
