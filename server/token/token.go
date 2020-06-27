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
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Processor is a class for constructing and verifying JWTs
type Processor struct {
	currentSecret []byte
	pastSecrets   [][]byte
	duration      int64
}

// Claims contains the information that is put into the sensor JWTs
type Claims struct {
	Group  uuid.UUID `json:"group"`
	Sensor uuid.UUID `json:"sensor"`
	jwt.StandardClaims
}

// NewProcessor creates a new instance capable of creating and verifying tokens
// duration is the number of seconds new tokens are valid for
func NewProcessor(currentSecret []byte, pastSecrets [][]byte, duration int64) Processor {
	return Processor{
		currentSecret: currentSecret,
		pastSecrets:   pastSecrets,
		duration:      duration,
	}
}

// NewToken creates a newly issued JWT for the given group/sensor combination
func (tokenProcessor Processor) NewToken(group uuid.UUID, sensor uuid.UUID) (string, int64, int64, error) {
	var retVal string

	issuedAt := time.Now().Unix()
	expiresAt := issuedAt + tokenProcessor.duration

	claims := Claims{
		group,
		sensor,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
			IssuedAt:  issuedAt,
			NotBefore: issuedAt,
		},
	}

	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := unsignedToken.SignedString(tokenProcessor.currentSecret)
	if err != nil {
		return retVal, expiresAt, issuedAt, fmt.Errorf("Failed to create signed token: %w", err)
	}

	retVal = tokenString

	return retVal, expiresAt, issuedAt, nil
}

// VerifyToken checks the token signature and time validity.
// Returns a boolean indicating if the token is valid or not.
func (tokenProcessor Processor) VerifyToken(tokenString string) (bool, Claims) {

	isValid, claims := tryTokenVerification(tokenString, tokenProcessor.currentSecret)
	if isValid {
		return true, claims
	}

	for _, secret := range tokenProcessor.pastSecrets {
		isValid, claims = tryTokenVerification(tokenString, secret)
		if isValid {
			return true, claims
		}
	}

	log.Info().Str("token", tokenString).Msg("Invalid token provided")
	return false, claims
}

func tryTokenVerification(tokenString string, secret []byte) (bool, Claims) {
	var claims Claims

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || method.Name != "HS256" {
			return nil, fmt.Errorf("Unexpected token signature algorithm '%s'", method.Name)
		}
		return secret, nil
	})

	if err != nil {
		log.Warn().Err(err).Msg("Token parsing failed")
		return false, claims
	}

	if !token.Valid {
		return false, claims
	}

	return true, claims
}
