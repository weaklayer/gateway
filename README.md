# Weaklayer Gateway

Welcome to the Weaklayer Gateway repository!
Weaklayer is a software system for Browser Detection and Response - like Endpoint Detection and Reponse (EDR) but in the browser.
The Weaklayer Gateway (a server application) receives security data from Weaklayer Sensors.

Weaklayer Gateway is intended to be simple and efficent middleware for integrating Weaklayer data into your security system. 
This means it does the following:
- Enforces sensor identity and authentication
- Verifies incoming data against schemas
- Integrate into downstream destinations for Wekalayer data

Note: While sensor data is validated against schemas, it is not sanitized. That is, the gateway will drop invalid data but it will not modify data. For example, the gateway does not modify strings to prevent cross-site scripting or SQL injection.

## Usage

Detailed usage instructions can be found in the [Docs section of the Weaklayer Website](https://weaklayer.com/docs/). This includes instructions for execution and configuration. Things that follow in this README are for Weaklayer Gateway development.

## Building From Source

This requires you have golang (1.14+) installed.
```
git clone https://github.com/weaklayer/gateway.git
cd gateway
go test ./...
go build -o weaklayer-gateway
```

This produces an executable called `weaklayer-gateway` for the platform you are currently on.
