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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

func useTLS(certificatePath string, keyPath string) bool {
	return certificatePath != "" && keyPath != ""
}

func getTLSConfig(certificatePath string, keyPath string, keyPassword string) (*tls.Config, error) {

	certificateFileBytes, err := ioutil.ReadFile(certificatePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read certificate file contents: %w", err)
	}

	keyPemBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read private key file contents: %w", err)
	}

	keyPem, _ := pem.Decode(keyPemBytes)
	if keyPem == nil {
		return nil, fmt.Errorf("Private key file is not in PEM format")
	}

	// Decrypt private key if it is encrypted
	if x509.IsEncryptedPEMBlock(keyPem) {
		decryptedKeyBytes, err := x509.DecryptPEMBlock(keyPem, []byte(keyPassword))
		if err != nil {
			return nil, fmt.Errorf("Failed to decrypt private key: %w", err)
		}
		decryptedKeyPem := &pem.Block{
			Type:  keyPem.Type,
			Bytes: decryptedKeyBytes,
		}

		// override the key pem bytes from the file
		keyPemBytes = pem.EncodeToMemory(decryptedKeyPem)
		if keyPemBytes == nil {
			return nil, fmt.Errorf("Failed to convert decrypted private key to PEM format")
		}
	}

	certificate, err := tls.X509KeyPair(certificateFileBytes, keyPemBytes)

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{certificate},
	}

	return tlsConfig, nil
}
