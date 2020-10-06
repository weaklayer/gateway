# Weaklayer Gateway (Reference Implementation)

Welcome to the Weaklayer Gateway (Reference Implementation) repository.
Weaklayer is a software system for Browser Detection and Response - like Endpoint Detection and Response (EDR) but in the browser.
The Weaklayer Gateway (a server application) receives security data from Weaklayer Sensors (browser extensions).

These are the Weaklayer Gateway Reference Implementation goals:
- Be suitable for use in production
- Be open source
- Enforce sensor identity and authentication
- Make data available to downstream systems
- Enable the use of all Weaklayer Sensor features
- Be the primary Weaklayer Gateway implementation used for Weaklayer Sensor development

These are not goals of the Weaklayer Gateway Reference Implementation:
- Graphical user interface
- Advanced administrative capabilities
- Detection analytics
- Direct integration with downstream systems

These things better belong in the Weaklayer Gateway Enterprise Edition.
Please see the [Weaklayer Website](https://weaklayer.com/contact/) for more details on the enterprise edition.

The idea is that the reference gateway implementation is simple software that gives you everything you need for Weaklayer to meaningfully contribute to your security stack.

Note that there is only one edition of the [Weaklayer Sensor](https://github.com/weaklayer/sensor), and it is open source.

## Usage

Usage instructions can be found in the [Docs section of the Weaklayer Website](https://weaklayer.com/docs/).
This includes instructions for execution and configuration. 
Things that follow in this README are for Weaklayer Gateway development.

Note: Sensor data is not sanitized / pruned. That is, the gateway will not drop or modify data presented to it as long as it meets a couple simple requirements. For example, the gateway does not modify strings to prevent cross-site scripting or SQL injection.

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
