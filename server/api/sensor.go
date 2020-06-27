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
	"net/http"
)

var lincensePaths = map[string]struct{}{"": {}, "/": {}, "/index": {}, "/index.html": {}, "/index.txt": {}, "/license": {}, "/license.html": {}, "/license.txt": {}}

// SensorAPI is the root HTTP Handler for the sensor API
type SensorAPI struct {
	InstallHandler InstallAPI
	EventsHandler  EventsAPI
}

func (sensorAPI SensorAPI) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {

	if request.Method == http.MethodPost {

		// This api only accepts json
		if request.Header.Get("Content-type") != "application/json" {
			responseWriter.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		// TODO: Other generic validations. E.g. request size

		switch request.URL.Path {
		case "/events":
			sensorAPI.EventsHandler.Handle(responseWriter, request)
		case "/install":
			sensorAPI.InstallHandler.Handle(responseWriter, request)
		default:
			responseWriter.WriteHeader(http.StatusNotFound)
		}

	} else if request.Method == http.MethodGet {
		if _, ok := lincensePaths[request.URL.Path]; ok {
			displayLicense(responseWriter, request)
		} else {
			responseWriter.WriteHeader(http.StatusNotFound)
		}
	} else {
		responseWriter.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func displayLicense(responseWriter http.ResponseWriter, request *http.Request) {
	licenseInfo := []byte(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Weaklayer Gateway</title>
</head>
<body>
<p>
This is Weaklayer Gateway.
</p>

<p>
Weaklayer Gateway is free software. It is available under the terms of the <a href="https://www.gnu.org/licenses/agpl.html">GNU Affero General Public License</a> (GNU AGPL). Please see the program source for the exact GNU AGPL version.
</p>

<p>
The Weaklayer Gateway source is available at
<a href="https://github.com/weaklayer/gateway">https://github.com/weaklayer/gateway</a>
</p>

<p>
The Weaklayer Sensor source is available at
<a href="https://github.com/weaklayer/sensor">https://github.com/weaklayer/sensor</a>
</p>

<p>
For more information, please see
<a href="https://weaklayer.com">https://weaklayer.com</a>
</p>

</body>
</html>
`)
	responseWriter.Header().Set("Content-type", "text/html")
	responseWriter.Write(licenseInfo)
}
