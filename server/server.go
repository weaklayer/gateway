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

package server

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/weaklayer/gateway/common/auth"
	"github.com/weaklayer/gateway/server/api"
	"github.com/weaklayer/gateway/server/processing"
	"github.com/weaklayer/gateway/server/token"
)

// Config is the struct of values required to start the Weaklayer Gateway Server
type Config struct {
	Sensor struct {
		API struct {
			Host string
			Port int32
		}
		Token struct {
			Duration int64 // this is the number of microseconds new tokens are valid for
			Secrets  struct {
				Current []byte
				Past    [][]byte
			}
		}
		Install struct {
			Verifiers []auth.Verifier
		}
	}
}

// Run runs the Weaklayer Gateway Server
func Run(config Config) error {

	log.Info().Msg("Starting Weaklayer Gateway Server")

	//TODO: accept past secrets
	tokenProcessor := token.NewProcessor(config.Sensor.Token.Secrets.Current, config.Sensor.Token.Secrets.Past, config.Sensor.Token.Duration/1000000)
	installAPI, err := api.NewInstallAPI(tokenProcessor, config.Sensor.Install.Verifiers)
	if err != nil {
		return fmt.Errorf("Failed to create sensor install API endpoint: %w", err)
	}

	eventProcessor := processing.EventProcessor{}
	eventsAPI, err := api.NewEventsAPI(tokenProcessor, eventProcessor)
	if err != nil {
		return fmt.Errorf("Failed to create sensor events API endpoint: %w", err)
	}

	sensorAPI := api.SensorAPI{
		EventsHandler:  eventsAPI,
		InstallHandler: installAPI,
	}

	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.Sensor.API.Host, config.Sensor.API.Port),
		Handler: sensorAPI,
	}

	return s.ListenAndServe()
}
