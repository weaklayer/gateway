# Weaklayer Gateway

Welcome to the Weaklayer Gateway repository!
Weaklayer is a software system for Browser Detection and Response - like Endpoint Detection and Response (EDR) but in the browser.
The Weaklayer Gateway (a server application) receives security data from Weaklayer Sensors.

Weaklayer Gateway is intended to be simple and efficient middleware for the following:
- Enforcing sensor identity and authentication
- Parsing incoming data
- Making data available to downstream systems

Note: Sensor data is not sanitized / pruned. That is, the gateway will not drop or modify data presented to it as long as it meets a couple simple requirements. For example, the gateway does not modify strings to prevent cross-site scripting or SQL injection.

## Usage

Detailed usage instructions can be found in the [Docs section of the Weaklayer Website](https://weaklayer.com/docs/). This includes instructions for execution and configuration. Things that follow in this README are for Weaklayer Gateway development.

## Building From Source

You may want to build from source to accommodate a platform not covered in the binary release or to incorporate your own modifications.

This requires you have golang (1.14+) installed.
```
git clone https://github.com/weaklayer/gateway.git
cd gateway
go test ./...
go build -o weaklayer-gateway
```

This produces an executable called `weaklayer-gateway` for the platform you are currently on.
