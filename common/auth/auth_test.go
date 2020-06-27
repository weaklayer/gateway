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

	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/xeipuuv/gojsonschema"
)

func TestKeySerialization(t *testing.T) {
	jsonKey := `{"group":"73d0710f-c4a4-468a-9087-a06073bebe8c","secret":"4QIuEulkopN+QbyMXOMVCKZBps0JjutFpI7U1OI6FgUFY6587Be9xLRPUSOK7fajM+RokZSwv3F31t5XAbTXnQ==","checksum":"iNg+ovQngoycxTRsb3byxjuE26E5eUQChVmmh54e2To="}`
	jsonKeyBytes := []byte(jsonKey)

	schemaLoader := gojsonschema.NewStringLoader(KeyJSONSchema)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		t.Fatal("Failed to load key JSON schema.", err)
	}

	documentLoader := gojsonschema.NewBytesLoader(jsonKeyBytes)
	result, err := schema.Validate(documentLoader)
	if err != nil {
		t.Fatal("Failed to validate against key JSON schema.", err)
	}

	if !result.Valid() {
		t.Fatal("Key JSON schema validation returned not valid.")
	}

	var key Key

	err = json.Unmarshal(jsonKeyBytes, &key)
	if err != nil {
		t.Fatal("Failed to deserialize key.", err)
	}

	serializedBytes, err := json.Marshal(key)
	if err != nil {
		t.Fatal("Failed to serialize key.", err)
	}

	if !bytes.Equal(jsonKeyBytes, serializedBytes) {
		t.Fatal("Processed key not equal to original.", jsonKey, string(serializedBytes))
	}
}

func TestVerifierSerialization(t *testing.T) {
	jsonVerifier := `{"group":"73d0710f-c4a4-468a-9087-a06073bebe8c","salt":"SGQ7Fw==","hash":"BioOJy7op91gsmlO+L9PMAsRrUWNKP5JJQZ6JwCjlk4=","checksum":"jDmzy23i1Db7on0hTLZTsnATlGg79QOwwHZprlmSJ18="}`
	jsonVerifierBytes := []byte(jsonVerifier)

	schemaLoader := gojsonschema.NewStringLoader(VerifierJSONSchema)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		t.Fatal("Failed to load verifier JSON schema.", err)
	}

	documentLoader := gojsonschema.NewBytesLoader(jsonVerifierBytes)
	result, err := schema.Validate(documentLoader)
	if err != nil {
		t.Fatal("Failed to validate against verifier JSON schema.", err)
	}

	if !result.Valid() {
		t.Fatal("Verifier JSON schema validation returned not valid.")
	}

	var verifier Verifier

	err = json.Unmarshal(jsonVerifierBytes, &verifier)
	if err != nil {
		t.Fatal("Failed to deserialize verifier.", err)
	}

	serializedBytes, err := json.Marshal(verifier)
	if err != nil {
		t.Fatal("Failed to serialize verifier.", err)
	}

	if !bytes.Equal(jsonVerifierBytes, serializedBytes) {
		t.Fatal("Processed verifier not equal to original.", jsonVerifier, string(serializedBytes))
	}
}

func TestKnownKeyVerification(t *testing.T) {
	jsonKey := `{"group":"73d0710f-c4a4-468a-9087-a06073bebe8c","secret":"4QIuEulkopN+QbyMXOMVCKZBps0JjutFpI7U1OI6FgUFY6587Be9xLRPUSOK7fajM+RokZSwv3F31t5XAbTXnQ==","checksum":"iNg+ovQngoycxTRsb3byxjuE26E5eUQChVmmh54e2To="}`
	jsonKeyBytes := []byte(jsonKey)

	var key Key

	err := json.Unmarshal(jsonKeyBytes, &key)
	if err != nil {
		t.Fatal("Failed to deserialize key.", err)
	}

	jsonVerifier := `{"group":"73d0710f-c4a4-468a-9087-a06073bebe8c","salt":"SGQ7Fw==","hash":"BioOJy7op91gsmlO+L9PMAsRrUWNKP5JJQZ6JwCjlk4=","checksum":"jDmzy23i1Db7on0hTLZTsnATlGg79QOwwHZprlmSJ18="}`
	jsonVerifierBytes := []byte(jsonVerifier)

	var verifier Verifier

	err = json.Unmarshal(jsonVerifierBytes, &verifier)
	if err != nil {
		t.Fatal("Failed to deserialize verifier.", err)
	}

	if !Verify(key, verifier) {
		t.Fatal("Install verification failed.")
	}
}

func TestNewKeyVerification(t *testing.T) {
	groupID := uuid.New()

	key, err := NewKey(groupID)
	if err != nil {
		t.Fatal("Failed to create new key.", err)
	}

	verifier, err := NewVerifier(key)
	if err != nil {
		t.Fatal("Failed to create new verifier.", err)
	}

	if !Verify(key, verifier) {
		t.Fatal("Install verification failed.")
	}
}
