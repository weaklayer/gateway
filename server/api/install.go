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

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/weaklayer/gateway/common/auth"
	"github.com/weaklayer/gateway/server/events"
	"github.com/weaklayer/gateway/server/processing"
	"github.com/weaklayer/gateway/server/token"
	"github.com/xeipuuv/gojsonschema"
)

// InstallResponse forms the response body that the sensor will receive on a successful install request
type InstallResponse struct {
	Token     string    `json:"token"`
	Sensor    uuid.UUID `json:"sensor"`
	ExpiresAt int64     `json:"expiresAt"`
	IssuedAt  int64     `json:"issuedAt"`
}

// InstallRequest is what the sensor sends in the HTTP body to request installation
type InstallRequest struct {
	Key   auth.Key `json:"key"`
	Label string   `json:"label"`
}

var installRequestJSONSchema = fmt.Sprintf(`
{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"title": "InstallRequest",
	"type": "object",
	"properties": {
		"label": {
			"type": "string"
		},
		"key": %s
	}
}
`, auth.KeyJSONSchema)

// InstallAPI handles requests to the /install path for sensor installation
type InstallAPI struct {
	tokenProcessor       token.Processor
	installRequestSchema *gojsonschema.Schema
	eventProcessor       processing.EventProcessor
	verifiers            []auth.Verifier
}

// NewInstallAPI provisions a sensor API with its required resources
func NewInstallAPI(tokenProcessor token.Processor, verifiers []auth.Verifier) (InstallAPI, error) {
	var installAPI InstallAPI

	schemaLoader := gojsonschema.NewStringLoader(installRequestJSONSchema)
	schemaVerifier, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return installAPI, fmt.Errorf("Failed to load install request JSON schema: %w", err)
	}

	return InstallAPI{
		tokenProcessor:       tokenProcessor,
		installRequestSchema: schemaVerifier,
		verifiers:            verifiers,
	}, nil
}

func (installAPI InstallAPI) parseInstallRequest(data []byte) (InstallRequest, error) {
	var installRequest InstallRequest

	documentLoader := gojsonschema.NewBytesLoader(data)
	result, err := installAPI.installRequestSchema.Validate(documentLoader)
	if err != nil {
		return installRequest, fmt.Errorf("Failed to validate install request against json schema: %w", err)
	}

	if !result.Valid() {
		return installRequest, fmt.Errorf("Install request did not match json schema")
	}

	err = json.Unmarshal(data, &installRequest)
	if err != nil {
		return installRequest, fmt.Errorf("Failed to unmarshal install request: %w", err)
	}

	if !installAPI.isInstallRequestValid(installRequest) {
		return installRequest, fmt.Errorf("Install request verification unsuccessful")
	}

	return installRequest, nil
}

// Handle validates and processes install requests
func (installAPI InstallAPI) Handle(responseWriter http.ResponseWriter, request *http.Request) {

	// Don't want any responses cached
	responseWriter.Header().Add("Cache-Control", "no-store")
	responseWriter.Header().Add("Pragma", "no-cache")

	bodyContents, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to read request body contents")
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	installRequest, err := installAPI.parseInstallRequest(bodyContents)
	if err != nil {
		log.Info().Err(err).Msg("Failed to parse install request")
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}

	var tokenProvided = false
	var providedToken string
	authHeader := request.Header.Get("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			providedToken = strings.TrimPrefix(authHeader, "Bearer ")
			tokenProvided = true
		} else {
			log.Info().Msg("Provided Authorization header does not match Bearer schema")
			responseWriter.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	isInstallationRenewel := false
	group := installRequest.Key.Group
	var sensor uuid.UUID

	if tokenProvided {
		isTokenValid, claims := installAPI.tokenProcessor.VerifyToken(providedToken)
		if isTokenValid {
			if auth.UUIDEquals(installRequest.Key.Group, claims.Group) {
				sensor = claims.Sensor
				isInstallationRenewel = true
			} else {
				log.Info().Msgf("Token group %s differs from the install key group %s. Proceeding as new install.", claims.Group.String(), installRequest.Key.Group.String())
			}
		} else {
			log.Info().Msg("Received an invalid JWT for install renewel. Proceeding as new install.")
		}
	}

	if !isInstallationRenewel {
		sensor, err = uuid.NewRandom()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to generate new sensor identifier")
			responseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	token, expiresAt, issuedAt, err := installAPI.tokenProcessor.NewToken(group, sensor)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create new sensor token")
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	// The installation was successful
	// We need to generate an event for this before responding
	installEvent := events.InstallEvent{
		SensorEvent: events.SensorEvent{
			Type:   events.Install,
			Time:   time.Now().UnixNano() / 1000,
			Sensor: sensor,
			Group:  group,
		},
		Label: installRequest.Label,
	}
	installAPI.eventProcessor.Consume([]events.Event{installEvent})

	response := InstallResponse{
		Token:     token,
		Sensor:    sensor,
		ExpiresAt: expiresAt * 1000000, // expiresAt is in seconds. convert to micros to match weaklayer convention
		IssuedAt:  issuedAt * 1000000,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to marshal response body")
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	// responseWriter.Write sets Content-Length and the status to 200
	// Therefore don't try to se the status on error
	responseWriter.Header().Add("Content-Type", "application/json")
	bytesWritten, err := responseWriter.Write(responseBytes)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to write response body")
	}

	if bytesWritten != len(responseBytes) {
		log.Warn().Msgf("Failed to write entire response body. Wrote %d bytes out of %d total.", bytesWritten, len(responseBytes))
	}

	return
}

func (installAPI InstallAPI) isInstallRequestValid(installRequest InstallRequest) bool {

	// TODO: put the verifiers into a map keyed by groupid and then only go through the verifies for the given group

	for _, verifier := range installAPI.verifiers {
		if auth.UUIDEquals(verifier.Group, installRequest.Key.Group) {
			if auth.Verify(installRequest.Key, verifier) {
				return true
			}
		}
	}

	return false
}
