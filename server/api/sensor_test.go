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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/weaklayer/gateway/common/auth"
	"github.com/weaklayer/gateway/server/processing"
	"github.com/weaklayer/gateway/server/token"
)

func TestInstallAndEvents(t *testing.T) {

	// Construct the api
	tokenSecret := []byte("seeeeeeecret")
	pastSecrets := make([][]byte, 2)
	pastSecrets[0] = []byte("seeeeeeecret1")
	pastSecrets[1] = []byte("seeeeeeecret2")

	var tokenDuration int64 = 10000
	tokenProcessor := token.NewProcessor(tokenSecret, pastSecrets, tokenDuration)

	group, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("Failed to create test group UUID: %v", err)
	}

	key, err := auth.NewKey(group)
	if err != nil {
		t.Fatalf("Failed to create test Key: %v", err)
	}

	verifier, err := auth.NewVerifier(key)
	if err != nil {
		t.Fatalf("Failed to create test Verifier: %v", err)
	}

	installAPI, err := NewInstallAPI(tokenProcessor, []auth.Verifier{verifier})
	if err != nil {
		t.Fatalf("Falied to create install API endpoint: %v", err)
	}

	eventsProcessor := processing.EventProcessor{}

	eventsAPI, err := NewEventsAPI(tokenProcessor, eventsProcessor)
	if err != nil {
		t.Fatalf("Falied to create events API endpoint: %v", err)
	}

	sensorAPI := SensorAPI{
		EventsHandler:  eventsAPI,
		InstallHandler: installAPI,
	}

	// Now create and issue request to the api
	installRequest := InstallRequest{
		Key:   key,
		Label: "Test sensor!",
	}

	bodyBytes, err := json.Marshal(installRequest)
	if err != nil {
		t.Fatalf("Failed to create request body: %v", err)
	}

	request, err := http.NewRequest("POST", "/install", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	request.Header.Add("Content-type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(sensorAPI.ServeHTTP)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("Install request failed with status code %d", responseRecorder.Code)
	}

	if responseRecorder.HeaderMap.Get("Cache-Control") != "no-store" {
		t.Fatalf("Incorrect Cache-Control header value '%s'", responseRecorder.HeaderMap.Get("Cache-Control"))
	}
	if responseRecorder.HeaderMap.Get("Pragma") != "no-cache" {
		t.Fatalf("Incorrect Pragma header value '%s'", responseRecorder.HeaderMap.Get("Pragma"))
	}

	// First request succeeded. Now try renewing the token with the Authorization header set.

	var installResponse InstallResponse

	err = json.Unmarshal(responseRecorder.Body.Bytes(), &installResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshall install response: %v", err)
	}

	request, err = http.NewRequest("POST", "/install", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	request.Header.Add("Content-type", "application/json")
	request.Header.Add("Authorization", "Bearer "+installResponse.Token)
	responseRecorder = httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("Install request failed with status code %d", responseRecorder.Code)
	}

	err = json.Unmarshal(responseRecorder.Body.Bytes(), &installResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshall install response: %v", err)
	}

	// Now submit some events since you have the auth token.
	eventsBody := `[
		{"time":1, "type":"PageLoad", "protocol": "https", "hostname": "weaklayer.com", "port": 99},
		{"type":"PageLoad", "time":4, "protocol": "https", "hostname": "weaklayer.com", "port": 443},
		{"type":"PageLoad", "time":88, "protocol": "http", "hostname": "weaklayer.com", "port": 80}
	]`

	request, err = http.NewRequest("POST", "/events", strings.NewReader(eventsBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	request.Header.Add("Content-type", "application/json")
	request.Header.Add("Authorization", "Bearer "+installResponse.Token)
	responseRecorder = httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("Install request failed with status code %d", responseRecorder.Code)
	}
}
