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
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

func newFile(groupDirectory string, maxFileSize int) (file, error) {

	filename := strconv.FormatInt(time.Now().UnixNano()/1000, 10) + ".json"

	// files being written to are 'dot' files
	inProgressPath := filepath.Join(groupDirectory, "."+filename)
	finalPath := filepath.Join(groupDirectory, filename)
	fileInstance, err := os.OpenFile(inProgressPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return file{}, fmt.Errorf("Failed to open file %s: %w", inProgressPath, err)
	}
	_, err = writeToFile(fileInstance, []byte("[\n"))
	if err != nil {
		return file{}, fmt.Errorf("Failed to initialize %s with JSON array opening: %w", fileInstance.Name(), err)
	}

	content := make(chan []byte)
	done := make(chan struct{})
	fileOutput := file{
		content:     content,
		doneChannel: done,
		doneFlag:    false,
		closeGuard:  &sync.Once{},
	}

	go process(fileInstance, finalPath, maxFileSize, content, done)

	return fileOutput, nil
}

type file struct {
	content     chan<- []byte
	doneChannel <-chan struct{}
	doneFlag    bool
	closeGuard  *sync.Once
}

// Close closes the content channel to tell the file writer goroutine that no more data is coming
// It is not safe to call write and close concurrently
// Calling close will cause write to return false in the future
func (file file) Close() {
	// we surround this in a sync.Once since there are a couple palces Close can be called from
	// this makes it okay to call multiple times
	file.closeGuard.Do(func() {
		file.doneFlag = true
		close(file.content)
	})
}

func writeToFile(file *os.File, content []byte) (int, error) {
	n, err := file.Write(content)

	if err != nil {
		return n, err
	}
	if n != len(content) {
		return n, fmt.Errorf("Wrote only %d out of %d bytes to %s", n, len(content), file.Name())
	}

	return n, nil
}

// Write queues data from writing to the file
// It is not safe to call write/close multiple times concurrently
// returns true if the data will be written
// returns false if the file is closed and data will not be written
func (file file) Write(data []byte) bool {

	if file.doneFlag {
		return false
	}

	// early exit if the file is done
	// prevents trying to read from file.content if it is closed
	select {
	case <-file.doneChannel:
		file.Close()
		return false
	default:
	}

	select {
	case <-file.doneChannel:
		file.Close()
		return false
	case file.content <- data:
		return true
	}
}

func process(file *os.File, finalPath string, maxFileSize int, content <-chan []byte, doneChannel chan<- struct{}) {
	// Closure to indicate we are done
	done := false
	sayDone := func() {
		if !done {
			done = true
			close(doneChannel)
		}
	}

	var totalBytesWritten int = 0
	write := func(content []byte) error {
		n, err := writeToFile(file, content)
		totalBytesWritten = totalBytesWritten + n
		return err
	}

	closeFile := func() {
		sayDone()

		err := write([]byte("\n]"))
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to write JSON array closure to %s", file.Name())
		}

		err = file.Close()
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to properly close file %v", file.Name())
		}

		err = os.Rename(file.Name(), finalPath)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to rename %s to %s", file.Name(), finalPath)
		}
	}

	defer closeFile()

	isFirstEvent := true
	for eventContent := range content {
		// Do one event per line. Append a comma and newline to previous event.
		if !isFirstEvent {
			err := write([]byte(",\n"))
			if err != nil {
				log.Warn().Err(err).Msgf("Error in filesystem output writing to %s", file.Name())
				return
			}
		}
		isFirstEvent = false

		// Write the event
		err := write(eventContent)
		if err != nil {
			log.Warn().Err(err).Msgf("Error in filesystem output writing to %s", file.Name())
			return
		}

		if totalBytesWritten >= maxFileSize {
			sayDone()
		}
	}
}
