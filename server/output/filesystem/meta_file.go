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

package filesystem

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/weaklayer/gateway/server/events"
)

func newMetaFile(groupDirectory string) (metaFile, error) {
	newFile, err := newFile(groupDirectory)
	if err != nil {
		return metaFile{}, fmt.Errorf("Failed to create first file in directory %s: %w", groupDirectory, err)
	}

	eventData := make(chan []byte, 10000)

	go metaProcess(groupDirectory, newFile, eventData)

	return metaFile{
		groupDirectory: groupDirectory,
		eventData:      eventData,
	}, nil
}

// metaFile handles the writing to and rotation of files
type metaFile struct {
	groupDirectory string
	eventData      chan<- []byte
}

// Close should only be called once after Consume is guaranteed not to be called again
func (metaFile metaFile) Close() {
	close(metaFile.eventData)
}

func (metaFile metaFile) Consume(events []events.Event) error {

	var encounteredError = false

	for _, event := range events {
		serializedBytes, err := json.Marshal(event)
		if err != nil {
			encounteredError = true
			log.Info().Err(err).Msg("Failed to serialize event. Discarding Event")
			continue
		}

		select {
		case metaFile.eventData <- serializedBytes:
		default:
			encounteredError = true
			log.Info().Msgf("Event queue for directory %s full. Discarding Event", metaFile.groupDirectory)
			continue
		}
	}

	if encounteredError {
		return fmt.Errorf("Encountered errors serializing events for filesystem")
	}

	return nil
}

func metaProcess(groupDirectory string, initialFile file, contentChannel <-chan []byte) {
	writingFile := initialFile

	rotateFile := func() error {
		newFile, err := newFile(groupDirectory)
		if err != nil {
			return err
		}

		oldFile := writingFile
		writingFile = newFile
		oldFile.Close()

		return nil
	}

	for eventContent := range contentChannel {
		contentWritten := writingFile.Write(eventContent)
		if contentWritten {
			continue
		}

		err := rotateFile()
		if err != nil {
			log.Info().Err(err).Msg("File rotation failed")
			continue
		}

		contentWritten = writingFile.Write(eventContent)
		if !contentWritten {
			log.Info().Msg("Writing to file failed after file rotation. Discarding event")
		}
	}

	writingFile.Close()
}
