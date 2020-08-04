# !/bin/sh

set -e

go test ./... > /dev/null

rm -rf dist
rm -rf .licenses

mkdir -p dist

licensed cache
licensed status
licensed notices

cp .licenses/NOTICE dist/
cp LICENSE dist/
cp example_config.yaml dist/
cp example_config.json dist/

GOOS=linux GOARCH=amd64 go build -o dist/weaklayer-gateway-linux-amd64
GOOS=windows GOARCH=amd64 go build -o dist/weaklayer-gateway-windows-amd64.exe
GOOS=darwin GOARCH=amd64 go build -o dist/weaklayer-gateway-macOS-amd64

cd dist
zip -r weaklayer-gateway-binary-release.zip ./
