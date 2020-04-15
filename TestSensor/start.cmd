@echo off
echo building client
go build -ldflags="-s -w" -o test-mqtt.exe main.go
echo starting
test-mqtt.exe -b tcp://127.0.0.1:1883 -c wohnzimmer -t data/temperatur/wohnzimmer