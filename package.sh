# !/bin/sh

go test ./... > /dev/null

rm -rf dist

mkdir -p dist

cp LICENSE dist/
cp example_config.yaml dist/

GOOS=linux GOARCH=amd64 go build -o dist/weaklayer-gateway-linux-amd64
GOOS=windows GOARCH=amd64 go build -o dist/weaklayer-gateway-windows-amd64.exe
GOOS=darwin GOARCH=amd64 go build -o dist/weaklayer-gateway-macOS-amd64

# This uses a locally modified version of https://github.com/google/go-licenses
# It is still expected to error saying Weaklayer Gateway is under the AGPL
../../go-licenses/go-licenses save "./" --save_path="./dist/notices"

# go-license makes everything read only
find ./dist/notices -type d -exec chmod 755 {} \;
find ./dist/notices -type f -exec chmod 644 {} \;

# Commands for cleaning out HCL
# Only need a pointer to the source for MPL 2.0 - not a copy of the source
find ./dist/notices/github.com/hashicorp/hcl/* | grep -v LICENSE | xargs rm -rf
rm -rf ./dist/notices/github.com/hashicorp/hcl/.*

echo "This directory contains reproductions of licenses/notices for the open source components used in the Weaklayer Gateway." > ./dist/notices/README.txt
echo "The directory structure tells you where to find the source of the package." >> ./dist/notices/README.txt
echo "For example, the source for 'github.com/hashicorp/hcl' can be found at 'https://github.com/hashicorp/hcl'." >> ./dist/notices/README.txt
echo "The LICENSE file in this directory is for Golang itself." >> ./dist/notices/README.txt
echo "The Weaklayer Gateway license (GNU AGPL) is present one directory up." >> ./dist/notices/README.txt
