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
	"crypto/tls"
	"testing"
)

func TestUseTLS(t *testing.T) {
	if useTLS("", "") {
		t.Fatal("Trying to use TLS when no certificate or key specified")
	}
}

func TestCertAndEncryptedKeyParsing(t *testing.T) {
	certPath := "test-crypt/example_cert.pem"
	keyPath := "test-crypt/example_key_encrypted.pem"
	keyPassword := "examplekeypassword"

	config, err := getTLSConfig(certPath, keyPath, keyPassword)
	if err != nil {
		t.Fatal("Failed to import TLS certificate and private key", err)
	}

	if config.MinVersion != tls.VersionTLS12 {
		t.Fatal("Minimum TLS version is not TLS1.2")
	}
}
func TestCertAndUnencryptedKeyParsing(t *testing.T) {
	certPath := "test-crypt/example_cert.pem"
	keyPath := "test-crypt/example_key_unencrypted.pem"
	keyPassword := ""

	config, err := getTLSConfig(certPath, keyPath, keyPassword)
	if err != nil {
		t.Fatal("Failed to import TLS certificate and private key", err)
	}

	if config.MinVersion != tls.VersionTLS12 {
		t.Fatal("Minimum TLS version is not TLS1.2")
	}
}

func TestMissingCert(t *testing.T) {
	certPath := "doesntexist.pem"
	keyPath := "test-crypt/example_key_unencrypted.pem"
	keyPassword := ""

	_, err := getTLSConfig(certPath, keyPath, keyPassword)
	if err == nil {
		t.Fatal("No error when trying to import missing certificate")
	}
}

func TestMissingKey(t *testing.T) {
	certPath := "test-crypt/example_cert.pem"
	keyPath := "doesntexist.pem"
	keyPassword := ""

	_, err := getTLSConfig(certPath, keyPath, keyPassword)
	if err == nil {
		t.Fatal("No error when trying to import missing key")
	}
}

func TestBadPassword(t *testing.T) {
	certPath := "test-crypt/example_cert.pem"
	keyPath := "test-crypt/example_key_encrypted.pem"
	keyPassword := "examplekeypassword11111"

	_, err := getTLSConfig(certPath, keyPath, keyPassword)
	if err == nil {
		t.Fatal("No error when trying to import encrypted key with bad password")
	}
}
