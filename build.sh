#!/bin/sh

(cd web; npm install; npm run build;)

rm -rf cmd/server/build/
cp -R web/build/ cmd/server/build/

GOARCH=amd64 GOOS=linux go build -o bin ./...
