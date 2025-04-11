#!/bin/sh
GOOS=windows GOARCH=arm64 go build -o ./gowin-arm64.exe -ldflags "-H windowsgui"