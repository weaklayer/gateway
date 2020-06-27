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

package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/pbkdf2"
)

// KeyJSONSchema should be used to validate keys in JSON form
const KeyJSONSchema = `
{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"title": "Key",
	"type": "object",
	"properties": {
		"group": {
			"type": "string",
			"format": "uuid",
			"example": "73d0710f-c4a4-468a-9087-a06073bebe8c",
			"description": "The sensor group that this key enables installation into."
		},
		"secret": {
			"type": "string",
			"pattern": "^[a-zA-z0-9+\/]{86}==$",
			"example": "4QIuEulkopN+QbyMXOMVCKZBps0JjutFpI7U1OI6FgUFY6587Be9xLRPUSOK7fajM+RokZSwv3F31t5XAbTXnQ==",
			"description": "A 512-bit secret value that acts as a password for installing a sensor."
		},
		"checksum": {
			"type": "string",
			"pattern": "^[a-zA-z0-9+\/]{43}=$",
			"example": "iNg+ovQngoycxTRsb3byxjuE26E5eUQChVmmh54e2To=",
			"description": "A 256-bit hash of the other fields to determine if mistakes were made when entering a key."
		}
	}
}
`

// Key is a value that administrators will configure the Weaklayer Sensor with.
// Its purpose is to authenticate the sensor to the gateway so we know it
// is allowed to send data.
type Key struct {
	Group    uuid.UUID `json:"group"`
	Secret   []byte    `json:"secret"`
	Checksum []byte    `json:"checksum"`
}

// Verifier is data provided to the Weaklayer Gateway so it can verify the key
// that a Weaklayer Sensor Provides.
type Verifier struct {
	Group    uuid.UUID `json:"group"`
	Salt     []byte    `json:"salt"`
	Hash     []byte    `json:"hash"`
	Checksum []byte    `json:"checksum"`
}

// VerifierJSONSchema should be used to validate verifiers in JSON form
const VerifierJSONSchema = `
{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"title": "Verifier",
	"type": "object",
	"properties": {
		"group": {
			"type": "string",
			"format": "uuid",
			"example": "73d0710f-c4a4-468a-9087-a06073bebe8c",
			"description": "The sensor group that this verifier enables installation into."
		},
		"salt": {
			"type": "string",
			"pattern": "^[a-zA-z0-9+\/]{6}==$",
			"example": "SGQ7Fw==",
			"description": "A 32-bit random value used in the calculation of the install key secret hash."
		},
		"hash": {
			"type": "string",
			"pattern": "^[a-zA-z0-9+\/]{43}=$",
			"example": "BioOJy7op91gsmlO+L9PMAsRrUWNKP5JJQZ6JwCjlk4=",
			"description": "A 256-bit hash of the install key secret and the validator salt. Note this is a much more expensive hash is used here than for the checksum."
		},
		"checksum": {
			"type": "string",
			"pattern": "^[a-zA-z0-9+\/]{43}=$",
			"example": "jDmzy23i1Db7on0hTLZTsnATlGg79QOwwHZprlmSJ18=",
			"description": "A hash of the other fields to determine if mistakes were made when entering a validator."
		}
	}
}
`

// NewKey creates a new install key with a random secret for the provided group
// On error, the default Key is returned with error set
func NewKey(group uuid.UUID) (Key, error) {
	var key Key

	secret, err := NewRandomBytes(64)
	if err != nil {
		return key, fmt.Errorf("Install key secret generation failed: %w", err)
	}

	checksum, err := calculateKeyChecksum(group, secret)
	if err != nil {
		return key, fmt.Errorf("Install key checksum calculation failed: %w", err)
	}

	key = Key{
		Group:    group,
		Secret:   secret,
		Checksum: checksum,
	}

	return key, nil
}

// NewVerifier creates a new install verifier from an existing key and a random salt
// On error, the default Verifier is returned with error set
func NewVerifier(key Key) (Verifier, error) {
	var verifier Verifier

	// 32-bit salt according to NIST reccomendations
	salt, err := NewRandomBytes(4)
	if err != nil {
		return verifier, fmt.Errorf("Install verifier salt generation failed: %w", err)
	}

	hash := calculateHash(key, salt)

	sum, err := calculateVerifierChecksum(key.Group, salt, hash)
	if err != nil {
		return verifier, fmt.Errorf("Install verifier checksum calculation failed: %w", err)
	}

	verifier = Verifier{
		Group:    key.Group,
		Salt:     salt,
		Hash:     hash,
		Checksum: sum,
	}

	return verifier, nil
}

// Verify ensures the following:
//   - GroupIDs match
//   - Key Checksum valid
//	 - Verifier Checksum valid
//   - Key secret matches Verifier hash
//
//  Returns true if all these conditions are met and no errors. False otherwise.
func Verify(key Key, verifier Verifier) bool {
	return UUIDEquals(key.Group, verifier.Group) && isKeyChecksumValid(key) && IsVerifierValid(verifier) && isHashValid(key, verifier)
}

// IsVerifierValid ensures the verifier's checksum matches
func IsVerifierValid(verifier Verifier) bool {
	calculatedChecksum, err := calculateVerifierChecksum(verifier.Group, verifier.Salt, verifier.Hash)
	if err != nil {
		log.Info().Err(err).Msg("Install verifier checksum calculation failed")
		return false
	}

	return bytes.Equal(calculatedChecksum, verifier.Checksum)
}

// UUIDEquals returns true if the two UUID contents are equal. False otherwise.
func UUIDEquals(u1 uuid.UUID, u2 uuid.UUID) bool {
	u1Bytes, err := u1.MarshalBinary()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to marshal UUID to bytes")
		return false
	}

	u2Bytes, err := u2.MarshalBinary()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to marshal UUID to bytes")
		return false
	}

	return bytes.Equal(u1Bytes, u2Bytes)
}

// CalculateHash produces an install key hash given the salt and install key
func calculateHash(key Key, Salt []byte) []byte {
	// 10000 iterations according to NIST reccomendations
	return pbkdf2.Key(key.Secret, Salt, 10000, 32, sha256.New)
}

func isHashValid(key Key, verifier Verifier) bool {
	calculatedHash := calculateHash(key, verifier.Salt)
	return bytes.Equal(verifier.Hash, calculatedHash)
}

func calculateChecksum(input []byte) ([]byte, error) {
	digest := sha256.New()
	b, err := digest.Write(input)
	if err != nil || b != len(input) {
		return nil, fmt.Errorf("Error calculating sha256 digest: %w", err)
	}

	return digest.Sum(nil), nil
}

func calculateVerifierChecksum(group uuid.UUID, salt []byte, hash []byte) ([]byte, error) {
	uuidBytes, err := group.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal UUID to bytes: %w", err)
	}

	sumInput := append(uuidBytes, salt...)
	sumInput = append(sumInput, hash...)

	sum, err := calculateChecksum(sumInput)
	if err != nil {
		return nil, fmt.Errorf("Failed to calculate verifier checksum: %w", err)
	}

	return sum, nil
}

func calculateKeyChecksum(group uuid.UUID, secret []byte) ([]byte, error) {
	uuidBytes, err := group.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal UUID to bytes: %w", err)
	}

	sumInput := append(uuidBytes, secret...)

	sum, err := calculateChecksum(sumInput)
	if err != nil {
		return nil, fmt.Errorf("Failed to calculate key checksum: %w", err)
	}

	return sum, nil
}

func isKeyChecksumValid(key Key) bool {

	calculatedChecksum, err := calculateKeyChecksum(key.Group, key.Secret)
	if err != nil {
		log.Warn().Err(err).Msg("Key checksum validation failed")
		return false
	}

	return bytes.Equal(calculatedChecksum, key.Checksum)
}

// NewRandomBytes generates a byte array full of cryptographic strength random data
func NewRandomBytes(length int) ([]byte, error) {
	retVal := make([]byte, length)
	numBytes, err := rand.Read(retVal)

	if err != nil {
		return nil, fmt.Errorf("Random byte generation failed: %w", err)
	}
	if numBytes != length {
		return nil, fmt.Errorf("Failed to generate enough random bytes: %w", err)
	}

	return retVal, nil
}
