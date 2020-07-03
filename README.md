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

## Running

Follow these instructions to get the gateway up and running:
1. Download the latest Weaklayer Gateway binaries from the releases page
2. Extract the files
3. Identify the Weaklayer Gateway binary that corresponds to your OS and architecture (e.g. Linux amd64)
4. Run the gateway's help command `./weaklayer-gateway help`

Executing the server command like `weaklayer-gateway server --config example_config.yaml` will start the Weaklayer Gateway with and example configuration. You should create your own config file by replacing the example values.

The other commands, specifically `key` and `secret` help you generate install keys and token signing secrets. 

__N.B. You should not use the example config content in production. You should generate your own install keys and token secrets.__

## Configuration

You need to generate your own install keys and token secrets.

Note: the output of these commands is in JSON to facilitate easier integration into secret rotation procedures. The fields map directly to the default config YAML though.

Use the `weaklayer-gateway key` command to generate an install key. Note that you can generate a new key for an existing group with the `--group` flag. The output will look like this.
```
{
  "key": {
    "group": "275b3fb4-1372-4aed-91a6-a8754dc45603",
    "secret": "HjnKN/a07KavBPSZwsjG0jiBKA5w7vjO96FL4TwfQEs/aeTfwPy6KMTVNK2vSrG40YY11yWZ0WK2waB6HoSJXg==",
    "checksum": "Q+7YikSi1unGfGTYay6PExVEBaLe+1MvPilPekPNrLg="
  },
  "verifier": {
    "group": "275b3fb4-1372-4aed-91a6-a8754dc45603",
    "salt": "m46oww==",
    "hash": "pKTsSxzgnsldCzWw635JBp5DkdpeQO1NCinxlJpIP68=",
    "checksum": "YYB8aa1qDM7W1Tuqnxp3uwAM8CrkhW/jMQhqBVCYfMY="
  }
}
```
The `verifier` object goes into your gateway config and the `key` object goes into your sensor config. This key allows sensors to install into a given group. You can generally give all sensors in a given group the same install key. The gateway will give each sensor its own unique identity upon successful installation.

Use the `weaklayer-gateway secret` to generate a token signing secret. The output will look like this.
```
{
  "secret": "6TZGioH/hfiFmWxM26O0beqGnjeIZ7BXT9nlptwS/rW0zBjI76SZyMnLR/bp0SuDpa7Cgx8VDGcIeGzenBXzXA=="
}
```
This is simply a big random value that the gateway uses to sign and verify auth tokens. Sensors present these tokens to the gateway on each request to prove their identity.

## Building From Source

This requires you have golang (1.14+) installed.
```
git clone https://github.com/weaklayer/gateway.git
cd gateway
go build -o weaklayer-gateway
```

Run unit tests with this commnad.
```
go test ./...
```
