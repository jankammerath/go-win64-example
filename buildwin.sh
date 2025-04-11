#!/bin/sh
rsrc -arch arm64 -manifest app.manifest -o app.syso
GOOS=windows GOARCH=arm64 go build -o ./gowin-arm64.exe -ldflags "-H windowsgui"
rm app.syso