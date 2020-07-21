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
	"context"
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/weaklayer/gateway/common/auth"
	"github.com/weaklayer/gateway/server/api"
	"github.com/weaklayer/gateway/server/output"
	"github.com/weaklayer/gateway/server/output/filesystem"
	"github.com/weaklayer/gateway/server/output/stdout"
	"github.com/weaklayer/gateway/server/token"
)

// Config is the struct of values required to start the Weaklayer Gateway Server
type Config struct {
	Sensor struct {
		API struct {
			Host  string
			Port  int32
			HTTPS struct {
				Certificate string
				Key         string
				Password    string
			}
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

func createEventOutput(config Config) (output.Output, error) {
	filesystemOutput, err := filesystem.NewFilesystemOutput(".")
	if err != nil {
		return output.NewTopOutput([]output.Output{}), err
	}
	outputs := []output.Output{stdout.NewStdoutOutput(), filesystemOutput}
	return output.NewTopOutput(outputs), nil
}

// Run runs the Weaklayer Gateway Server
func Run(config Config) error {

	log.Info().Msg("Starting Weaklayer Gateway Server")

	topLevelEventOutput, err := createEventOutput(config)
	if err != nil {
		return fmt.Errorf("Failed to create desired outputs: %w", err)
	}

	tokenProcessor := token.NewProcessor(config.Sensor.Token.Secrets.Current, config.Sensor.Token.Secrets.Past, config.Sensor.Token.Duration/1000000)
	installAPI, err := api.NewInstallAPI(tokenProcessor, topLevelEventOutput, config.Sensor.Install.Verifiers)
	if err != nil {
		return fmt.Errorf("Failed to create sensor install API endpoint: %w", err)
	}

	eventsAPI, err := api.NewEventsAPI(tokenProcessor, topLevelEventOutput)
	if err != nil {
		return fmt.Errorf("Failed to create sensor events API endpoint: %w", err)
	}

	sensorAPI := api.SensorAPI{
		EventsHandler:  eventsAPI,
		InstallHandler: installAPI,
	}

	var server *http.Server
	if useTLS(config.Sensor.API.HTTPS.Certificate, config.Sensor.API.HTTPS.Key) {
		tlsConfig, err := getTLSConfig(config.Sensor.API.HTTPS.Certificate, config.Sensor.API.HTTPS.Key, config.Sensor.API.HTTPS.Password)
		if err != nil {
			return fmt.Errorf("Failed to produce TLS config: %w", err)
		}

		server = &http.Server{
			ErrorLog:  stdlog.New(log.Logger, "", 0),
			Addr:      fmt.Sprintf("%s:%d", config.Sensor.API.Host, config.Sensor.API.Port),
			Handler:   sensorAPI,
			TLSConfig: tlsConfig,
		}

		go func() {
			err := server.ListenAndServeTLS("", "")
			if err != nil && err != http.ErrServerClosed {
				log.Error().Err(err).Msg("HTTP server error")
			}
		}()
	} else {
		server = &http.Server{
			ErrorLog: stdlog.New(log.Logger, "", 0),
			Addr:     fmt.Sprintf("%s:%d", config.Sensor.API.Host, config.Sensor.API.Port),
			Handler:  sensorAPI,
		}
		go func() {
			err := server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Error().Err(err).Msg("HTTP server error")
			}
		}()
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	// Stop the HTTP server. Give 5 seconds max for this.
	context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = server.Shutdown(context)
	if err != nil {
		log.Error().Err(err).Msg("Error shutting down HTTP server")
	}

	// Requests are stopped now.
	// Close outputs. Wait 1 seconds for it to happen.
	topLevelEventOutput.Close()
	time.Sleep(1 * time.Second)

	return nil
}
