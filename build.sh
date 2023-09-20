#!/bin/sh

if [ -z "$NO_BUILD_WEB" ]; then
  (
    cd web
    npm install
    npm run build
  )
fi
rm -rf cmd/server/build/
cp -R web/build/ cmd/server/build/

GOARCH=amd64 GOOS=linux go build -o bin ./...
