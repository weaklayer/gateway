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
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/weaklayer/gateway/server/events"
)

// NewFilesystemOutput creates a FilesystemOutput instance
func NewFilesystemOutput(directory string) (FilesystemOutput, error) {

	// Attempt to create the directory if it does not exist.
	err := createDirectory(directory)
	if err != nil {
		return FilesystemOutput{}, fmt.Errorf("Failed to create directory %s for filesystem output: %w", directory, err)
	}

	filesystemOutput := FilesystemOutput{
		directory:             directory,
		metaFiles:             make(map[uuid.UUID]metaFile),
		metaFileCreationMutex: &sync.Mutex{},
	}

	return filesystemOutput, nil
}

// FilesystemOutput is an event output that writes events to the filesystem
type FilesystemOutput struct {
	directory             string
	metaFiles             map[uuid.UUID]metaFile
	metaFileCreationMutex *sync.Mutex
}

// Close closes are underlying file descriptors for the FilesystemOutput
// Close should only be called once after Consume is guaranteed not to be called again
func (filesystemOutput FilesystemOutput) Close() {

	metaFiles := filesystemOutput.metaFiles
	filesystemOutput.metaFiles = make(map[uuid.UUID]metaFile)

	for _, metaFile := range metaFiles {
		metaFile.Close()
	}
}

// Consume takes the events and writes them to a channel for processing
func (filesystemOutput FilesystemOutput) Consume(events []events.Event) error {

	// All events in a single call will have the same group and sensor
	group := events[0].GetGroup()

	metaFile, err := filesystemOutput.getGroupMetaFile(group)
	if err != nil {
		return fmt.Errorf("Failed to write event to filesystem: %w", err)
	}

	return metaFile.Consume(events)
}

func (filesystemOutput FilesystemOutput) getGroupMetaFile(group uuid.UUID) (metaFile, error) {
	var metaFileInstance metaFile
	var ok bool
	if metaFileInstance, ok = filesystemOutput.metaFiles[group]; !ok {
		return filesystemOutput.createAndStoreGroupMetaFile(group)
	}

	return metaFileInstance, nil
}

func (filesystemOutput FilesystemOutput) createAndStoreGroupMetaFile(group uuid.UUID) (metaFile, error) {
	filesystemOutput.metaFileCreationMutex.Lock()
	defer filesystemOutput.metaFileCreationMutex.Unlock()

	var metaFileInstance metaFile
	var ok bool
	if metaFileInstance, ok = filesystemOutput.metaFiles[group]; !ok {

		metaFileDirectoryPath := filepath.Join(filesystemOutput.directory, group.String())
		err := createDirectory(metaFileDirectoryPath)
		if err != nil {
			return metaFileInstance, fmt.Errorf("Failed to create directory %s for filesystem output: %w", metaFileDirectoryPath, err)
		}

		metaFileInstance, err = newMetaFile(metaFileDirectoryPath)
		if err != nil {
			return metaFileInstance, fmt.Errorf("Failed to create file for writing: %w", err)
		}

		filesystemOutput.metaFiles[group] = metaFileInstance
	}

	return metaFileInstance, nil
}

func createDirectory(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
	}
	return err
}
