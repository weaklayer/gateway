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

package token

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/weaklayer/gateway/common/auth"
)

func TestNewToken(t *testing.T) {
	pastSecrets := make([][]byte, 2)
	pastSecrets[0] = []byte("seeeeeeecret1")
	pastSecrets[1] = []byte("seeeeeeecret2")
	tokenProcessor := NewProcessor([]byte("hello"), pastSecrets, 24*60)

	group, err := uuid.NewRandom()
	if err != nil {
		t.Fatal("Failed to generate group identifier.", err)
	}

	sensor, err := uuid.NewRandom()
	if err != nil {
		t.Fatal("Failed to generate sensor identifier.", err)
	}

	tokenString, _, _, err := tokenProcessor.NewToken(group, sensor)
	if err != nil {
		t.Fatal("Failed to create token", err)
	}

	isTokenValid, claims := tokenProcessor.VerifyToken(tokenString)

	if !isTokenValid {
		t.Fatal("Token validation failed.", err)
	}

	if !auth.UUIDEquals(group, claims.Group) {
		t.Fatal("Token group identifier did not match expected value.")
	}

	if !auth.UUIDEquals(sensor, claims.Sensor) {
		t.Fatal("Token group identifier did not match expected value.")
	}
}

func TestPastTokenSecret(t *testing.T) {
	pastSecrets1 := make([][]byte, 0)
	tokenProcessor1 := NewProcessor([]byte("hello"), pastSecrets1, 24*60)

	pastSecrets2 := make([][]byte, 1)
	pastSecrets2[0] = []byte("hello")
	tokenProcessor2 := NewProcessor([]byte("hellosdfa"), pastSecrets2, 24*60)

	group, err := uuid.NewRandom()
	if err != nil {
		t.Fatal("Failed to generate group identifier.", err)
	}

	sensor, err := uuid.NewRandom()
	if err != nil {
		t.Fatal("Failed to generate sensor identifier.", err)
	}

	tokenString, _, _, err := tokenProcessor1.NewToken(group, sensor)
	if err != nil {
		t.Fatal("Failed to create token", err)
	}

	isTokenValid, claims := tokenProcessor2.VerifyToken(tokenString)

	if !isTokenValid {
		t.Fatal("Token validation failed.", err)
	}

	if !auth.UUIDEquals(group, claims.Group) {
		t.Fatal("Token group identifier did not match expected value.")
	}

	if !auth.UUIDEquals(sensor, claims.Sensor) {
		t.Fatal("Token group identifier did not match expected value.")
	}
}

func TestSigningAlgNone(t *testing.T) {
	pastSecrets := make([][]byte, 2)
	pastSecrets[0] = []byte("seeeeeeecret1")
	pastSecrets[1] = []byte("seeeeeeecret2")
	tokenProcessor := NewProcessor([]byte("hello"), pastSecrets, 24*60)

	group, err := uuid.NewRandom()
	if err != nil {
		t.Fatal("Failed to generate group identifier.", err)
	}
	sensor, err := uuid.NewRandom()
	if err != nil {
		t.Fatal("Failed to generate sensor identifier.", err)
	}
	issuedAt := time.Now().Unix()
	expiresAt := issuedAt + 10000

	claims := Claims{
		group,
		sensor,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
			IssuedAt:  issuedAt,
			NotBefore: issuedAt,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)

	tokenString, err := token.SigningString()
	if err != nil {
		t.Fatal("Failed to create test token.", err)
	}

	isTokenValid, _ := tokenProcessor.VerifyToken(tokenString)
	if isTokenValid {
		t.Fatal("Token with None signing method should not be seen as valid.")
	}
}
