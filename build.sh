#!/bin/bash

mkdir -p bin
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/jpcbf ./app
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/jpcbf.exe ./app
ERR=$?

if [ $ERR -eq 0 ]; then
  echo "Build complete"
else
  echo "Build error"
fi
